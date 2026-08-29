package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	colorWhite       = lipgloss.Color("#FFFFFF")
	colorGray        = lipgloss.Color("#888888")
	colorLightGray   = lipgloss.Color("#CCCCCC")
	colorDarkGray    = lipgloss.Color("#666666")
	colorError       = lipgloss.Color("#FF5555")
	colorPlaceholder = lipgloss.Color("#666666")

	baseStyle = lipgloss.NewStyle().
			Padding(1, 2)

	sessionListStyle = lipgloss.NewStyle().
				MarginBottom(1)

	currentSessionStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorWhite)

	otherSessionStyle = lipgloss.NewStyle().
				Foreground(colorGray)

	labelStyle = lipgloss.NewStyle().
			Foreground(colorLightGray).
			MarginRight(1)

	inputStyle = lipgloss.NewStyle().
			Padding(0, 1).
			MarginBottom(1)

	focusedInputStyle = lipgloss.NewStyle().
				Padding(0, 1).
				MarginBottom(1).
				Foreground(colorWhite)

	textInputStyle = lipgloss.NewStyle().
			Foreground(colorWhite)

	textInputPlaceholderStyle = lipgloss.NewStyle().
					Foreground(colorPlaceholder)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorError).
			MarginTop(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(colorDarkGray).
			MarginTop(1)

	spinnerStyle = lipgloss.NewStyle().
			Foreground(colorWhite)
)

func BaseStyle() lipgloss.Style {
	return baseStyle
}

func SessionListStyle() lipgloss.Style {
	return sessionListStyle
}

func CurrentSessionStyle() lipgloss.Style {
	return currentSessionStyle
}

func OtherSessionStyle() lipgloss.Style {
	return otherSessionStyle
}

func LabelStyle() lipgloss.Style {
	return labelStyle
}

func InputStyle(focused bool) lipgloss.Style {
	if focused {
		return focusedInputStyle
	}
	return inputStyle
}

func TextInputStyle() lipgloss.Style {
	return textInputStyle
}

func TextInputPlaceholderStyle() lipgloss.Style {
	return textInputPlaceholderStyle
}

func ErrorStyle() lipgloss.Style {
	return errorStyle
}

func HelpStyle() lipgloss.Style {
	return helpStyle
}

func SpinnerStyle() lipgloss.Style {
	return spinnerStyle
}
