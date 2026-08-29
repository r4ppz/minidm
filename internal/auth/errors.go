package auth

import (
	"errors"
	"fmt"
)

type ErrorCode int

const (
	CodeAuthFailed ErrorCode = iota
	CodeUserUnknown
	CodeAccountLocked
	CodeAccountExpired
	CodePermissionDenied
	CodeSystemError
)

var errorMessages = map[ErrorCode]string{
	CodeAuthFailed:       "Invalid username or password",
	CodeUserUnknown:      "User not found",
	CodeAccountLocked:    "Account locked",
	CodeAccountExpired:   "Account expired",
	CodePermissionDenied: "Access denied",
	CodeSystemError:      "Authentication service unavailable",
}

type Error struct {
	Code    ErrorCode
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Err
}

func (e *Error) Is(target error) bool {
	var authErr *Error
	if errors.As(target, &authErr) {
		return e.Code == authErr.Code
	}
	return false
}

func newError(code ErrorCode, err error) *Error {
	msg := errorMessages[code]
	if msg == "" {
		msg = errorMessages[CodeSystemError]
	}
	return &Error{Code: code, Message: msg, Err: err}
}
