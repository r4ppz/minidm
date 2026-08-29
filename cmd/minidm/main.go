package main

import (
	"errors"
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
	if err := run(); err != nil {
		log.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	sessions := session.Discover()
	if len(sessions) == 0 {
		return errors.New("no sessions found")
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Info("received SIGTERM, shutting down")
		os.Exit(0)
	}()

	model := ui.NewModel(sessions)
	p := tea.NewProgram(model, tea.WithAltScreen())

	result, err := p.Run()
	if err != nil {
		return err
	}

	final, ok := result.(ui.Model)
	if !ok {
		return errors.New("unexpected program result")
	}
	if !final.Authenticated() {
		return nil
	}

	log.Info("launching session",
		"session", final.AuthenticatedSession().Name,
		"user", final.AuthenticatedUser())

	return auth.RunSession(final.AuthenticatedUser(), final.AuthenticatedSession(), final.PAMTx())
}
