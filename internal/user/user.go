package user

import (
	"bufio"
	"os"
	ouser "os/user"
	"strconv"
	"strings"

	"github.com/r4ppz/minidm/internal/log"
)

type Info struct {
	Uid     uint32
	Gid     uint32
	Groups  []uint32
	HomeDir string
	Shell   string
}

func Lookup(username string) (*Info, error) {
	log.Debugf("Looking up user: %s", username)
	osUser, err := ouser.Lookup(username)
	if err != nil {
		log.Errorf("Failed to lookup user %s: %v", username, err)
		return nil, err
	}

	uid, err := strconv.ParseUint(osUser.Uid, 10, 32)
	if err != nil {
		log.Errorf("Failed to parse UID for user %s: %v", username, err)
		return nil, err
	}
	gid, err := strconv.ParseUint(osUser.Gid, 10, 32)
	if err != nil {
		log.Errorf("Failed to parse GID for user %s: %v", username, err)
		return nil, err
	}

	info := &Info{
		Uid:     uint32(uid),
		Gid:     uint32(gid),
		HomeDir: osUser.HomeDir,
		Shell:   lookupShell(username),
	}

	info.Groups = append(info.Groups, uint32(gid))
	groupIDs, err := osUser.GroupIds()
	if err == nil {
		for _, groupID := range groupIDs {
			if parsed, err := strconv.ParseUint(groupID, 10, 32); err == nil {
				info.Groups = append(info.Groups, uint32(parsed))
			}
		}
	}
	log.Debugf("User %s lookup successful: UID=%d, GID=%d, Home=%s, Shell=%s", username, info.Uid, info.Gid, info.HomeDir, info.Shell)
	return info, nil
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
	if err := scanner.Err(); err != nil {
		return fallback
	}
	return fallback
}
