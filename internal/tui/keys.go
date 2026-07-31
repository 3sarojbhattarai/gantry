package tui

import (
	"github.com/charmbracelet/bubbles/key"
)

// keyMap holds the TUI's bindings. Keys mirror lazydocker where practical so
// muscle memory transfers.
type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	Tab      key.Binding
	ShiftTab key.Binding
	Cat      key.Binding // 1-4 select a category
	Refresh  key.Binding
	Start    key.Binding
	Stop     key.Binding
	Restart  key.Binding
	Kill     key.Binding
	Remove   key.Binding
	Exec     key.Binding
	Help     key.Binding
	Quit     key.Binding
}

func defaultKeys() keyMap {
	return keyMap{
		Up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Tab:      key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next pane")),
		ShiftTab: key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("⇧tab", "prev pane")),
		Cat:      key.NewBinding(key.WithKeys("1", "2", "3", "4"), key.WithHelp("1-4", "jump to pane")),
		Refresh:  key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
		Start:    key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "start")),
		Stop:     key.NewBinding(key.WithKeys("S"), key.WithHelp("S", "stop")),
		Restart:  key.NewBinding(key.WithKeys("R"), key.WithHelp("R", "restart")),
		Kill:     key.NewBinding(key.WithKeys("K"), key.WithHelp("K", "kill")),
		Remove:   key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "remove")),
		Exec:     key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "exec shell")),
		Help:     key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		Quit:     key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	}
}

// ShortHelp and FullHelp satisfy help.KeyMap.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Tab, k.Start, k.Stop, k.Remove, k.Exec, k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Tab, k.ShiftTab, k.Cat},
		{k.Start, k.Stop, k.Restart, k.Kill, k.Remove, k.Exec},
		{k.Refresh, k.Help, k.Quit},
	}
}
