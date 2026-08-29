package minidm

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/msteinert/pam/v2"
)

type SessionType string

const (
	SessionWayland SessionType = "wayland"
	SessionX11     SessionType = "x11"
)

type Session struct {
	ID          string
	Name        string
	Exec        string
	DesktopName string
	Type        SessionType
}

func DiscoverSessions() []Session {
	var sessions []Session

	targets := []struct {
		dir         string
		sessionType SessionType
	}{
		{"/usr/share/wayland-sessions", SessionWayland},
		{"/usr/share/xsessions", SessionX11},
	}

	for _, target := range targets {
		files, err := os.ReadDir(target.dir)
		if err != nil {
			Debugf("Session directory %s: %v", target.dir, err)
			continue
		}

		for _, file := range files {
			if !strings.HasSuffix(file.Name(), ".desktop") {
				continue
			}

			fullpath := filepath.Join(target.dir, file.Name())
			sess, err := parseDesktopFile(fullpath, target.sessionType)
			if err != nil {
				Debugf("Parse %s: %v", fullpath, err)
				continue
			}
			if sess.Exec != "" {
				sessions = append(sessions, sess)
			}
		}
	}

	Infof("Discovered %d sessions", len(sessions))
	return sessions
}

func parseDesktopFile(path string, sessionType SessionType) (Session, error) {
	file, err := os.Open(path)
	if err != nil {
		return Session{}, err
	}
	defer file.Close()

	base := filepath.Base(path)
	session := Session{
		ID:   strings.TrimSuffix(base, ".desktop"),
		Type: sessionType,
	}

	inDesktopEntry := false

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			inDesktopEntry = line == "[Desktop Entry]"
			continue
		}
		if !inDesktopEntry {
			continue
		}
		if !strings.Contains(line, "=") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "Name":
			if session.Name == "" {
				session.Name = value
			}
		case "Exec":
			if session.Exec == "" {
				session.Exec = value
			}
		case "DesktopNames":
			if session.DesktopName == "" {
				session.DesktopName = value
			}
		}
	}

	if session.Name == "" {
		session.Name = session.ID
	}
	if session.DesktopName == "" {
		session.DesktopName = session.Name
	}

	return session, scanner.Err()
}

// parseDesktopExec parses a .desktop Exec value into an argv list following
// the freedesktop.org Desktop Entry Specification.
//
// Whitespace separates arguments unless quoted. Single quotes preserve their
// content literally (field codes are not expanded). Double quotes preserve
// their content except that a backslash may escape the special characters.
// Backslash outside quotes escapes the next character. Field codes such as
// %f or %U are removed, since a session launcher provides no files or URIs;
// a literal percent sign is written as %%.
func parseDesktopExec(s string) []string {
	var args []string
	var cur strings.Builder
	var inSingle, inDouble bool
	hasArg := false

	flush := func() {
		if hasArg || cur.Len() > 0 {
			args = append(args, cur.String())
			cur.Reset()
			hasArg = false
		}
	}

	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]

		switch {
		case inSingle:
			if r == '\'' {
				inSingle = false
			} else {
				cur.WriteRune(r)
			}

		case inDouble:
			switch r {
			case '"':
				inDouble = false
			case '\\':
				if i+1 < len(runes) {
					next := runes[i+1]
					switch next {
					case '"', '$', '`', '\\', '\n':
						cur.WriteRune(next)
						i++
					default:
						cur.WriteRune(r)
					}
				} else {
					cur.WriteRune(r)
				}
			case '%':
				if i+1 < len(runes) {
					next := runes[i+1]
					if next == '%' {
						cur.WriteRune('%')
						i++
					} else {
						i++
					}
				}
			default:
				cur.WriteRune(r)
			}

		default:
			switch r {
			case ' ', '\t', '\n':
				flush()
			case '\'':
				inSingle = true
				hasArg = true
			case '"':
				inDouble = true
				hasArg = true
			case '\\':
				if i+1 < len(runes) {
					cur.WriteRune(runes[i+1])
					i++
				} else {
					cur.WriteRune(r)
				}
			case '%':
				if i+1 < len(runes) {
					next := runes[i+1]
					if next == '%' {
						cur.WriteRune('%')
						i++
					} else {
						i++
					}
				}
			default:
				cur.WriteRune(r)
			}
		}
	}
	flush()
	return args
}

func RunSession(user string, session Session, tx *pam.Transaction) error {
	Infof("Starting session %s for user %s", session.Name, user)

	// Credentials are already established by Authenticate; this function owns
	// the rest of the PAM transaction lifecycle so every return path is
	// balanced even if session opening fails.
	credEstablished := true
	sessionOpened := false
	defer func() {
		if sessionOpened {
			if err := tx.CloseSession(0); err != nil {
				Errorf("PAM close session: %v", err)
			}
		}
		if credEstablished {
			if err := tx.SetCred(pam.DeleteCred); err != nil {
				Errorf("PAM delete credentials: %v", err)
			}
		}
		if err := tx.End(); err != nil {
			Errorf("PAM end: %v", err)
		}
	}()

	userInfo, err := LookupUser(user)
	if err != nil {
		Errorf("Lookup user %s: %v", user, err)
		return err
	}

	// pam_systemd uses XDG session metadata when opening the session, so it
	// must be set in the PAM environment before OpenSession runs.
	seat := os.Getenv("XDG_SEAT")
	if seat == "" {
		seat = "seat0"
	}
	metadata := map[string]string{
		"XDG_SESSION_TYPE":    string(session.Type),
		"XDG_SESSION_DESKTOP": session.ID,
		"XDG_CURRENT_DESKTOP": session.DesktopName,
		"XDG_SESSION_CLASS":   "user",
		"XDG_SEAT":            seat,
	}
	if vtnr := os.Getenv("XDG_VTNR"); vtnr != "" {
		metadata["XDG_VTNR"] = vtnr
	}
	for k, v := range metadata {
		if err := tx.PutEnv(k + "=" + v); err != nil {
			Errorf("PAM PutEnv %s: %v", k, err)
			return err
		}
	}

	if err := tx.OpenSession(0); err != nil {
		Errorf("PAM open session for %s: %v", user, err)
		return err
	}
	sessionOpened = true

	pamEnv, err := tx.GetEnvList()
	if err != nil {
		Errorf("PAM get env for %s: %v", user, err)
		return err
	}
	Debugf("PAM env for %s: %v", user, pamEnv)

	// Build the final environment as a map so PAM-supplied values (such as
	// XDG_RUNTIME_DIR from pam_systemd) cannot be duplicated by our own.
	env := map[string]string{
		"HOME":    userInfo.HomeDir,
		"USER":    user,
		"LOGNAME": user,
		"SHELL":   userInfo.Shell,
		"PATH":    "/usr/local/sbin:/usr/local/bin:/usr/bin:/bin",
	}
	for k, v := range pamEnv {
		env[k] = v
	}

	envList := make([]string, 0, len(env))
	for k, v := range env {
		envList = append(envList, k+"="+v)
	}

	argv := parseDesktopExec(session.Exec)
	if len(argv) == 0 {
		err := fmt.Errorf("empty Exec for session %s", session.ID)
		Errorf("%v", err)
		return err
	}

	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Env = envList
	cmd.Dir = userInfo.HomeDir
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
		Credential: &syscall.Credential{
			Uid:    userInfo.Uid,
			Gid:    userInfo.Gid,
			Groups: userInfo.Groups,
		},
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		Errorf("Start session %s for %s: %v", session.Name, user, err)
		return err
	}

	// Forward termination requests to the running session so that PAM cleanup
	// still runs in order when the service manager stops minidm.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM)
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-sigCh:
				if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
					Errorf("Signal session %s: %v", user, err)
				}
			case <-done:
				return
			}
		}
	}()

	Debugf("Exec: %v", argv)
	err = cmd.Wait()
	close(done)
	signal.Stop(sigCh)

	if err != nil {
		Errorf("Session %s for %s: %v", session.Name, user, err)
	}
	return err
}
