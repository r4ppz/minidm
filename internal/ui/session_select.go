package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	minidm "github.com/r4ppz/minidm/internal"
)

type SessionListModel struct {
	sessions    []minidm.Session
	selectedIdx int
}

func NewSessionListModel(sessions []minidm.Session) SessionListModel {
	return SessionListModel{
		sessions:    sessions,
		selectedIdx: 0,
	}
}

type sessionListKeyMap struct {
	next key.Binding
	prev key.Binding
}

var sessionListKeys = sessionListKeyMap{
	next: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "next session"),
	),
	prev: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "prev session"),
	),
}

func (m SessionListModel) Init() tea.Cmd {
	return nil
}

func (m SessionListModel) Update(msg tea.Msg) (SessionListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, sessionListKeys.next):
			m = m.nextSession()
			return m, nil

		case key.Matches(msg, sessionListKeys.prev):
			m = m.prevSession()
			return m, nil
		}
	}

	return m, nil
}

func (m SessionListModel) View() string {
	if len(m.sessions) == 0 {
		return ""
	}

	var viewContent string
	if len(m.sessions) == 1 {
		viewContent += "Session:\n"
	} else {
		viewContent += fmt.Sprintf("Sessions (%d):\n", len(m.sessions))
	}

	for i, sess := range m.sessions {
		prefix := "  "
		if i == m.selectedIdx {
			prefix = CurrentSessionStyle.Render("▸ ")
			viewContent += prefix + CurrentSessionStyle.Render(sess.Name+" ("+string(sess.Type)+")")
		} else {
			viewContent += prefix + OtherSessionStyle.Render(sess.Name+" ("+string(sess.Type)+")")
		}
		if i < len(m.sessions)-1 {
			viewContent += "\n"
		}
	}

	return SessionListStyle.Render(viewContent)
}

func (m SessionListModel) nextSession() SessionListModel {
	if len(m.sessions) > 0 {
		m.selectedIdx = (m.selectedIdx + 1) % len(m.sessions)
	}
	return m
}

func (m SessionListModel) prevSession() SessionListModel {
	if len(m.sessions) > 0 {
		m.selectedIdx = (m.selectedIdx - 1 + len(m.sessions)) % len(m.sessions)
	}
	return m
}

func (m SessionListModel) SelectedSession() minidm.Session {
	if len(m.sessions) > 0 && m.selectedIdx < len(m.sessions) {
		return m.sessions[m.selectedIdx]
	}
	return minidm.Session{}
}

func (m SessionListModel) SetSessions(sessions []minidm.Session) SessionListModel {
	m.sessions = sessions
	if m.selectedIdx >= len(m.sessions) {
		if len(m.sessions) > 0 {
			m.selectedIdx = len(m.sessions) - 1
		} else {
			m.selectedIdx = 0
		}
	}
	return m
}
