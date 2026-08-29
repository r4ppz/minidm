package ui

import (
	"errors"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/msteinert/pam/v2"
	"github.com/r4ppz/minidm/internal/auth"
	"github.com/r4ppz/minidm/internal/log"
	"github.com/r4ppz/minidm/internal/session"
)

type Model struct {
	sessionList     SessionListModel
	credentialInput CredentialInputModel
	spinner         spinner.Model

	authenticated        bool
	authenticatedUser    string
	authenticatedSession session.Session
	pamTx                *pam.Transaction
}

func NewModel(sessions []session.Session) Model {
	spin := spinner.New()
	spin.Spinner = spinner.Dot
	spin.Style = SpinnerStyle

	return Model{
		sessionList:     NewSessionListModel(sessions),
		credentialInput: NewCredentialInputModel(),
		spinner:         spin,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.credentialInput.Init(), m.spinner.Tick)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

	case CredentialSubmitMsg:
		log.Info("login attempt", "user", msg.Username)
		return m, tea.Batch(
			authenticate(msg.Username, msg.Password, m.sessionList.SelectedSession()),
			m.spinner.Tick,
		)

	case AuthResultMsg:
		if msg.Err != nil {
			log.Error("auth failed", "user", msg.Username, "err", msg.Err)
			m.credentialInput.SetError(errorMessage(msg.Err))
			m.credentialInput.ClearPassword()
			m.credentialInput.ClearSubmitting()
		} else {
			log.Info("auth success", "user", msg.Username)
			m.authenticated = true
			m.authenticatedUser = msg.Username
			m.authenticatedSession = msg.Session
			m.pamTx = msg.PAMTx
			return m, tea.Quit
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	var cmds []tea.Cmd

	sessionListModel, cmd := m.sessionList.Update(msg)
	m.sessionList = sessionListModel
	cmds = append(cmds, cmd)

	credentialInputModel, cmd := m.credentialInput.Update(msg)
	m.credentialInput = credentialInputModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	view := m.sessionList.View() + "\n\n" + m.credentialInput.View()

	if m.credentialInput.Submitting() {
		view += "\n" + m.spinner.View() + " Authenticating..."
	} else if m.authenticated {
		view += "\n" + m.spinner.View() + " Starting session..."
	}

	return view
}

func authenticate(username, password string, sess session.Session) tea.Cmd {
	return func() tea.Msg {
		tx, err := auth.Authenticate(username, password)
		return AuthResultMsg{
			Err:      err,
			PAMTx:    tx,
			Username: username,
			Session:  sess,
		}
	}
}

func errorMessage(err error) string {
	var authErr *auth.Error
	if errors.As(err, &authErr) {
		return authErr.Message
	}
	return "Authentication failed"
}

func (m Model) Authenticated() bool                   { return m.authenticated }
func (m Model) AuthenticatedUser() string             { return m.authenticatedUser }
func (m Model) AuthenticatedSession() session.Session { return m.authenticatedSession }
func (m Model) PAMTx() *pam.Transaction               { return m.pamTx }
