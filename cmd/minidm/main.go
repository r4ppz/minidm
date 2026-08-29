package main

import (
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	minidm "github.com/r4ppz/minidm/internal"
	"github.com/r4ppz/minidm/internal/ui"
)

var sessionProc *os.Process

func main() {
	sessions := minidm.DiscoverSessions()

	if len(sessions) == 0 {
		minidm.Errorf("No sessions found")
		os.Exit(1)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM)
	go func() {
		<-sigCh
		minidm.Infof("Received SIGTERM, shutting down")
		if sessionProc != nil {
			sessionProc.Signal(syscall.SIGTERM)
		}
		os.Exit(0)
	}()

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

	proc, err := minidm.RunSession(final.AuthenticatedUser(), final.AuthenticatedSession(), final.PAMTx())
	if err != nil {
		minidm.Errorf("Session error: %v", err)
		os.Exit(1)
	}
	sessionProc = proc
}
