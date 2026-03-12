package tui

// userSubmitMsg is sent when the user presses Enter with non-empty input.
type userSubmitMsg struct{ text string }

// internalCommandRunMsg starts an agent run without posting a user chat item.
type internalCommandRunMsg struct {
	prompt      string
	hidden      bool
	statusLabel string
}
