package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"syscall"

	"github.com/msteinert/pam/v2"
	"golang.org/x/term"
)

const sessionCmd = "/usr/bin/start-hyprland"

func main() {
	if os.Getuid() != 0 {
		fmt.Fprintln(os.Stderr, "mindm: must be run as root")
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("Minimal Display Manager")

		fmt.Print("Username: ")
		username := readLine(reader)
		if username == "" {
			continue
		}

		fmt.Print("Password: ")
		pw, _ := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()

		if err := login(username, pw); err != nil {
			fmt.Fprintf(os.Stderr, "login failed: %v\n", err)
			continue
		}
	}
}

func readLine(r *bufio.Reader) string {
	line, _ := r.ReadString('\n')
	return strings.TrimSpace(line)
}

func currentTTY() string {
	tty, err := os.Readlink("/proc/self/fd/0")
	if err != nil || tty == "" {
		return ""
	}
	return tty
}

func shellFor(username string) string {
	f, err := os.Open("/etc/passwd")
	if err != nil {
		return "/bin/sh"
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		parts := strings.Split(s.Text(), ":")
		if len(parts) > 6 && parts[0] == username {
			return parts[6]
		}
	}
	if err := s.Err(); err != nil {
		return "/bin/sh"
	}
	return "/bin/sh"
}

func login(user string, password []byte) error {
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

	if tty := currentTTY(); tty != "" {
		t.SetItem(pam.Tty, tty)
	}

	if err := t.Authenticate(0); err != nil {
		return err
	}
	if err := t.AcctMgmt(0); err != nil {
		return err
	}
	if err := t.SetCred(pam.EstablishCred); err != nil {
		return err
	}
	if err := t.OpenSession(0); err != nil {
		return err
	}
	defer t.CloseSession(0)  // runs after the session exits
	env, _ := t.GetEnvList() // XDG_RUNTIME_DIR, XDG_SESSION_ID from pam_systemd

	return runSession(user, env)
}

func runSession(username string, pamEnv map[string]string) error {
	u, err := user.Lookup(username)
	if err != nil {
		return err
	}

	uid, _ := strconv.ParseUint(u.Uid, 10, 32)
	gid, _ := strconv.ParseUint(u.Gid, 10, 32)

	groups := []uint32{uint32(gid)} // primary group, then supplementary ones
	gids, err := u.GroupIds()
	if err == nil {
		for _, g := range gids {
			if v, err := strconv.ParseUint(g, 10, 32); err == nil {
				groups = append(groups, uint32(v))
			}
		}
	}

	env := []string{
		"HOME=" + u.HomeDir,
		"USER=" + username,
		"LOGNAME=" + username,
		"SHELL=" + shellFor(username),
		"PATH=/usr/local/bin:/usr/bin:/bin",
	}
	for k, v := range pamEnv {
		env = append(env, k+"="+v)
	}

	cmd := exec.Command(sessionCmd)
	cmd.Env = env
	cmd.Dir = u.HomeDir
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid), Groups: groups},
	}
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	return cmd.Run() // blocks until Hyprland quits, then we loop back to the login prompt
}
