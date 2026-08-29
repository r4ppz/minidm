package minidm

import (
	"fmt"
	"strings"

	"github.com/msteinert/pam/v2"
)

func Authenticate(user string, password string, session Session) (*pam.Transaction, error) {
	Infof("Authenticating user: %s", user)

	var pamErrMsg string

	handler := func(style pam.Style, msg string) (string, error) {
		switch style {
		case pam.PromptEchoOff:
			return password, nil
		case pam.PromptEchoOn:
			return user, nil
		case pam.ErrorMsg:
			pamErrMsg = msg
			Errorf("PAM error: %s", msg)
			return "", nil
		case pam.TextInfo:
			Debugf("PAM info: %s", msg)
			return "", nil
		default:
			return "", fmt.Errorf("unknown PAM style %v", style)
		}
	}

	tx, err := pam.StartFunc("minidm", user, handler)
	if err != nil {
		Errorf("PAM start: %v", err)
		return nil, NewAuthError(ErrSystemError, fmt.Errorf("pam start: %w", err))
	}

	tty, ok, err := CurrentTTY()
	if err != nil {
		Errorf("Get TTY: %v", err)
		tx.End()
		return nil, NewAuthError(ErrSystemError, fmt.Errorf("get tty: %w", err))
	}
	if ok && tty != "" {
		Debugf("PAM TTY: %s", tty)
		if err := tx.SetItem(pam.Tty, tty); err != nil {
			Errorf("Set PAM TTY: %v", err)
			tx.End()
			return nil, NewAuthError(ErrSystemError, fmt.Errorf("set tty: %w", err))
		}
	}

	if err := tx.Authenticate(0); err != nil {
		if pamErrMsg != "" {
			err = fmt.Errorf("%s: %w", pamErrMsg, err)
		}
		Errorf("Auth failed for %s: %v", user, err)
		tx.End()
		return nil, mapPAMError(err)
	}

	if err := tx.AcctMgmt(0); err != nil {
		Errorf("Acct mgmt failed for %s: %v", user, err)
		tx.End()
		return nil, NewAuthError(ErrSystemError, fmt.Errorf("account management: %w", err))
	}

	if err := tx.SetCred(pam.EstablishCred); err != nil {
		Errorf("Set cred failed for %s: %v", user, err)
		tx.End()
		return nil, NewAuthError(ErrSystemError, fmt.Errorf("establish credentials: %w", err))
	}

	Infof("Auth success: %s", user)
	return tx, nil
}

func mapPAMError(err error) *AuthError {
	if err == nil {
		return nil
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "authentication failure"), strings.Contains(msg, "auth failed"):
		return NewAuthError(ErrAuthFailed, err)
	case strings.Contains(msg, "user unknown"), strings.Contains(msg, "no such user"):
		return NewAuthError(ErrUserUnknown, err)
	case strings.Contains(msg, "account expired"), strings.Contains(msg, "password expired"):
		return NewAuthError(ErrAccountExpired, err)
	case strings.Contains(msg, "account locked"), strings.Contains(msg, "locked"):
		return NewAuthError(ErrAccountLocked, err)
	case strings.Contains(msg, "permission denied"), strings.Contains(msg, "access denied"):
		return NewAuthError(ErrPermissionDenied, err)
	default:
		return NewAuthError(ErrSystemError, err)
	}
}
