package styles

import "github.com/charmbracelet/lipgloss"

var (
	ColorBlue   = lipgloss.Color("#65BFCC")
	ColorPurple = lipgloss.Color("#CF58D8")

	HeaderName = lipgloss.
			NewStyle().
			Foreground(ColorBlue).
			Bold(true)

	SectionHeader = lipgloss.
			NewStyle().
			Bold(true).
			Underline(true).
			MarginTop(1).
			MarginBottom(1).
			Padding(0)

	Url = lipgloss.
		NewStyle().
		Foreground(ColorPurple).
		Bold(true)
)
