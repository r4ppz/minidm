package main

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	minidm "github.com/r4ppz/minidm/internal"
	"github.com/r4ppz/minidm/internal/ui"
)

func main() {
	if os.Getuid() != 0 {
		println("minidm must be run as root")
		os.Exit(1)
	}

	sessions := loadSessions()

	m := ui.NewModel(sessions)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		println("Error running TUI:", err.Error())
		os.Exit(1)
	}
}

func loadSessions() []minidm.Session {
	sessions, err := minidm.DiscoverSessions()
	if err == nil && len(sessions) > 0 {
		return sessions
	}

	return []minidm.Session{
		{
			ID:          "shell",
			Name:        "Default Shell",
			Exec:        "/bin/bash",
			DesktopName: "Shell",
			Type:        minidm.SessionWayland,
		},
	}
}
