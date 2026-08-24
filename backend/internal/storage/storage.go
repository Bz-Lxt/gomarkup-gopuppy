package storage

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopuppy/internal/clock"
	"gopuppy/internal/config"
	"gopuppy/internal/domain"
)

const MaxFileBytes = 20 * 1024 * 1024

type Object struct {
	Key         string
	Size        int64
	ContentType string
	SHA256      string
}

type Store interface {
	Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
	Get(ctx context.Context, key string) (io.ReadCloser, string, error)
	Delete(ctx context.Context, key string) error
	SignedURL(ctx context.Context, key string, ttl time.Duration) (string, error)
	Driver() domain.StorageDriver
}

func New(cfg *config.Config) (Store, error) {
	switch domain.StorageDriver(cfg.StorageDriver) {
	case domain.DriverLocal, "":
		if err := os.MkdirAll(cfg.StorageLocalRoot, 0o755); err != nil {
			return nil, err
		}
		return &Local{Root: cfg.StorageLocalRoot, Secret: cfg.JWTSecret}, nil
	case domain.DriverOSS:
		return &Remote{
			Kind:      domain.DriverOSS,
			Endpoint:  cfg.OSSEndpoint,
			Bucket:    cfg.OSSBucket,
			AccessKey: cfg.OSSAccessKeyID,
			Secret:    cfg.OSSAccessKeySecret,
		}, nil
	case domain.DriverCOS:
		return &Remote{
			Kind:      domain.DriverCOS,
			Endpoint:  cfg.COSBucketURL,
			AccessKey: cfg.COSSecretID,
			Secret:    cfg.COSSecretKey,
		}, nil
	default:
		return nil, fmt.Errorf("unknown storage driver %s", cfg.StorageDriver)
	}
}

func BuildKey(petID, kind, sha, ext string, at time.Time) string {
	ym := at.In(clock.Beijing).Format("2006-01")
	return fmt.Sprintf("pets/%s/%s/%s/%s%s", petID, kind, ym, sha, ext)
}

func ExtForMIME(mime string) string {
	switch mime {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "application/pdf":
		return ".pdf"
	default:
		return ""
	}
}

func Sniff(head []byte) (string, error) {
	if len(head) >= 3 && head[0] == 0xFF && head[1] == 0xD8 && head[2] == 0xFF {
		return "image/jpeg", nil
	}
	if len(head) >= 8 && bytes.Equal(head[:8], []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}) {
		return "image/png", nil
	}
	if len(head) >= 12 && bytes.Equal(head[:4], []byte("RIFF")) && bytes.Equal(head[8:12], []byte("WEBP")) {
		return "image/webp", nil
	}
	if len(head) >= 4 && bytes.Equal(head[:4], []byte("%PDF")) {
		return "application/pdf", nil
	}
	return "", domain.ErrUnsupportedMedia
}

func SanitizeFilename(name string) error {
	if name == "" || strings.Contains(name, "..") || strings.Contains(name, "/") ||
		strings.Contains(name, "\\") || strings.Contains(name, "\x00") || filepath.IsAbs(name) {
		return domain.ErrPathTraversal
	}
	return nil
}

type Local struct {
	Root   string
	Secret string
}

func (l *Local) Driver() domain.StorageDriver { return domain.DriverLocal }

func (l *Local) abs(key string) (string, error) {
	if strings.Contains(key, "..") || strings.HasPrefix(key, "/") {
		return "", domain.ErrPathTraversal
	}
	return filepath.Join(l.Root, filepath.FromSlash(key)), nil
}

func (l *Local) Put(_ context.Context, key string, r io.Reader, size int64, _ string) error {
	p, err := l.abs(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()
	n, err := io.Copy(f, io.LimitReader(r, size+1))
	if err != nil {
		return err
	}
	if n > size && size > 0 {
		_ = os.Remove(p)
		return domain.ErrTooLarge
	}
	return nil
}

func (l *Local) Get(_ context.Context, key string) (io.ReadCloser, string, error) {
	p, err := l.abs(key)
	if err != nil {
		return nil, "", err
	}
	f, err := os.Open(p)
	if err != nil {
		return nil, "", err
	}
	return f, "", nil
}

func (l *Local) Delete(_ context.Context, key string) error {
	p, err := l.abs(key)
	if err != nil {
		return err
	}
	return os.Remove(p)
}

func (l *Local) SignedURL(_ context.Context, key string, ttl time.Duration) (string, error) {
	exp := clock.Now().Add(ttl).Unix()
	mac := hmac.New(sha256.New, []byte(l.Secret))
	_, _ = fmt.Fprintf(mac, "%s:%d", key, exp)
	sig := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("local://%s?exp=%d&sig=%s", key, exp, sig), nil
}

func (l *Local) VerifySig(key, sig string, exp int64) bool {
	if clock.Now().Unix() > exp {
		return false
	}
	mac := hmac.New(sha256.New, []byte(l.Secret))
	_, _ = fmt.Fprintf(mac, "%s:%d", key, exp)
	want := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(want), []byte(sig))
}

// Remote implements a minimal OSS/COS PUT/GET using HTTP. Credentials empty
// => operations fail with a clear configuration error (driver is wired, not mocked).
type Remote struct {
	Kind      domain.StorageDriver
	Endpoint  string
	Bucket    string
	AccessKey string
	Secret    string
}

func (r *Remote) Driver() domain.StorageDriver { return r.Kind }

func (r *Remote) configured() error {
	if r.AccessKey == "" || r.Secret == "" || r.Endpoint == "" {
		return fmt.Errorf("%s driver not configured: missing credentials (see README §7)", r.Kind)
	}
	return nil
}

func (r *Remote) url(key string) string {
	base := strings.TrimRight(r.Endpoint, "/")
	if r.Bucket != "" && !strings.Contains(base, r.Bucket) {
		return fmt.Sprintf("%s/%s/%s", base, r.Bucket, key)
	}
	return base + "/" + key
}

func (r *Remote) Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) error {
	if err := r.configured(); err != nil {
		return err
	}
	uploadCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Minute)
	defer cancel()
	req, err := http.NewRequestWithContext(uploadCtx, http.MethodPut, r.url(key), body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)
	req.ContentLength = size
	req.SetBasicAuth(r.AccessKey, r.Secret)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("%s put %d: %s", r.Kind, resp.StatusCode, string(b))
	}
	return nil
}

func (r *Remote) Get(ctx context.Context, key string) (io.ReadCloser, string, error) {
	if err := r.configured(); err != nil {
		return nil, "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.url(key), nil)
	if err != nil {
		return nil, "", err
	}
	req.SetBasicAuth(r.AccessKey, r.Secret)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	if resp.StatusCode >= 400 {
		_ = resp.Body.Close()
		return nil, "", fmt.Errorf("%s get %d", r.Kind, resp.StatusCode)
	}
	return resp.Body, resp.Header.Get("Content-Type"), nil
}

func (r *Remote) Delete(ctx context.Context, key string) error {
	if err := r.configured(); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, r.url(key), nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(r.AccessKey, r.Secret)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 && resp.StatusCode != 404 {
		return fmt.Errorf("%s delete %d", r.Kind, resp.StatusCode)
	}
	return nil
}

func (r *Remote) SignedURL(_ context.Context, key string, ttl time.Duration) (string, error) {
	if err := r.configured(); err != nil {
		return "", err
	}
	exp := clock.Now().Add(ttl).Unix()
	mac := hmac.New(sha256.New, []byte(r.Secret))
	_, _ = fmt.Fprintf(mac, "%s:%d", key, exp)
	sig := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("%s?exp=%d&sig=%s", r.url(key), exp, sig), nil
}

func HashReader(r io.Reader) (sum string, data []byte, err error) {
	h := sha256.New()
	buf, err := io.ReadAll(io.LimitReader(r, MaxFileBytes+1))
	if err != nil {
		return "", nil, err
	}
	if int64(len(buf)) > MaxFileBytes {
		return "", nil, domain.ErrTooLarge
	}
	_, _ = h.Write(buf)
	return hex.EncodeToString(h.Sum(nil)), buf, nil
}
