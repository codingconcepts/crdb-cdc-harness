package styles

import "github.com/charmbracelet/lipgloss"

var (
	StatsAreaColour             = "#6933ff"
	WritingAreadActiveColour    = "#00fced"
	WritingBorderActiveColour   = "#00fced"
	WritingAreaInactiveColour   = "#ec3f96"
	WritingBorderInactiveColour = "#ec3f96"
	StatsLabelColour            = "#394455"
	StatsValueColour            = "#d6dbe7"
	HelpAreaBorderColour        = "#394455"

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(StatsAreaColour)).
			Padding(1).
			Width(45)

	ActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(WritingAreadActiveColour)).
			PaddingLeft(1)

	InactiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(WritingAreaInactiveColour)).
			PaddingLeft(1)

	StatsLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(StatsLabelColour))

	StatsValueStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(StatsValueColour))
)
