package ui

import "github.com/charmbracelet/lipgloss"

var (
	white       = lipgloss.Color("#FFFFFF")
	gray        = lipgloss.Color("#888888")
	lightGray   = lipgloss.Color("#CCCCCC")
	darkGray    = lipgloss.Color("#666666")
	errorColor  = lipgloss.Color("#FF5555")
	placeholder = lipgloss.Color("#666666")
)

var (
	SessionListStyle = lipgloss.NewStyle().
				MarginBottom(1)

	CurrentSessionStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(white)

	OtherSessionStyle = lipgloss.NewStyle().
				Foreground(gray)

	LabelStyle = lipgloss.NewStyle().
			Foreground(lightGray).
			MarginRight(1)

	InputStyle = lipgloss.NewStyle().
			Padding(0, 1).
			MarginBottom(1)

	FocusedInputStyle = lipgloss.NewStyle().
				Padding(0, 1).
				MarginBottom(1).
				Foreground(white)

	TextInputStyle = lipgloss.NewStyle().
			Foreground(white)

	TextInputPlaceholderStyle = lipgloss.NewStyle().
					Foreground(placeholder)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			MarginTop(1)

	HelpStyle = lipgloss.NewStyle().
			Foreground(darkGray).
			MarginTop(1)

	SpinnerStyle = lipgloss.NewStyle().
			Foreground(white)
)

func Input(focused bool) lipgloss.Style {
	if focused {
		return FocusedInputStyle
	}
	return InputStyle
}
