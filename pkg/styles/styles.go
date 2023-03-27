package styles

import "github.com/charmbracelet/lipgloss"

var (
	ColorBlue   = lipgloss.Color("#65BFCC")
	ColorPurple = lipgloss.Color("#CF58D8")
	ColorGray   = lipgloss.Color("#B2B2B2")

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

	Description = lipgloss.
			NewStyle().
			Faint(true).
			Foreground(ColorGray)

	PrimaryText = lipgloss.
			NewStyle().
			Foreground(ColorPurple)

	SecondaryText = lipgloss.
			NewStyle().
			Foreground(ColorBlue)
)
