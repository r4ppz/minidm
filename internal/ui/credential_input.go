package ui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	minidm "github.com/r4ppz/minidm/internal"
)

type CredentialInputModel struct {
	usernameInput textinput.Model
	passwordInput textinput.Model
	focusIndex    int
	errMsg        string
	submitting    bool
}

func NewCredentialInputModel() CredentialInputModel {
	usernameInput := textinput.New()
	usernameInput.Placeholder = "username"
	usernameInput.Focus()
	usernameInput.CharLimit = 256
	usernameInput.Width = 30
	usernameInput.Prompt = ""
	usernameInput.TextStyle = TextInputStyle()
	usernameInput.PlaceholderStyle = TextInputPlaceholderStyle()

	passwordInput := textinput.New()
	passwordInput.Placeholder = "password"
	passwordInput.EchoMode = textinput.EchoPassword
	passwordInput.EchoCharacter = '•'
	passwordInput.CharLimit = 256
	passwordInput.Width = 30
	passwordInput.Prompt = ""
	passwordInput.TextStyle = TextInputStyle()
	passwordInput.PlaceholderStyle = TextInputPlaceholderStyle()

	return CredentialInputModel{
		usernameInput: usernameInput,
		passwordInput: passwordInput,
		focusIndex:    0,
	}
}

type credentialInputKeyMap struct {
	up    key.Binding
	down  key.Binding
	enter key.Binding
	esc   key.Binding
}

var credentialInputKeys = credentialInputKeyMap{
	up: key.NewBinding(
		key.WithKeys("up", "shift+tab"),
		key.WithHelp("↑/shift+tab", "prev field"),
	),
	down: key.NewBinding(
		key.WithKeys("down", "tab"),
		key.WithHelp("↓/tab", "next field"),
	),
	enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "submit/next"),
	),
	esc: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "clear error"),
	),
}

func (m CredentialInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m CredentialInputModel) Update(msg tea.Msg) (CredentialInputModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, credentialInputKeys.esc):
			m.errMsg = ""
		case key.Matches(msg, credentialInputKeys.up), key.Matches(msg, credentialInputKeys.down):
			m.toggleFocus()
		case key.Matches(msg, credentialInputKeys.enter):
			if m.focusIndex == 0 {
				if m.usernameInput.Value() != "" {
					m.focusIndex = 1
					m.updateFocus()
				}
			} else if m.focusIndex == 1 {
				if !m.submitting && m.usernameInput.Value() != "" && m.passwordInput.Value() != "" {
					m.submitting = true
					return m, func() tea.Msg {
						return CredentialSubmitMsg{
							Username: m.usernameInput.Value(),
							Password: m.passwordInput.Value(),
							Session:  minidm.Session{},
						}
					}
				}
			}
		}
	}

	if m.focusIndex == 0 {
		m.usernameInput, _ = m.usernameInput.Update(msg)
		cmds = append(cmds, textinput.Blink)
	} else {
		m.passwordInput, _ = m.passwordInput.Update(msg)
		cmds = append(cmds, textinput.Blink)
	}

	return m, tea.Batch(cmds...)
}

func (m *CredentialInputModel) toggleFocus() {
	m.focusIndex = (m.focusIndex + 1) % 2
	m.updateFocus()
}

func (m *CredentialInputModel) updateFocus() {
	if m.focusIndex == 0 {
		m.usernameInput.Focus()
		m.passwordInput.Blur()
	} else {
		m.usernameInput.Blur()
		m.passwordInput.Focus()
	}
}

func (m CredentialInputModel) View() string {
	var viewContent string

	usernameStyle := InputStyle(m.focusIndex == 0)
	passwordStyle := InputStyle(m.focusIndex == 1)

	usernameLabel := LabelStyle().Render("Username:")
	passwordLabel := LabelStyle().Render("Password:")

	viewContent += usernameLabel + usernameStyle.Render(m.usernameInput.View()) + "\n"
	viewContent += passwordLabel + passwordStyle.Render(m.passwordInput.View())

	if m.errMsg != "" {
		viewContent += "\n" + ErrorStyle().Render(m.errMsg)
	}

	viewContent += "\n" + HelpStyle().Render(
		credentialInputKeys.up.Help().Key+" "+credentialInputKeys.up.Help().Desc,
		credentialInputKeys.down.Help().Key+" "+credentialInputKeys.down.Help().Desc,
		credentialInputKeys.enter.Help().Key+" "+credentialInputKeys.enter.Help().Desc,
		credentialInputKeys.esc.Help().Key+" "+credentialInputKeys.esc.Help().Desc,
	)

	return viewContent
}

func (m *CredentialInputModel) SetError(err string) {
	m.errMsg = err
	m.submitting = false
}

func (m *CredentialInputModel) SetSession(session minidm.Session) {
	m.usernameInput.Reset()
	m.passwordInput.Reset()
	m.errMsg = ""
	m.submitting = false
	m.focusIndex = 0
	m.updateFocus()
}

func (m CredentialInputModel) Submitting() bool {
	return m.submitting
}

func (m CredentialInputModel) Username() string {
	return m.usernameInput.Value()
}

func (m CredentialInputModel) Password() string {
	return m.passwordInput.Value()
}

func (m *CredentialInputModel) Reset() {
	m.usernameInput.Reset()
	m.passwordInput.Reset()
	m.errMsg = ""
	m.submitting = false
	m.focusIndex = 0
	m.updateFocus()
}
