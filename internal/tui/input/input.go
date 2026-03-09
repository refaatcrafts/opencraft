package input

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"
)

// Model wraps the bubbles textarea with custom styling.
type Model struct {
	Ta textarea.Model
}

func New() Model {
	ta := textarea.New()
	ta.Placeholder = "Type a message..."
	ta.ShowLineNumbers = false
	ta.CharLimit = 0
	ta.SetHeight(3)
	ta.FocusedStyle.Base = lipgloss.NewStyle()
	ta.BlurredStyle.Base = lipgloss.NewStyle()
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.Focus()
	return Model{Ta: ta}
}

func (m *Model) SetWidth(w int) {
	inner := w - 4
	if inner < 4 {
		inner = 4
	}
	m.Ta.SetWidth(inner)
}
