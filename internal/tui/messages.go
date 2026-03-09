package tui

// userSubmitMsg is sent when the user presses Enter with non-empty input.
type userSubmitMsg struct{ text string }
