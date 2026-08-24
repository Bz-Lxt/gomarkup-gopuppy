package httputil

import (
	"encoding/json"
	"errors"
	"net/http"

	"gopuppy/internal/domain"
)

type Envelope struct {
	Code      string      `json:"code"`
	Message   string      `json:"message"`
	RequestID string      `json:"request_id"`
	Data      interface{} `json:"data,omitempty"`
}

func JSON(w http.ResponseWriter, status int, requestID, code, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{
		Code:      code,
		Message:   message,
		RequestID: requestID,
		Data:      data,
	})
}

func OK(w http.ResponseWriter, requestID string, data interface{}) {
	JSON(w, http.StatusOK, requestID, "OK", "ok", data)
}

func Created(w http.ResponseWriter, requestID string, data interface{}) {
	JSON(w, http.StatusCreated, requestID, "CREATED", "created", data)
}

func Error(w http.ResponseWriter, requestID string, err error) {
	status, code, msg := MapError(err)
	JSON(w, status, requestID, code, msg, nil)
}

func MapError(err error) (int, string, string) {
	switch {
	case errors.Is(err, domain.ErrUnauthorized), errors.Is(err, domain.ErrInvalidCredential):
		return http.StatusUnauthorized, "UNAUTHORIZED", err.Error()
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden, "FORBIDDEN", err.Error()
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, "NOT_FOUND", "resource not found"
	case errors.Is(err, domain.ErrConflict), errors.Is(err, domain.ErrAlreadyMember):
		return http.StatusConflict, "CONFLICT", err.Error()
	case errors.Is(err, domain.ErrValidation), errors.Is(err, domain.ErrUnsupportedMedia),
		errors.Is(err, domain.ErrPathTraversal), errors.Is(err, domain.ErrInviteExpired),
		errors.Is(err, domain.ErrInviteUsed):
		return http.StatusUnprocessableEntity, "VALIDATION", err.Error()
	case errors.Is(err, domain.ErrTooLarge):
		return http.StatusRequestEntityTooLarge, "TOO_LARGE", err.Error()
	default:
		return http.StatusInternalServerError, "INTERNAL", "internal error"
	}
}

func DecodeJSON(r *http.Request, dst interface{}) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return errors.Join(domain.ErrValidation, err)
	}
	return nil
}
