package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"syscall"

	internal "github.com/r4ppz/mindm/internal"
	"golang.org/x/sys/unix"
)

func main() {
	if os.Getuid() != 0 {
		fmt.Fprintln(os.Stderr, "mindm: must be run as root")
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("Minimal Display Manager")

		fmt.Print("Username: ")
		username, err := readLine(reader)
		if err != nil {
			break
		}
		if username == "" {
			continue
		}

		fmt.Print("Password: ")
		pw, err := readPassword(reader)
		if err != nil {
			fmt.Fprintln(os.Stderr, "read password:", err)
			break
		}
		fmt.Println()

		if err := internal.Login(username, pw); err != nil {
			fmt.Fprintf(os.Stderr, "login failed: %v\n", err)
			continue
		}
	}
}

// readLine reads a line from r and trims surrounding whitespace. It
// returns an error on EOF so callers can stop reading.
func readLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	return strings.TrimSpace(line), err
}

// readPassword reads a line from r with echo disabled. It reads from
// the same reader as the username so the input stays in sync, even for
// piped or pasted input.
func readPassword(r *bufio.Reader) ([]byte, error) {
	fd := int(syscall.Stdin)
	old, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	if err == nil {
		noecho := *old
		noecho.Lflag &^= unix.ECHO
		if err := unix.IoctlSetTermios(fd, unix.TCSETS, &noecho); err != nil {
			return nil, err
		}
		defer unix.IoctlSetTermios(fd, unix.TCSETS, old)
	}

	line, err := r.ReadBytes('\n')
	line = bytes.TrimRight(line, "\r\n")
	return line, err
}
