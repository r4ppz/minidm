package mindm

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
			return string(password), nil // pam_unix asks -> give it
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

	if tty := CurrentTTY(); tty != "" {
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
	defer t.CloseSession(0) // runs after the session exits
	defer t.SetCred(pam.DeleteCred)
	env, err := t.GetEnvList() // XDG_RUNTIME_DIR, XDG_SESSION_ID from pam_systemd
	if err != nil {
		return err
	}

	return RunSession(user, env)
}
