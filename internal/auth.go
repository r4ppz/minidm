package minidm

import (
	"fmt"
	"os"

	"github.com/msteinert/pam/v2"
)

func Login(user string, password string) error {
	handler := func(s pam.Style, msg string) (string, error) {
		switch s {
		case pam.PromptEchoOff:
			return password, nil
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
	}

	tx, err := pam.StartFunc("login", user, handler)
	if err != nil {
		return fmt.Errorf("pam start failed: %w", err)
	}
	defer tx.End()

	tty, ok, err := CurrentTTY()
	if err != nil {
		return fmt.Errorf("failed to get current tty: %w", err)
	}

	if ok && tty != "" {
		if err := tx.SetItem(pam.Tty, tty); err != nil {
			return fmt.Errorf("failed to set tty: %w", err)
		}
	}

	if err := tx.Authenticate(0); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	if err := tx.AcctMgmt(0); err != nil {
		return fmt.Errorf("account management check failed: %w", err)
	}

	if err := tx.SetCred(pam.EstablishCred); err != nil {
		return fmt.Errorf("failed to establish credentials: %w", err)
	}
	defer tx.SetCred(pam.DeleteCred)

	if err := tx.OpenSession(0); err != nil {
		return fmt.Errorf("failed to open session: %w", err)
	}
	defer tx.CloseSession(0)

	env, err := tx.GetEnvList()
	if err != nil {
		return err
	}

	return RunSession(user, env)
}
