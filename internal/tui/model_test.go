package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"opencraft/internal/agent"
	"opencraft/internal/config"
	"opencraft/internal/tui/chat"
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

func TestSlashOpensCommandPaletteWhenInputEmpty(t *testing.T) {
	m := New(&config.Config{Model: "demo"}, nil)
	m.state = stateIdle

	updated, _ := m.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'/'},
	})
	next := updated.(Model)

	if !next.paletteOpen {
		t.Fatal("expected command palette to open when typing / in empty composer")
	}
}

func TestInitAGENTSCommandDispatchesHiddenRun(t *testing.T) {
	m := New(&config.Config{Model: "demo"}, nil)
	m.cwd = filepath.Join(string(filepath.Separator), "tmp", "demo")
	m.paletteItems = []paletteCommand{
		{id: commandInitAGENTS, label: "Initialize AGENTS.md", shortcut: "enter"},
	}
	m.paletteIndex = 0

	cmd := m.runSelectedPaletteCommand()
	if cmd == nil {
		t.Fatal("expected command init action to return an internal run command")
	}
	msg := cmd()

	runMsg, ok := msg.(internalCommandRunMsg)
	if !ok {
		t.Fatalf("expected internalCommandRunMsg, got %T", msg)
	}
	if !runMsg.hidden {
		t.Fatal("expected init command run to be hidden")
	}
	if runMsg.statusLabel != "Generating AGENTS.md with AI" {
		t.Fatalf("statusLabel = %q", runMsg.statusLabel)
	}
	if !strings.Contains(runMsg.prompt, "Overwrite AGENTS.md if it already exists") {
		t.Fatalf("prompt does not enforce overwrite behavior:\n%s", runMsg.prompt)
	}
	if !strings.Contains(runMsg.prompt, "Do not create backups") {
		t.Fatalf("prompt does not enforce no-backup behavior:\n%s", runMsg.prompt)
	}
}

func TestStatusSummaryShowsAGENTSLoadedWhileThinking(t *testing.T) {
	m := New(&config.Config{Model: "demo"}, nil)
	m.state = stateThinking
	m.guidelinesPath = "AGENTS.md"

	got := m.statusSummary()
	if !strings.Contains(got, "AGENTS.md loaded") {
		t.Fatalf("expected AGENTS loaded indicator, got %q", got)
	}
}

func TestGuidelinesLoadedMessageUsesRelativePath(t *testing.T) {
	m := New(&config.Config{Model: "demo"}, nil)
	m.cwd = t.TempDir()
	m.state = stateThinking

	nested := filepath.Join(m.cwd, "nested", "AGENTS.md")
	if err := os.MkdirAll(filepath.Dir(nested), 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}

	updated, _ := m.Update(agent.GuidelinesLoadedMsg{Path: nested})
	next := updated.(Model)
	if next.guidelinesPath != filepath.Join("nested", "AGENTS.md") {
		t.Fatalf("guidelinesPath = %q", next.guidelinesPath)
	}
	if len(next.chat.Items) != 2 {
		t.Fatalf("expected 2 chat items (tool call/result), got %d", len(next.chat.Items))
	}
	call := next.chat.Items[0]
	if call.ToolName != "AGENTS.md" {
		t.Fatalf("tool call name = %q", call.ToolName)
	}
	if call.Kind != chat.KindToolCall {
		t.Fatalf("expected first item to be tool call, got kind=%v", call.Kind)
	}
	expectedPath := filepath.ToSlash(filepath.Join("nested", "AGENTS.md"))
	if !strings.Contains(call.Content, `"event":"loaded"`) || !strings.Contains(call.Content, `"path":"`+expectedPath+`"`) {
		t.Fatalf("tool call content = %q", call.Content)
	}
	result := next.chat.Items[1]
	if result.ToolName != "AGENTS.md" {
		t.Fatalf("tool result name = %q", result.ToolName)
	}
	if result.Kind != chat.KindToolResult {
		t.Fatalf("expected second item to be tool result, got kind=%v", result.Kind)
	}
	if result.IsError {
		t.Fatal("expected AGENTS.md loaded event result to be non-error")
	}
	if !strings.Contains(result.Content, "Loaded repository instructions from "+filepath.Join("nested", "AGENTS.md")) {
		t.Fatalf("tool result content = %q", result.Content)
	}
}

func TestGuidelinesLoadedEventRendersInTranscript(t *testing.T) {
	m := New(&config.Config{Model: "demo"}, nil)
	m.width = 100
	m.height = 30
	m.updateLayoutAndSize()
	m.cwd = t.TempDir()
	m.state = stateThinking

	nested := filepath.Join(m.cwd, "nested", "AGENTS.md")
	if err := os.MkdirAll(filepath.Dir(nested), 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}

	updated, _ := m.Update(agent.GuidelinesLoadedMsg{Path: nested})
	next := updated.(Model)
	rendered := next.chat.Render("")
	if !strings.Contains(rendered, "AGENTS.md") {
		t.Fatalf("expected AGENTS.md event in transcript, got:\n%s", rendered)
	}
	if !strings.Contains(rendered, "completed") {
		t.Fatalf("expected completed-style event in transcript, got:\n%s", rendered)
	}
}

func TestInternalHiddenRunDoesNotAddUserChatItem(t *testing.T) {
	m := New(&config.Config{Model: "demo"}, nil)

	updated, _ := m.Update(internalCommandRunMsg{
		prompt:      "hidden internal prompt",
		hidden:      true,
		statusLabel: "Generating AGENTS.md with AI",
	})
	next := updated.(Model)

	if len(next.chat.Items) != 0 {
		t.Fatalf("expected no user chat item for hidden run, got %d", len(next.chat.Items))
	}
	if next.state != stateThinking {
		t.Fatalf("state = %v, want thinking", next.state)
	}
	if next.commandStatus != "Generating AGENTS.md with AI" {
		t.Fatalf("commandStatus = %q", next.commandStatus)
	}
}
