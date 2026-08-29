package session

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

func parseDesktopFile(path string, typ Type) (Session, error) {
	file, err := os.Open(path)
	if err != nil {
		return Session{}, err
	}
	defer file.Close()

	sess := Session{
		ID:   strings.TrimSuffix(filepath.Base(path), ".desktop"),
		Type: typ,
	}

	inDesktopEntry := false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			inDesktopEntry = line == "[Desktop Entry]"
			continue
		}
		if !inDesktopEntry || !strings.Contains(line, "=") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "Name":
			if sess.Name == "" {
				sess.Name = value
			}
		case "Exec":
			if sess.Exec == "" {
				sess.Exec = value
			}
		case "DesktopNames":
			if sess.DesktopName == "" {
				sess.DesktopName = value
			}
		}
	}

	if sess.Name == "" {
		sess.Name = sess.ID
	}
	if sess.DesktopName == "" {
		sess.DesktopName = sess.Name
	}
	return sess, scanner.Err()
}
