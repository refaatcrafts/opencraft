package chat

import (
	"strings"
	"testing"

	xansi "github.com/charmbracelet/x/ansi"
)

func TestRenderKeepsLinesWithinWidth(t *testing.T) {
	var m Model
	m.SetSize(24, 8)
	m.AddToolResult("write_file", strings.Repeat("ABCDEFGHIJKL", 6), false)

	rendered := m.Render("")
	for _, line := range strings.Split(rendered, "\n") {
		if got := xansi.StringWidth(xansi.Strip(line)); got > 24 {
			t.Fatalf("line width %d exceeds limit: %q", got, xansi.Strip(line))
		}
	}
}

func TestToolEventsRenderAsCompactTranscript(t *testing.T) {
	var m Model
	m.SetSize(60, 12)
	m.AddToolCall("list_dir", `{"path":"."}`)
	m.AddToolResult("list_dir", "a\nb\nc\nd\ne\nf\ng\nh\ni\nj", false)

	rendered := xansi.Strip(m.Render(""))
	if strings.Contains(rendered, "tool list_dir\n") {
		t.Fatal("expected compact tool rendering, got old boxed header")
	}
	if !strings.Contains(rendered, "list_dir") {
		t.Fatal("expected tool name in render")
	}
	if !strings.Contains(rendered, "path=.") {
		t.Fatal("expected summarized tool args in render")
	}
	if !strings.Contains(rendered, "... 5 more lines") {
		t.Fatal("expected truncated tool output hint")
	}
	if strings.Contains(rendered, "path=.\n\n") {
		t.Fatal("expected tool card to stay grouped")
	}
}

func TestSelectionReturnsPlainTextAcrossLines(t *testing.T) {
	m := Model{
		lines: []renderedLine{
			{Plain: "alpha beta", Styled: "alpha beta"},
			{Plain: "gamma delta", Styled: "gamma delta"},
		},
	}

	if !m.StartSelection(strings.Index(m.lines[0].Plain, "beta"), 0) {
		t.Fatal("expected selection to start")
	}
	if !m.UpdateSelection(strings.Index(m.lines[1].Plain, "gam")+len("gam"), 1) {
		t.Fatal("expected selection to update")
	}

	got := m.FinishSelection()
	want := "beta\ngam"
	if got != want {
		t.Fatalf("selected text = %q, want %q", got, want)
	}
}

func TestClearSelectionResetsState(t *testing.T) {
	m := Model{
		lines: []renderedLine{
			{Plain: "one two", Styled: "one two"},
		},
	}

	m.StartSelection(0, 0)
	m.UpdateSelection(3, 0)
	if !m.HasSelection() {
		t.Fatal("expected active selection")
	}
	if !m.ClearSelection() {
		t.Fatal("expected clear selection to report change")
	}
	if m.HasSelection() {
		t.Fatal("selection should be cleared")
	}
}
