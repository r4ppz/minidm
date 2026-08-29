package auth

import (
	"errors"
	"fmt"

	"github.com/msteinert/pam/v2"
)

// Error wraps a user-facing message around an underlying error.
type Error struct {
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *Error) Unwrap() error { return e.Err }

// wrapError attaches a user-facing message to err. If err is already an
// *Error, it is returned unchanged so the original classification survives.
func wrapError(message string, err error) error {
	if err == nil {
		return nil
	}
	var e *Error
	if errors.As(err, &e) {
		return err
	}
	return &Error{Message: message, Err: err}
}

// classifyPAMError maps a PAM error code to a user-facing message.
func classifyPAMError(err error) error {
	var pamErr pam.Error
	if errors.As(err, &pamErr) {
		switch pamErr {
		case pam.ErrAuth, pam.ErrCredInsufficient:
			return &Error{Message: "Invalid username or password", Err: err}
		case pam.ErrUserUnknown:
			return &Error{Message: "User not found", Err: err}
		case pam.ErrAcctExpired, pam.ErrCredExpired, pam.ErrAuthtokExpired:
			return &Error{Message: "Account expired", Err: err}
		case pam.ErrPermDenied:
			return &Error{Message: "Access denied", Err: err}
		}
	}
	return &Error{Message: "Authentication service unavailable", Err: err}
}
