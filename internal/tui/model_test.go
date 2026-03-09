package tui

import (
	"strings"
	"testing"

	"opencraft/internal/config"
)

func TestUpdateLayoutAndSizeMarksTooSmall(t *testing.T) {
	m := New(&config.Config{Model: "demo"}, nil)
	m.width = minWidth - 1
	m.height = minHeight - 1

	m.updateLayoutAndSize()

	if !m.layout.tooSmall {
		t.Fatal("expected tooSmall layout")
	}
}

func TestUpdateLayoutAndSizeEnablesCompactMode(t *testing.T) {
	m := New(&config.Config{Model: "demo"}, nil)
	m.width = compactWidthBreakpoint - 1
	m.height = compactHeightBreakpoint + 4

	m.updateLayoutAndSize()

	if m.layout.tooSmall {
		t.Fatal("did not expect tooSmall layout")
	}
	if !m.layout.compact {
		t.Fatal("expected compact layout")
	}
}

func TestRefreshViewportPreservesOffsetWhenNotAtBottom(t *testing.T) {
	m := New(&config.Config{Model: "demo"}, nil)
	m.width = 100
	m.height = 30
	m.updateLayoutAndSize()

	for i := 0; i < 20; i++ {
		m.chat.AddAssistant("line " + string(rune('A'+i)))
	}
	m.refreshViewport(true)
	m.viewport.ScrollUp(5)
	offset := m.viewport.YOffset

	m.chat.AddAssistant("new content")
	m.refreshViewport(false)

	if m.viewport.YOffset != offset {
		t.Fatalf("viewport offset = %d, want %d", m.viewport.YOffset, offset)
	}
}

func TestViewportMatchesChatContentBox(t *testing.T) {
	m := New(&config.Config{Model: "demo"}, nil)
	m.width = 100
	m.height = 30

	m.updateLayoutAndSize()

	wantWidth := m.layout.panelWidth - m.chatPanelStyle().GetHorizontalFrameSize()
	wantHeight := m.layout.chatHeight - m.chatPanelStyle().GetVerticalFrameSize()
	if m.viewport.Width != wantWidth {
		t.Fatalf("viewport width = %d, want %d", m.viewport.Width, wantWidth)
	}
	if m.viewport.Height != wantHeight {
		t.Fatalf("viewport height = %d, want %d", m.viewport.Height, wantHeight)
	}
}

func TestStatusSummaryReflectsThinkingAndToolStates(t *testing.T) {
	m := New(&config.Config{Model: "demo"}, nil)

	m.state = stateThinking
	if got := m.statusSummary(); got == "Ready" || got == "" {
		t.Fatalf("expected thinking summary, got %q", got)
	}

	m.state = stateToolRunning
	m.activeToolName = "list_dir"
	if got := m.statusSummary(); got == "Ready" || got == "" || !contains(got, "list_dir") {
		t.Fatalf("expected tool status summary, got %q", got)
	}
}

func contains(s, sub string) bool {
	return strings.Contains(s, sub)
}
