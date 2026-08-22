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
		return err
	}

	defer tx.End()
	defer tx.CloseSession(0)
	defer tx.SetCred(pam.DeleteCred)

	tty, err := CurrentTTY()
	if err != nil {
		return err
	}

	if tty != "" {
		if err := tx.SetItem(pam.Tty, tty); err != nil {
			return err
		}
	}

	if err := tx.Authenticate(0); err != nil {
		return err
	}

	if err := tx.AcctMgmt(0); err != nil {
		return err
	}

	if err := tx.OpenSession(0); err != nil {
		return err
	}

	env, err := tx.GetEnvList()
	if err != nil {
		return err
	}

	return RunSession(user, env)
}
