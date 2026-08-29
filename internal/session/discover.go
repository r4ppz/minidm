package session

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/r4ppz/minidm/internal/log"
)

func Discover() []Session {
	var sessions []Session

	targets := []struct {
		dir string
		typ Type
	}{
		{"/usr/share/wayland-sessions", Wayland},
		{"/usr/share/xsessions", X11},
	}

	for _, t := range targets {
		files, err := os.ReadDir(t.dir)
		if err != nil {
			log.Debugf("Session directory %s: %v", t.dir, err)
			continue
		}
		for _, file := range files {
			if !strings.HasSuffix(file.Name(), ".desktop") {
				continue
			}
			path := filepath.Join(t.dir, file.Name())
			sess, err := parseDesktopFile(path, t.typ)
			if err != nil {
				log.Debugf("Parse %s: %v", path, err)
				continue
			}
			if sess.Exec != "" {
				sessions = append(sessions, sess)
			}
		}
	}

	log.Infof("Discovered %d sessions", len(sessions))
	return sessions
}
