package minidm

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"

	"golang.org/x/term"
)

type UserInfo struct {
	Uid     uint32
	Gid     uint32
	Groups  []uint32
	HomeDir string
	Shell   string
}

func LookupUser(username string) (*UserInfo, error) {
	usr, err := user.Lookup(username)
	if err != nil {
		return nil, err
	}

	uid, err := strconv.ParseUint(usr.Uid, 10, 32)
	if err != nil {
		return nil, err
	}
	gid, err := strconv.ParseUint(usr.Gid, 10, 32)
	if err != nil {
		return nil, err
	}

	info := &UserInfo{
		Uid:     uint32(uid),
		Gid:     uint32(gid),
		HomeDir: usr.HomeDir,
		Shell:   shellFor(username),
	}

	info.Groups = append(info.Groups, uint32(gid))
	gids, err := usr.GroupIds()
	if err == nil {
		for _, g := range gids {
			if v, err := strconv.ParseUint(g, 10, 32); err == nil {
				info.Groups = append(info.Groups, uint32(v))
			}
		}
	}
	return info, nil
}

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

func shellFor(username string) string {
	fallback := "/bin/sh"

	file, err := os.Open("/etc/passwd")
	if err != nil {
		return fallback
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), ":")
		if len(parts) > 6 && parts[0] == username {
			return parts[6]
		}
	}

	if err := scanner.Err(); err != nil {
		return fallback
	}

	return fallback
}
