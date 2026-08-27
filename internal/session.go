package minidm

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
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

func DiscoverSessions() ([]Session, error) {
	var sessions []Session

	targets := []struct {
		dir         string
		sessionType SessionType
	}{
		{"/usr/share/wayland-sessions", SessionWayland},
		{"/usr/share/xsession", SessionX11},
	}

	for _, target := range targets {
		files, err := os.ReadDir(target.dir)
		if err != nil {
			continue
		}

		for _, file := range files {
			if !strings.HasSuffix(file.Name(), ".desktop") {
				continue
			}

			fullpath := filepath.Join(target.dir, file.Name())
			sess, err := parseDesktopFile(fullpath, target.sessionType)
			if err == nil && sess.Exec != "" {
				sessions = append(sessions, sess)
			}

		}
	}

	return sessions, nil
}

func parseDesktopFile(path string, sType SessionType) (Session, error) {
	file, err := os.Open(path)
	if err != nil {
		return Session{}, nil
	}
	defer file.Close()

	base := filepath.Base(path)
	sess := Session{
		ID:   strings.TrimSuffix(base, ".desktop"),
		Type: sType,
	}

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "Name":
			if sess.Name == "" {
				sess.Name = val
			}
		case "Exec":
			sess.Exec = val
		case "DesktopNames":
			sess.DesktopName = val
		}
	}

	if sess.Name == "" {
		sess.Name = sess.ID
	}

	if sess.DesktopName == "" {
		sess.DesktopName = sess.Name
	}

	return sess, scanner.Err()
}

func RunSession(user string, sess Session, pamEnv map[string]string) error {
	u, err := LookupUser(user)
	if err != nil {
		return err
	}

	// Base environment
	env := []string{
		"HOME=" + u.HomeDir,
		"USER=" + user,
		"LOGNAME=" + user,
		"SHELL=" + u.Shell,
		"PATH=/usr/local/sbin:/usr/local/bin:/usr/bin:/bin",

		// Dynamic XDG variables based on detected session:
		"XDG_SESSION_TYPE=" + string(sess.Type),
		"XDG_SESSION_DESKTOP=" + sess.ID,
		"XDG_CURRENT_DESKTOP=" + sess.DesktopName,
		"XDG_SESSION_CLASS=user",
	}

	for k, v := range pamEnv {
		env = append(env, k+"="+v)
	}

	cmd := exec.Command(u.Shell, "-l", "-c", "exec "+sess.Exec)
	cmd.Env = env
	cmd.Dir = u.HomeDir
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{Uid: u.Uid, Gid: u.Gid, Groups: u.Groups},
	}
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	return cmd.Run()
}
