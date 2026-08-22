package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"syscall"

	auth "github.com/r4ppz/minidm/internal"
	"golang.org/x/sys/unix"
)

func main() {
	if os.Getuid() != 0 {
		fmt.Fprintln(os.Stderr, "Must be run as root")
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("\033[2J\033[3J\033[H") // clear term
		fmt.Println("Minimal Display Manager")

		fmt.Print("Username: ")
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

		if err := auth.Login(username, password); err != nil {
			fmt.Fprintf(os.Stderr, "login failed: %v\n", err)
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
