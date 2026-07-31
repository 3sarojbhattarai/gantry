package tui

import "github.com/charmbracelet/lipgloss"

// styles groups the lipgloss styles the view uses. Colors are ANSI 256 so they
// adapt to the user's terminal theme.
type styles struct {
	title     lipgloss.Style
	tab       lipgloss.Style
	activeTab lipgloss.Style
	paneTitle lipgloss.Style
	selected  lipgloss.Style
	dim       lipgloss.Style
	sep       lipgloss.Style
	errText   lipgloss.Style
}

func newStyles() styles {
	return styles{
		title:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("27")).Padding(0, 1),
		tab:       lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Padding(0, 1),
		activeTab: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("240")).Padding(0, 1),
		paneTitle: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("81")),
		selected:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("238")),
		dim:       lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		sep:       lipgloss.NewStyle().Foreground(lipgloss.Color("238")),
		errText:   lipgloss.NewStyle().Foreground(lipgloss.Color("203")),
	}
}
