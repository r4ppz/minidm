package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	internal "github.com/r4ppz/mindm/internal"
	"golang.org/x/term"
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
		username := readLine(reader)
		if username == "" {
			continue
		}

		fmt.Print("Password: ")
		pw, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Fprintln(os.Stderr, "read password:", err)
			continue
		}
		fmt.Println()

		if err := internal.Login(username, pw); err != nil {
			fmt.Fprintf(os.Stderr, "login failed: %v\n", err)
			continue
		}
	}
}

func readLine(r *bufio.Reader) string {
	line, _ := r.ReadString('\n')
	return strings.TrimSpace(line)
}
