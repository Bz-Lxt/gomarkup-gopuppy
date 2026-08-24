package domain

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrConflict          = errors.New("conflict")
	ErrValidation        = errors.New("validation")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrInvalidCredential = errors.New("invalid credential")
	ErrInviteExpired     = errors.New("invite expired")
	ErrInviteUsed        = errors.New("invite used")
	ErrAlreadyMember     = errors.New("already a member")
	ErrTransient         = errors.New("transient")
	ErrPermanent         = errors.New("permanent")
	ErrUnsupportedMedia  = errors.New("unsupported media")
	ErrTooLarge          = errors.New("payload too large")
	ErrPathTraversal     = errors.New("path traversal")
)
