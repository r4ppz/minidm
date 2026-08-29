package auth

import (
	"fmt"
	"strings"

	"github.com/msteinert/pam/v2"
	"github.com/r4ppz/minidm/internal/log"
	"github.com/r4ppz/minidm/internal/user"
)

// Authenticate performs PAM authentication, account management, and credential
// establishment for the given user. The returned transaction must be passed to
// RunSession to complete the PAM lifecycle (open/close session, delete
// credentials, end transaction).
func Authenticate(username, password string) (*pam.Transaction, error) {
	log.Infof("Authenticating user: %s", username)

	var pamErrMsg string

	handler := func(style pam.Style, msg string) (string, error) {
		switch style {
		case pam.PromptEchoOff:
			return password, nil
		case pam.PromptEchoOn:
			return username, nil
		case pam.ErrorMsg:
			pamErrMsg = msg
			log.Errorf("PAM error: %s", msg)
			return "", nil
		case pam.TextInfo:
			log.Debugf("PAM info: %s", msg)
			return "", nil
		default:
			return "", fmt.Errorf("unknown PAM style %v", style)
		}
	}

	tx, err := pam.StartFunc("minidm", username, handler)
	if err != nil {
		log.Errorf("PAM start: %v", err)
		return nil, newError(CodeSystemError, fmt.Errorf("pam start: %w", err))
	}

	tty, ok, err := user.CurrentTTY()
	if err != nil {
		log.Errorf("Get TTY: %v", err)
		tx.End()
		return nil, newError(CodeSystemError, fmt.Errorf("get tty: %w", err))
	}
	if ok && tty != "" {
		log.Debugf("PAM TTY: %s", tty)
		if err := tx.SetItem(pam.Tty, tty); err != nil {
			log.Errorf("Set PAM TTY: %v", err)
			tx.End()
			return nil, newError(CodeSystemError, fmt.Errorf("set tty: %w", err))
		}
	}

	if err := tx.Authenticate(0); err != nil {
		if pamErrMsg != "" {
			err = fmt.Errorf("%s: %w", pamErrMsg, err)
		}
		log.Errorf("Auth failed for %s: %v", username, err)
		tx.End()
		return nil, mapPAMError(err)
	}

	if err := tx.AcctMgmt(0); err != nil {
		log.Errorf("Acct mgmt failed for %s: %v", username, err)
		tx.End()
		return nil, newError(CodeSystemError, fmt.Errorf("account management: %w", err))
	}

	if err := tx.SetCred(pam.EstablishCred); err != nil {
		log.Errorf("Set cred failed for %s: %v", username, err)
		tx.End()
		return nil, newError(CodeSystemError, fmt.Errorf("establish credentials: %w", err))
	}

	log.Infof("Auth success: %s", username)
	return tx, nil
}

func mapPAMError(err error) *Error {
	if err == nil {
		return nil
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "authentication failure"), strings.Contains(msg, "auth failed"):
		return newError(CodeAuthFailed, err)
	case strings.Contains(msg, "user unknown"), strings.Contains(msg, "no such user"):
		return newError(CodeUserUnknown, err)
	case strings.Contains(msg, "account expired"), strings.Contains(msg, "password expired"):
		return newError(CodeAccountExpired, err)
	case strings.Contains(msg, "account locked"), strings.Contains(msg, "locked"):
		return newError(CodeAccountLocked, err)
	case strings.Contains(msg, "permission denied"), strings.Contains(msg, "access denied"):
		return newError(CodePermissionDenied, err)
	default:
		return newError(CodeSystemError, err)
	}
}
