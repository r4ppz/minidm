package minidm

import (
	"errors"
	"fmt"
)

type AuthErrorCode int

const (
	ErrAuthFailed AuthErrorCode = iota
	ErrUserUnknown
	ErrAccountLocked
	ErrAccountExpired
	ErrPermissionDenied
	ErrSystemError
)

var authErrorMessages = map[AuthErrorCode]string{
	ErrAuthFailed:       "Invalid username or password",
	ErrUserUnknown:      "User not found",
	ErrAccountLocked:    "Account locked",
	ErrAccountExpired:   "Account expired",
	ErrPermissionDenied: "Access denied",
	ErrSystemError:      "Authentication service unavailable",
}

type AuthError struct {
	Code    AuthErrorCode
	Message string
	Err     error
}

func (e *AuthError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AuthError) Unwrap() error {
	return e.Err
}

func (e *AuthError) Is(target error) bool {
	var authErr *AuthError
	if errors.As(target, &authErr) {
		return e.Code == authErr.Code
	}
	return false
}

func NewAuthError(code AuthErrorCode, err error) *AuthError {
	msg := authErrorMessages[code]
	if msg == "" {
		msg = authErrorMessages[ErrSystemError]
	}
	return &AuthError{Code: code, Message: msg, Err: err}
}
