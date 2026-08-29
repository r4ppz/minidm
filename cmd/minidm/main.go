package main

import (
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/r4ppz/minidm/internal/auth"
	"github.com/r4ppz/minidm/internal/log"
	"github.com/r4ppz/minidm/internal/session"
	"github.com/r4ppz/minidm/internal/ui"
)

func main() {
	sessions := session.Discover()

	if len(sessions) == 0 {
		log.Errorf("No sessions found")
		os.Exit(1)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Infof("Received SIGTERM, shutting down")
		os.Exit(0)
	}()

	model := ui.NewModel(sessions)
	p := tea.NewProgram(model, tea.WithAltScreen())

	result, err := p.Run()
	if err != nil {
		log.Errorf("TUI error: %v", err)
		os.Exit(1)
	}

	final := result.(ui.Model)

	if !final.Authenticated() {
		return
	}

	log.Infof("Launching session %s for user %s", final.AuthenticatedSession().Name, final.AuthenticatedUser())

	if err := auth.RunSession(final.AuthenticatedUser(), final.AuthenticatedSession(), final.PAMTx()); err != nil {
		log.Errorf("Session error: %v", err)
		os.Exit(1)
	}
}
