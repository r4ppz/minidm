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
		fmt.Fprintln(os.Stderr, "minidm: must be run as root")
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Minimal Display Manager")

	for {

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
		pw, err := readPassword(reader)
		if err != nil {
			fmt.Fprintln(os.Stderr, "read password:", err)
			return
		}
		fmt.Println()

		if err := auth.Login(username, pw); err != nil {
			fmt.Fprintf(os.Stderr, "login failed: %v\n", err)
			continue
		}
	}
}

func readLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	return strings.TrimSpace(line), err
}

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
