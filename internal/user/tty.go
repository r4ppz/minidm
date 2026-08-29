package user

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

func CurrentTTY() (ttyPath string, isTTY bool, err error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return "", false, nil
	}
	tty, err := os.Readlink(fmt.Sprintf("/proc/self/fd/%d", fd))
	if err != nil {
		return "", false, err
	}
	return tty, true, nil
}
