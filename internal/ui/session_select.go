package ui

import (
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
		key.WithKeys("tab"),
		key.WithHelp("tab", "next session"),
	),
	prev: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev session"),
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
			if len(m.sessions) > 0 {
				m.selectedIdx = (m.selectedIdx + 1) % len(m.sessions)
			}
		case key.Matches(msg, sessionListKeys.prev):
			if len(m.sessions) > 0 {
				m.selectedIdx = (m.selectedIdx - 1 + len(m.sessions)) % len(m.sessions)
			}
		}
	}
	return m, nil
}

func (m SessionListModel) View() string {
	if len(m.sessions) == 0 {
		return ""
	}

	var viewContent string
	viewContent += "Session: "

	for i, sess := range m.sessions {
		if i > 0 {
			viewContent += "  "
		}

		if i == m.selectedIdx {
			viewContent += CurrentSessionStyle().Render("▸ " + sess.Name + " (" + string(sess.Type) + ")")
		} else {
			viewContent += OtherSessionStyle().Render(sess.Name + " (" + string(sess.Type) + ")")
		}
	}

	return SessionListStyle().Render(viewContent)
}

func (m SessionListModel) SelectedSession() minidm.Session {
	if len(m.sessions) > 0 && m.selectedIdx < len(m.sessions) {
		return m.sessions[m.selectedIdx]
	}
	return minidm.Session{}
}

func (m *SessionListModel) SetSessions(sessions []minidm.Session) {
	m.sessions = sessions
	if m.selectedIdx >= len(sessions) {
		m.selectedIdx = 0
	}
}

func (m *SessionListModel) SetSelectedIdx(idx int) {
	if idx >= 0 && idx < len(m.sessions) {
		m.selectedIdx = idx
	}
}
