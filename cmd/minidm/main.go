package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	auth "github.com/r4ppz/minidm/internal"
	session "github.com/r4ppz/minidm/internal"
	"golang.org/x/sys/unix"
)

func main() {
	if os.Getuid() != 0 {
		fmt.Fprintln(os.Stderr, "Must be run as root")
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)
	var lastErr string

	for {
		sessions, err := session.DiscoverSessions()
		if err != nil || len(sessions) == 0 {
			sessions = []session.Session{
				{
					ID:          "shell",
					Name:        "Default Shell",
					Exec:        "/bin/bash",
					DesktopName: "Shell",
					Type:        session.SessionWayland,
				},
			}
		}

		fmt.Print("\033[2J\033[3J\033[H") // clear term
		fmt.Printf("\nWelcome! :0\n")

		fmt.Println("\nAvailable Sessions:")
		for i, s := range sessions {
			fmt.Printf("[%d] %s (%s)\n", i+1, s.Name, s.Type)
		}

		selectedIndex := 0
		if len(sessions) > 1 {
			fmt.Printf("\nSelect Session [default: 1]: ")
			input, _ := readLine(reader)
			if input != "" {
				if idx, err := strconv.Atoi(input); err == nil && idx >= 1 && idx <= len(sessions) {
					selectedIndex = idx - 1
				} else {
					lastErr = "Invalid session choice"
					continue
				}
			}
		}
		selectedSession := sessions[selectedIndex]

		if lastErr != "" {
			fmt.Fprintln(os.Stderr, lastErr)
			lastErr = ""
		}

		fmt.Printf("\nSelected Session: %s\n", selectedSession.Name)

		fmt.Print("\nUsername: ")
		username, err := readLine(reader)
		if err != nil {
			fmt.Fprintln(os.Stderr, "read username:", err)
			return
		}
		if username == "" {
			continue
		}

		fmt.Print("Password: ")
		password, err := readPassword(reader)
		if err != nil {
			fmt.Fprintln(os.Stderr, "read password:", err)
			return
		}
		if password == "" {
			continue
		}

		fmt.Println()

		if err := auth.Login(username, password, selectedSession); err != nil {
			lastErr = err.Error()
			continue
		}
	}
}

func readLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	return strings.TrimSpace(line), err
}

func readPassword(r *bufio.Reader) (string, error) {
	fd := int(syscall.Stdin)
	old, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	if err == nil {
		noecho := *old
		noecho.Lflag &^= unix.ECHO
		if err := unix.IoctlSetTermios(fd, unix.TCSETS, &noecho); err != nil {
			return "", err
		}
		defer unix.IoctlSetTermios(fd, unix.TCSETS, old)
	}

	line, err := r.ReadBytes('\n')
	line = bytes.TrimRight(line, "\r\n")
	return string(line), err
}
