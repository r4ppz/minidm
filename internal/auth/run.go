package auth

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/msteinert/pam/v2"
	"github.com/r4ppz/minidm/internal/log"
	"github.com/r4ppz/minidm/internal/session"
	"github.com/r4ppz/minidm/internal/user"
)

func RunSession(username string, sess session.Session, tx *pam.Transaction) (err error) {
	log.Info("starting session", "session", sess.Name, "user", username)

	defer func() {
		if err := tx.CloseSession(0); err != nil {
			log.Error("PAM close session", "err", err)
		}
		if err := tx.SetCred(pam.DeleteCred); err != nil {
			log.Error("PAM delete credentials", "err", err)
		}
		if err := tx.End(); err != nil {
			log.Error("PAM end", "err", err)
		}
	}()

	info, err := user.Lookup(username)
	if err != nil {
		return fmt.Errorf("lookup user: %w", err)
	}

	if err := setSessionMetadata(tx, sess); err != nil {
		return fmt.Errorf("set session metadata: %w", err)
	}

	if err := tx.OpenSession(0); err != nil {
		return fmt.Errorf("PAM open session: %w", err)
	}

	pamEnv, err := tx.GetEnvList()
	if err != nil {
		return fmt.Errorf("PAM get env: %w", err)
	}
	log.Debug("PAM environment", "env", pamEnv)

	argv := session.ParseExec(sess.Exec)
	if len(argv) == 0 {
		return fmt.Errorf("empty Exec for session %s", sess.ID)
	}

	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Env = buildEnv(username, info, pamEnv)
	cmd.Dir = info.HomeDir
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
		Credential: &syscall.Credential{
			Uid:    info.Uid,
			Gid:    info.Gid,
			Groups: info.Groups,
		},
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start session: %w", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM)
	done := make(chan struct{})
	go func() {
		select {
		case <-sigCh:
			if sigErr := cmd.Process.Signal(syscall.SIGTERM); sigErr != nil {
				log.Error("forward SIGTERM", "err", sigErr)
			}
		case <-done:
		}
	}()

	log.Debug("exec", "argv", argv)
	err = cmd.Wait()
	close(done)
	signal.Stop(sigCh)

	if err != nil {
		return fmt.Errorf("session exited: %w", err)
	}
	return nil
}

func setSessionMetadata(tx *pam.Transaction, sess session.Session) error {
	seat := os.Getenv("XDG_SEAT")
	if seat == "" {
		seat = "seat0"
	}
	metadata := map[string]string{
		"XDG_SESSION_TYPE":    string(sess.Type),
		"XDG_SESSION_DESKTOP": sess.ID,
		"XDG_CURRENT_DESKTOP": sess.DesktopName,
		"XDG_SESSION_CLASS":   "user",
		"XDG_SEAT":            seat,
	}
	if vtnr := os.Getenv("XDG_VTNR"); vtnr != "" {
		metadata["XDG_VTNR"] = vtnr
	}
	for k, v := range metadata {
		if err := tx.PutEnv(k + "=" + v); err != nil {
			return err
		}
	}
	return nil
}

func buildEnv(username string, info *user.Info, pamEnv map[string]string) []string {
	env := map[string]string{
		"HOME":    info.HomeDir,
		"USER":    username,
		"LOGNAME": username,
		"SHELL":   info.Shell,
		"PATH":    "/usr/local/sbin:/usr/local/bin:/usr/bin:/bin",
	}
	for k, v := range pamEnv {
		env[k] = v
	}
	out := make([]string, 0, len(env))
	for k, v := range env {
		out = append(out, k+"="+v)
	}
	return out
}
