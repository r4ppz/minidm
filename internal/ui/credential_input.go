package ui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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
	usernameInput.TextStyle = TextInputStyle
	usernameInput.PlaceholderStyle = TextInputPlaceholderStyle

	passwordInput := textinput.New()
	passwordInput.Placeholder = "password"
	passwordInput.EchoMode = textinput.EchoPassword
	passwordInput.EchoCharacter = '•'
	passwordInput.CharLimit = 256
	passwordInput.Width = 30
	passwordInput.Prompt = ""
	passwordInput.TextStyle = TextInputStyle
	passwordInput.PlaceholderStyle = TextInputPlaceholderStyle

	return CredentialInputModel{
		usernameInput: usernameInput,
		passwordInput: passwordInput,
		focusIndex:    0,
	}
}

type credentialInputKeyMap struct {
	next  key.Binding
	prev  key.Binding
	enter key.Binding
}

var credentialInputKeys = credentialInputKeyMap{
	next: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next field"),
	),
	prev: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev field"),
	),
	enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "submit/next"),
	),
}

func (m CredentialInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m CredentialInputModel) Update(msg tea.Msg) (CredentialInputModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.errMsg != "" {
			m.errMsg = ""
		}
		switch {
		case key.Matches(msg, credentialInputKeys.next), key.Matches(msg, credentialInputKeys.prev):
			m.toggleFocus()
		case key.Matches(msg, credentialInputKeys.enter):
			if m.focusIndex == 0 {
				if m.usernameInput.Value() == "" {
					m.errMsg = "Username required"
				} else {
					m.focusIndex = 1
					m.updateFocus()
				}
			} else if m.focusIndex == 1 {
				if m.usernameInput.Value() == "" {
					m.errMsg = "Username required"
				} else if m.passwordInput.Value() == "" {
					m.errMsg = "Password required"
				} else if !m.submitting {
					m.submitting = true
					user := m.usernameInput.Value()
					pass := m.passwordInput.Value()
					return m, func() tea.Msg {
						return CredentialSubmitMsg{Username: user, Password: pass}
					}
				}
			}
		}
	}

	var cmds []tea.Cmd

	if m.focusIndex == 0 {
		m.usernameInput, _ = m.usernameInput.Update(msg)
	} else {
		m.passwordInput, _ = m.passwordInput.Update(msg)
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

	usernameStyle := Input(m.focusIndex == 0)
	passwordStyle := Input(m.focusIndex == 1)

	usernameLabel := LabelStyle.Render("Username:")
	passwordLabel := LabelStyle.Render("Password:")

	viewContent += usernameLabel + usernameStyle.Render(m.usernameInput.View()) + "\n"
	viewContent += passwordLabel + passwordStyle.Render(m.passwordInput.View())

	if m.errMsg != "" {
		viewContent += "\n" + ErrorStyle.Render(m.errMsg)
	}

	viewContent += "\n" + HelpStyle.Render(
		credentialInputKeys.prev.Help().Key+" "+credentialInputKeys.prev.Help().Desc,
		credentialInputKeys.next.Help().Key+" "+credentialInputKeys.next.Help().Desc,
		credentialInputKeys.enter.Help().Key+" "+credentialInputKeys.enter.Help().Desc,
	)

	return viewContent
}

func (m *CredentialInputModel) SetError(err string) {
	m.errMsg = err
	m.submitting = false
}

func (m *CredentialInputModel) ClearPassword() {
	m.passwordInput.Reset()
}

func (m *CredentialInputModel) ClearSubmitting() {
	m.submitting = false
}
