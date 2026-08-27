package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	minidm "github.com/r4ppz/minidm/internal"
	"golang.org/x/term"
)

func main() {
	if os.Getuid() != 0 {
		fmt.Fprintln(os.Stderr, "Error: minidm must be run as root")
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)
	var lastErr string

	for {
		clearScreen()

		if lastErr != "" {
			fmt.Printf("Error: %s\n\n", lastErr)
			lastErr = ""
		}

		sessions := loadSessions()
		selectedSession, err := selectSession(reader, sessions)
		if err != nil {
			lastErr = err.Error()
			continue
		}

		username, password, err := captureCredentials(reader)
		if err != nil {
			lastErr = err.Error()
			continue
		}

		if err := minidm.Login(username, password, selectedSession); err != nil {
			lastErr = fmt.Sprintf("Authentication failed: %v", err)
		}
	}
}

func clearScreen() {
	fmt.Print("\033[2J\033[3J\033[H")
}

func loadSessions() []minidm.Session {
	sessions, err := minidm.DiscoverSessions()
	if err == nil && len(sessions) > 0 {
		return sessions
	}

	return []minidm.Session{
		{
			ID:          "shell",
			Name:        "Default Shell",
			Exec:        "/bin/bash",
			DesktopName: "Shell",
			Type:        minidm.SessionWayland,
		},
	}
}

func selectSession(r *bufio.Reader, sessions []minidm.Session) (minidm.Session, error) {
	fmt.Println("Welcome!")
	fmt.Println("\nAvailable Sessions:")
	for i, s := range sessions {
		fmt.Printf("[%d] %s (%s)\n", i+1, s.Name, s.Type)
	}

	if len(sessions) == 1 {
		return sessions[0], nil
	}

	fmt.Printf("\nSelect Session [default: 1]: ")
	input, err := readLine(r)
	if err != nil {
		return minidm.Session{}, fmt.Errorf("failed to read session choice: %w", err)
	}

	if input == "" {
		return sessions[0], nil
	}

	idx, err := strconv.Atoi(input)
	if err != nil || idx < 1 || idx > len(sessions) {
		return minidm.Session{}, fmt.Errorf("invalid session choice '%s'", input)
	}

	return sessions[idx-1], nil
}

func captureCredentials(r *bufio.Reader) (string, string, error) {
	fmt.Print("\nUsername: ")
	username, err := readLine(r)
	if err != nil {
		return "", "", fmt.Errorf("failed to read username: %w", err)
	}
	if username == "" {
		return "", "", fmt.Errorf("username cannot be empty")
	}

	fmt.Print("Password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", "", fmt.Errorf("failed to read password: %w", err)
	}

	password := string(bytePassword)
	if password == "" {
		return "", "", fmt.Errorf("password cannot be empty")
	}

	return username, password, nil
}

func readLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}
