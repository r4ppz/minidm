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

func (model Model) Init() tea.Cmd {
	return tea.Batch(model.credentialInput.Init(), model.spinner.Tick)
}

func (model Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return model, tea.Quit
		}

	case CredentialSubmitMsg:
		log.Infof("Login attempt: %s", msg.Username)
		selectedSession := model.sessionList.SelectedSession()
		return model, tea.Batch(
			authenticate(msg.Username, msg.Password, selectedSession),
			model.spinner.Tick,
		)

	case AuthResultMsg:
		if msg.Err != nil {
			var authErr *auth.Error
			var userMsg string
			if errors.As(msg.Err, &authErr) {
				userMsg = authErr.Message
			} else {
				userMsg = "Authentication failed"
			}
			log.Errorf("Auth failed for %s: %v", msg.Username, msg.Err)
			model.credentialInput.SetError(userMsg)
			model.credentialInput.ClearPassword()
			model.credentialInput.ClearSubmitting()
		} else {
			log.Infof("Auth success: %s", msg.Username)
			model.authenticated = true
			model.authenticatedUser = msg.Username
			model.authenticatedSession = msg.Session
			model.pamTx = msg.PAMTx
			return model, tea.Quit
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		model.spinner, cmd = model.spinner.Update(msg)
		return model, cmd
	}

	var cmds []tea.Cmd

	sessionListModel, cmd := model.sessionList.Update(msg)
	model.sessionList = sessionListModel
	cmds = append(cmds, cmd)

	credentialInputModel, cmd := model.credentialInput.Update(msg)
	model.credentialInput = credentialInputModel
	cmds = append(cmds, cmd)

	return model, tea.Batch(cmds...)
}

func (model Model) View() string {
	if model.authenticated {
		return model.spinner.View() + " Starting session..."
	}

	return model.sessionList.View() + "\n\n" + model.credentialInput.View()
}

func authenticate(username, password string, sess session.Session) tea.Cmd {
	return func() tea.Msg {
		tx, err := auth.Authenticate(username, password)
		if err != nil {
			return AuthResultMsg{Err: err, Username: username, Session: sess}
		}
		return AuthResultMsg{PAMTx: tx, Username: username, Session: sess}
	}
}

func (model Model) Authenticated() bool                   { return model.authenticated }
func (model Model) AuthenticatedUser() string             { return model.authenticatedUser }
func (model Model) AuthenticatedSession() session.Session { return model.authenticatedSession }
func (model Model) PAMTx() *pam.Transaction               { return model.pamTx }
