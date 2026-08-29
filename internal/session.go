package minidm

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
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
			session.Exec = value
		case "DesktopNames":
			session.DesktopName = value
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

func RunSession(user string, session Session, tx *pam.Transaction) (*os.Process, error) {
	Infof("Starting session %s for user %s", session.Name, user)

	defer tx.End()

	userInfo, err := LookupUser(user)
	if err != nil {
		Errorf("Lookup user %s: %v", user, err)
		return nil, err
	}

	runtimeDir := "/run/user/" + strconv.FormatUint(uint64(userInfo.Uid), 10)
	if err := os.MkdirAll(runtimeDir, 0700); err != nil {
		Errorf("Create XDG_RUNTIME_DIR %s: %v", runtimeDir, err)
		return nil, err
	} else if err := os.Chown(runtimeDir, int(userInfo.Uid), int(userInfo.Gid)); err != nil {
		Errorf("Chown XDG_RUNTIME_DIR %s: %v", runtimeDir, err)
	}

	if err := tx.OpenSession(0); err != nil {
		Errorf("PAM open session for %s: %v", user, err)
		return nil, err
	}
	defer tx.SetCred(pam.DeleteCred)
	defer tx.CloseSession(0)

	pamEnv, err := tx.GetEnvList()
	if err != nil {
		Errorf("PAM get env for %s: %v", user, err)
		return nil, err
	}

	Infof("PAM env for %s: %v", user, pamEnv)

	seat := os.Getenv("XDG_SEAT")
	if seat == "" {
		seat = "seat0"
	}

	env := []string{
		"HOME=" + userInfo.HomeDir,
		"USER=" + user,
		"LOGNAME=" + user,
		"SHELL=" + userInfo.Shell,
		"PATH=/usr/local/sbin:/usr/local/bin:/usr/bin:/bin",
		"XDG_RUNTIME_DIR=" + runtimeDir,
		"XDG_SESSION_TYPE=" + string(session.Type),
		"XDG_SESSION_DESKTOP=" + session.ID,
		"XDG_CURRENT_DESKTOP=" + session.DesktopName,
		"XDG_SESSION_CLASS=user",
		"XDG_SEAT=" + seat,
	}

	if vtnr := os.Getenv("XDG_VTNR"); vtnr != "" {
		env = append(env, "XDG_VTNR="+vtnr)
	}

	for k, v := range pamEnv {
		env = append(env, k+"="+v)
	}

	cmd := exec.Command(userInfo.Shell, "-l", "-c", "exec "+session.Exec)
	cmd.Env = env
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

	Debugf("Exec: %s", session.Exec)
	err = cmd.Run()
	if err != nil {
		Errorf("Session %s for %s: %v", session.Name, user, err)
	}
	return cmd.Process, err
}
