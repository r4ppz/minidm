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

// RunSession completes the PAM session lifecycle and blocks until the desktop
// exits. It owns cleanup (close session, delete credentials, end transaction)
// so every return path is balanced even on error.
func RunSession(username string, sess session.Session, tx *pam.Transaction) error {
	log.Infof("Starting session %s for user %s", sess.Name, username)

	credEstablished := true
	sessionOpened := false
	defer func() {
		if sessionOpened {
			if err := tx.CloseSession(0); err != nil {
				log.Errorf("PAM close session: %v", err)
			}
		}
		if credEstablished {
			if err := tx.SetCred(pam.DeleteCred); err != nil {
				log.Errorf("PAM delete credentials: %v", err)
			}
		}
		if err := tx.End(); err != nil {
			log.Errorf("PAM end: %v", err)
		}
	}()

	info, err := user.Lookup(username)
	if err != nil {
		log.Errorf("Lookup user %s: %v", username, err)
		return err
	}

	if err := setSessionMetadata(tx, sess); err != nil {
		return err
	}

	if err := tx.OpenSession(0); err != nil {
		log.Errorf("PAM open session for %s: %v", username, err)
		return err
	}
	sessionOpened = true

	pamEnv, err := tx.GetEnvList()
	if err != nil {
		log.Errorf("PAM get env for %s: %v", username, err)
		return err
	}
	log.Debugf("PAM env for %s: %v", username, pamEnv)

	env := buildEnv(username, info, pamEnv)

	argv := session.ParseExec(sess.Exec)
	if len(argv) == 0 {
		err := fmt.Errorf("empty Exec for session %s", sess.ID)
		log.Errorf("%v", err)
		return err
	}

	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Env = env
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
		log.Errorf("Start session %s for %s: %v", sess.Name, username, err)
		return err
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM)
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-sigCh:
				if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
					log.Errorf("Signal session %s: %v", username, err)
				}
			case <-done:
				return
			}
		}
	}()

	log.Debugf("Exec: %v", argv)
	err = cmd.Wait()
	close(done)
	signal.Stop(sigCh)

	if err != nil {
		log.Errorf("Session %s for %s: %v", sess.Name, username, err)
	}
	return err
}

// setSessionMetadata puts XDG session metadata into the PAM environment before
// OpenSession runs, so pam_systemd can see it when creating the session.
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
			log.Errorf("PAM PutEnv %s: %v", k, err)
			return err
		}
	}
	return nil
}

// buildEnv constructs the final environment. PAM-supplied values (such as
// XDG_RUNTIME_DIR from pam_systemd) take precedence over our defaults to avoid
// duplicates.
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

	envList := make([]string, 0, len(env))
	for k, v := range env {
		envList = append(envList, k+"="+v)
	}
	return envList
}
