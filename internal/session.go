package minidm

import (
	"os"
	"os/exec"
	"syscall"
)

// Hardcoded for now
const sessionCmd = "/usr/bin/start-hyprland"

func RunSession(user string, pamEnv map[string]string) error {
	u, err := LookupUser(user)
	if err != nil {
		return err
	}

	env := []string{
		"HOME=" + u.HomeDir,
		"USER=" + user,
		"LOGNAME=" + user,
		"SHELL=" + u.Shell,
		"PATH=/usr/local/bin:/usr/bin:/bin",
		"XDG_SESSION_TYPE=wayland",
		"XDG_SESSION_CLASS=user",
	}
	for k, v := range pamEnv {
		env = append(env, k+"="+v)
	}

	cmd := exec.Command(sessionCmd)
	cmd.Env = env
	cmd.Dir = u.HomeDir
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{Uid: u.Uid, Gid: u.Gid, Groups: u.Groups},
	}
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	return cmd.Run()
}
