package models

import "github.com/charmbracelet/bubbles/key"

var (
	Keys = KeyMap{
		Toggle: key.NewBinding(
			key.WithKeys("w"),
			key.WithHelp("w", "toggle db writing"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
)

// KeyMap holds the keys used by the application.
type KeyMap struct {
	Toggle key.Binding
	Quit   key.Binding
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Toggle, k.Quit}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Toggle, k.Quit},
	}
}
