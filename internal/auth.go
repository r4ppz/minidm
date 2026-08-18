package minidm

import (
	"fmt"
	"os"

	"github.com/msteinert/pam/v2"
)

// Login runs the full PAM transaction for user and, on success,
// launches the session. It blocks until the session exits.
func Login(user string, password []byte) error {
	t, err := pam.StartFunc("login", user, func(s pam.Style, msg string) (string, error) {
		switch s {
		case pam.PromptEchoOff:
			return string(password), nil
		case pam.PromptEchoOn:
			return user, nil
		case pam.ErrorMsg:
			fmt.Fprintln(os.Stderr, msg)
			return "", nil
		case pam.TextInfo:
			fmt.Println(msg)
			return "", nil
		default:
			return "", fmt.Errorf("unknown style %v", s)
		}
	})
	if err != nil {
		return err
	}

	defer t.End()
	defer t.CloseSession(0)
	defer t.SetCred(pam.DeleteCred)

	tty, err := CurrentTTY()
	if err != nil {
		return err
	}

	if tty != "" {
		if err := t.SetItem(pam.Tty, tty); err != nil {
			return err
		}
	}

	if err := t.Authenticate(0); err != nil {
		return err
	}

	if err := t.AcctMgmt(0); err != nil {
		return err
	}

	if err := t.OpenSession(0); err != nil {
		return err
	}

	env, err := t.GetEnvList()
	if err != nil {
		return err
	}

	return RunSession(user, env)
}
