package ui

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	minidm "github.com/r4ppz/minidm/internal"
)

type Model struct {
	sessions        []minidm.Session
	sessionList     SessionListModel
	credentialInput CredentialInputModel
	spinner         spinner.Model
	authenticating  bool
}

func NewModel(sessions []minidm.Session) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle()

	return Model{
		sessions:        sessions,
		sessionList:     NewSessionListModel(sessions),
		credentialInput: NewCredentialInputModel(),
		spinner:         s,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.sessionList.Init(), m.credentialInput.Init(), m.spinner.Tick)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

	case CredentialSubmitMsg:
		m.authenticating = true
		selectedSession := m.sessionList.SelectedSession()
		return m, tea.Batch(m.authenticate(msg.Username, msg.Password, selectedSession), m.spinner.Tick)

	case AuthResultMsg:
		m.authenticating = false
		if msg.Err != nil {
			m.credentialInput.SetError(msg.Err.Error())
			cmds = append(cmds, m.credentialInput.Init())
		} else {
			return m, tea.Quit
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	sessionListModel, cmd := m.sessionList.Update(msg)
	m.sessionList = sessionListModel
	cmds = append(cmds, cmd)

	credentialInputModel, cmd := m.credentialInput.Update(msg)
	m.credentialInput = credentialInputModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	var viewContent string

	if m.authenticating {
		viewContent += m.sessionList.View() + "\n\n"
		viewContent += m.spinner.View() + " Authenticating..."
		return BaseStyle().Render(viewContent)
	}

	viewContent += m.sessionList.View() + "\n\n"
	viewContent += m.credentialInput.View()

	return BaseStyle().Render(viewContent)
}

func (m *Model) authenticate(username, password string, session minidm.Session) tea.Cmd {
	return func() tea.Msg {
		err := minidm.Login(username, password, session)
		return AuthResultMsg{Err: err}
	}
}
