package tui

import (
	"context"

	"github.com/3sarojbhattarai/gantry/internal/docker"
	tea "github.com/charmbracelet/bubbletea"
)

// Run starts the read-only TUI against client, taking over the terminal until
// the user quits or ctx is cancelled.
func Run(ctx context.Context, client docker.Client) error {
	p := tea.NewProgram(
		New(ctx, client),
		tea.WithAltScreen(),
		tea.WithContext(ctx),
	)
	_, err := p.Run()
	return err
}
