package user

import (
	"bufio"
	"fmt"
	"os"
	ouser "os/user"
	"strconv"
	"strings"

	"github.com/r4ppz/minidm/internal/log"
	"golang.org/x/term"
)

type Info struct {
	Uid     uint32
	Gid     uint32
	Groups  []uint32
	HomeDir string
	Shell   string
}

func Lookup(username string) (*Info, error) {
	osUser, err := ouser.Lookup(username)
	if err != nil {
		return nil, fmt.Errorf("lookup user: %w", err)
	}

	uid, err := strconv.ParseUint(osUser.Uid, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("parse UID: %w", err)
	}
	gid, err := strconv.ParseUint(osUser.Gid, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("parse GID: %w", err)
	}

	info := &Info{
		Uid:     uint32(uid),
		Gid:     uint32(gid),
		HomeDir: osUser.HomeDir,
		Shell:   lookupShell(username),
	}

	info.Groups = append(info.Groups, uint32(gid))
	if groupIDs, err := osUser.GroupIds(); err == nil {
		for _, gidStr := range groupIDs {
			if parsed, err := strconv.ParseUint(gidStr, 10, 32); err == nil {
				info.Groups = append(info.Groups, uint32(parsed))
			}
		}
	}

	log.Debug("user lookup",
		"user", username,
		"uid", info.Uid,
		"gid", info.Gid,
		"home", info.HomeDir,
		"shell", info.Shell)
	return info, nil
}

func CurrentTTY() (path string, isTTY bool, err error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return "", false, nil
	}
	path, err = os.Readlink(fmt.Sprintf("/proc/self/fd/%d", fd))
	if err != nil {
		return "", false, err
	}
	return path, true, nil
}

func lookupShell(username string) string {
	const fallback = "/bin/sh"

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
	return fallback
}
