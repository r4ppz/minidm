package main

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	minidm "github.com/r4ppz/minidm/internal"
	"github.com/r4ppz/minidm/internal/ui"
)

func main() {
	sessions, err := minidm.DiscoverSessions()
	if err != nil {
		minidm.Errorf("Failed to discover sessions: %v", err)
		os.Exit(1)
	}

	if len(sessions) == 0 {
		minidm.Errorf("No sessions found")
		os.Exit(1)
	}

	model := ui.NewModel(sessions)
	p := tea.NewProgram(model, tea.WithAltScreen())

	result, err := p.Run()
	if err != nil {
		minidm.Errorf("TUI error: %v", err)
		os.Exit(1)
	}

	final := result.(ui.Model)

	if !final.Authenticated() {
		return
	}

	minidm.Infof("Launching session %s for user %s", final.AuthenticatedSession().Name, final.AuthenticatedUser())

	if err := minidm.RunSession(final.AuthenticatedUser(), final.AuthenticatedSession(), final.PAMTx()); err != nil {
		minidm.Errorf("Session error: %v", err)
		os.Exit(1)
	}
}
