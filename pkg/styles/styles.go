package styles

import "github.com/charmbracelet/lipgloss"

var (
	ColorBlue   = lipgloss.AdaptiveColor{Dark: "#65BFCC", Light: "#00008b"}
	ColorPurple = lipgloss.AdaptiveColor{Dark: "#CF58D8", Light: "#be02bf"}
	ColorGray   = lipgloss.AdaptiveColor{Dark: "#B2B2B2", Light: "#000000"}

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

	Divider = lipgloss.
		NewStyle().
		Bold(true).
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
