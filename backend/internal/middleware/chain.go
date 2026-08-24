package middleware

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"gopuppy/internal/auth"
	"gopuppy/internal/domain"
	"gopuppy/internal/httputil"
)

type ctxKey string

const (
	CtxRequestID ctxKey = "request_id"
	CtxUserID    ctxKey = "user_id"
)

func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get("X-Request-ID")
			if id == "" {
				b := make([]byte, 8)
				_, _ = rand.Read(b)
				id = hex.EncodeToString(b)
			}
			ww := &hijackWriter{ResponseWriter: w}
			ctx := context.WithValue(r.Context(), CtxRequestID, id)
			ww.Header().Set("X-Request-ID", id)
			next.ServeHTTP(ww, r.WithContext(ctx))
		})
	}
}

func RequestIDFrom(ctx context.Context) string {
	if v, ok := ctx.Value(CtxRequestID).(string); ok {
		return v
	}
	return ""
}

func UserIDFrom(ctx context.Context) (uuid.UUID, bool) {
	v, ok := ctx.Value(CtxUserID).(uuid.UUID)
	return v, ok
}

func Recover(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("panic", "err", rec, "request_id", RequestIDFrom(r.Context()))
					httputil.JSON(w, http.StatusInternalServerError, RequestIDFrom(r.Context()), "INTERNAL", "internal error", nil)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func CORS(origins []string) func(http.Handler) http.Handler {
	allow := map[string]struct{}{}
	for _, o := range origins {
		allow[o] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" {
				if _, ok := allow[origin]; ok {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Vary", "Origin")
					w.Header().Set("Access-Control-Allow-Credentials", "true")
					w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
					w.Header().Set("Access-Control-Max-Age", "600")
				}
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func Auth(issuer *auth.Issuer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := bearer(r)
			if raw == "" {
				httputil.Error(w, RequestIDFrom(r.Context()), domain.ErrUnauthorized)
				return
			}
			c, err := issuer.Parse(raw)
			if err != nil || c.Kind != "access" {
				httputil.Error(w, RequestIDFrom(r.Context()), domain.ErrUnauthorized)
				return
			}
			ctx := context.WithValue(context.Background(), CtxUserID, c.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func BearerToken(r *http.Request) string {
	return bearer(r)
}

func bearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(h), "bearer ") {
		return strings.TrimSpace(h[7:])
	}
	return strings.TrimSpace(r.URL.Query().Get("token"))
}

// hijackWriter preserves http.Hijacker so gorilla/websocket can upgrade.
type hijackWriter struct {
	http.ResponseWriter
}

func (w *hijackWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return h.Hijack()
}

func (w *hijackWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
