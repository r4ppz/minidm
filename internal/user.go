package minidm

import (
	"bufio"
	"os"
	"os/user"
	"strconv"
	"strings"
)

type UserInfo struct {
	Uid     uint32
	Gid     uint32
	Groups  []uint32
	HomeDir string
	Shell   string
}

func LookupUser(username string) (*UserInfo, error) {
	u, err := user.Lookup(username)
	if err != nil {
		return nil, err
	}

	uid, err := strconv.ParseUint(u.Uid, 10, 32)
	if err != nil {
		return nil, err
	}
	gid, err := strconv.ParseUint(u.Gid, 10, 32)
	if err != nil {
		return nil, err
	}

	info := &UserInfo{
		Uid:     uint32(uid),
		Gid:     uint32(gid),
		HomeDir: u.HomeDir,
		Shell:   shellFor(username),
	}

	info.Groups = append(info.Groups, uint32(gid))
	gids, err := u.GroupIds()
	if err == nil {
		for _, g := range gids {
			if v, err := strconv.ParseUint(g, 10, 32); err == nil {
				info.Groups = append(info.Groups, uint32(v))
			}
		}
	}
	return info, nil
}

func CurrentTTY() (string, error) {
	tty, err := os.Readlink("/proc/self/fd/0")
	if err != nil {
		return "", err
	}
	if tty == "" {
		return "", nil
	}

	return tty, nil
}

func shellFor(username string) string {
	fallback := "/bin/sh"

	f, err := os.Open("/etc/passwd")
	if err != nil {
		return fallback
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	if err := s.Err(); err != nil {
		return fallback
	}

	for s.Scan() {
		parts := strings.Split(s.Text(), ":")
		if len(parts) > 6 && parts[0] == username {
			return parts[6]
		}
	}

	return fallback
}
