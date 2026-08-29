package auth

import (
	"errors"
	"fmt"

	"github.com/msteinert/pam/v2"
	"github.com/r4ppz/minidm/internal/log"
	"github.com/r4ppz/minidm/internal/user"
)

func Authenticate(username, password string) (tx *pam.Transaction, err error) {
	log.Info("authenticating user", "user", username)

	tx, err = pam.StartFunc("minidm", username, func(style pam.Style, msg string) (string, error) {
		switch style {
		case pam.PromptEchoOff:
			return password, nil
		case pam.PromptEchoOn:
			return username, nil
		case pam.ErrorMsg:
			log.Error("PAM error message", "msg", msg)
			return "", nil
		case pam.TextInfo:
			log.Debug("PAM info", "msg", msg)
			return "", nil
		default:
			return "", fmt.Errorf("unknown PAM style %v", style)
		}
	})
	if err != nil {
		return nil, wrapError("Authentication service unavailable", err)
	}
	defer func() {
		if err != nil {
			tx.End()
		}
	}()

	if tty, ok, ttyErr := user.CurrentTTY(); ttyErr != nil {
		err = wrapError("Authentication service unavailable", ttyErr)
		return
	} else if ok && tty != "" {
		log.Debug("PAM TTY", "tty", tty)
		if err = tx.SetItem(pam.Tty, tty); err != nil {
			err = wrapError("Authentication service unavailable", err)
			return
		}
	}

	if err = tx.Authenticate(0); err != nil {
		err = classifyPAMError(err)
		return
	}

	if err = tx.AcctMgmt(0); err != nil {
		err = classifyPAMError(err)
		return
	}

	if err = tx.SetCred(pam.EstablishCred); err != nil {
		err = wrapError("Authentication service unavailable", err)
		return
	}

	log.Info("auth success", "user", username)
	return tx, nil
}

func IsAuthError(err error) bool {
	var target *Error
	return errors.As(err, &target)
}
