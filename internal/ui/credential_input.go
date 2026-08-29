package ui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type FocusField int

const (
	FocusUsername FocusField = iota
	FocusPassword
)

type CredentialInputModel struct {
	usernameInput textinput.Model
	passwordInput textinput.Model
	focused       FocusField
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
		focused:       FocusUsername,
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
			m = m.toggleFocus()
			return m, nil

		case key.Matches(msg, credentialInputKeys.enter):
			switch m.focused {
			case FocusUsername:
				if m.usernameInput.Value() == "" {
					m.errMsg = "Username required"
				} else {
					m.focused = FocusPassword
					m = m.updateFocus()
				}
				return m, nil

			case FocusPassword:
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
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	switch m.focused {
	case FocusUsername:
		m.usernameInput, cmd = m.usernameInput.Update(msg)
	case FocusPassword:
		m.passwordInput, cmd = m.passwordInput.Update(msg)
	}

	return m, cmd
}

func (m CredentialInputModel) toggleFocus() CredentialInputModel {
	if m.focused == FocusUsername {
		m.focused = FocusPassword
	} else {
		m.focused = FocusUsername
	}
	return m.updateFocus()
}

func (m CredentialInputModel) updateFocus() CredentialInputModel {
	if m.focused == FocusUsername {
		m.usernameInput.Focus()
		m.passwordInput.Blur()
	} else {
		m.usernameInput.Blur()
		m.passwordInput.Focus()
	}
	return m
}

func (m CredentialInputModel) View() string {
	var viewContent string

	usernameStyle := Input(m.focused == FocusUsername)
	passwordStyle := Input(m.focused == FocusPassword)

	usernameLabel := LabelStyle.Render("Username:")
	passwordLabel := LabelStyle.Render("Password:")

	viewContent += usernameLabel + usernameStyle.Render(m.usernameInput.View()) + "\n"
	viewContent += passwordLabel + passwordStyle.Render(m.passwordInput.View())

	if m.errMsg != "" {
		viewContent += "\n" + ErrorStyle.Render(m.errMsg)
	}

	if m.submitting {
		return viewContent
	}

	viewContent += "\n" + HelpStyle.Render(
		credentialInputKeys.prev.Help().Key+" "+credentialInputKeys.prev.Help().Desc,
		credentialInputKeys.next.Help().Key+" "+credentialInputKeys.next.Help().Desc,
		credentialInputKeys.enter.Help().Key+" "+credentialInputKeys.enter.Help().Desc,
	)

	return viewContent
}

func (m CredentialInputModel) Submitting() bool {
	return m.submitting
}

func (m CredentialInputModel) SetError(err string) CredentialInputModel {
	m.errMsg = err
	m.submitting = false
	return m
}

func (m CredentialInputModel) ClearPassword() CredentialInputModel {
	m.passwordInput.Reset()
	return m
}

func (m CredentialInputModel) ClearSubmitting() CredentialInputModel {
	m.submitting = false
	return m
}
