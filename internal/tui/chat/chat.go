package chat

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	xansi "github.com/charmbracelet/x/ansi"
)

type position struct {
	Line   int
	Column int
}

type renderedLine struct {
	Styled string
	Plain  string
}

// Model holds the ordered list of chat items and their rendered form.
type Model struct {
	Items  []Item
	width  int
	height int

	lines []renderedLine

	selecting bool
	anchor    position
	cursor    position
}

func New() Model {
	return Model{}
}

// SetSize updates the rendering size for the chat area.
func (m *Model) SetSize(width, height int) {
	if width < 12 {
		width = 12
	}
	if height < 1 {
		height = 1
	}
	m.width = width
	m.height = height
	SetGlamourWidth(width)
}

// SetWidth keeps compatibility with older call sites.
func (m *Model) SetWidth(width int) {
	m.SetSize(width, m.height)
}

// AddUser appends a user message.
func (m *Model) AddUser(content string) {
	m.Items = append(m.Items, Item{
		Kind:      KindUser,
		Content:   content,
		Timestamp: time.Now(),
	})
}

// AddAssistant appends a completed assistant message.
func (m *Model) AddAssistant(content string) {
	m.Items = append(m.Items, Item{
		Kind:    KindAssistant,
		Content: content,
	})
}

// AddToolCall appends a tool call item.
func (m *Model) AddToolCall(toolName, input string) {
	m.Items = append(m.Items, Item{
		Kind:     KindToolCall,
		ToolName: toolName,
		Content:  input,
	})
}

// AddToolResult appends a tool result item.
func (m *Model) AddToolResult(toolName, output string, isError bool) {
	m.Items = append(m.Items, Item{
		Kind:     KindToolResult,
		ToolName: toolName,
		Content:  output,
		IsError:  isError,
	})
}

// AddError appends an error item.
func (m *Model) AddError(msg string) {
	m.Items = append(m.Items, Item{
		Kind:    KindError,
		Content: msg,
	})
}

// Render returns the rendered chat content and refreshes the internal line map.
func (m *Model) Render(streamBuf string) string {
	m.lines = m.renderLines(streamBuf)
	if len(m.lines) == 0 {
		return ""
	}

	out := make([]string, len(m.lines))
	for i, line := range m.lines {
		if rng, ok := m.selectionRangeForLine(i); ok {
			out[i] = highlightPlainLine(line.Plain, rng.start, rng.end)
			continue
		}
		out[i] = line.Styled
	}

	return strings.Join(out, "\n")
}

// ClearSelection clears any active or finished selection.
func (m *Model) ClearSelection() bool {
	changed := m.selecting || m.hasSelection()
	m.selecting = false
	m.anchor = position{}
	m.cursor = position{}
	return changed
}

// StartSelection begins a mouse-driven selection.
func (m *Model) StartSelection(x, y int) bool {
	pos, ok := m.clampPosition(x, y)
	if !ok {
		return false
	}
	m.selecting = true
	m.anchor = pos
	m.cursor = pos
	return true
}

// UpdateSelection updates the current selection while dragging.
func (m *Model) UpdateSelection(x, y int) bool {
	if !m.selecting {
		return false
	}
	pos, ok := m.clampPosition(x, y)
	if !ok {
		return false
	}
	if pos == m.cursor {
		return false
	}
	m.cursor = pos
	return true
}

// FinishSelection ends dragging and returns the selected content.
func (m *Model) FinishSelection() string {
	m.selecting = false
	if !m.hasSelection() {
		return ""
	}
	return m.SelectedText()
}

// HasSelection reports whether a non-empty selection exists.
func (m *Model) HasSelection() bool {
	return m.hasSelection()
}

// SelectedText returns the selected text without styling.
func (m *Model) SelectedText() string {
	if !m.hasSelection() {
		return ""
	}

	start, end := normalizeRange(m.anchor, m.cursor)
	out := make([]string, 0, end.Line-start.Line+1)
	for lineIdx := start.Line; lineIdx <= end.Line; lineIdx++ {
		runes := []rune(m.lines[lineIdx].Plain)
		lineStart := 0
		lineEnd := len(runes)
		if lineIdx == start.Line {
			lineStart = clamp(start.Column, 0, len(runes))
		}
		if lineIdx == end.Line {
			lineEnd = clamp(end.Column, 0, len(runes))
		}
		if lineStart > lineEnd {
			lineStart, lineEnd = lineEnd, lineStart
		}
		out = append(out, string(runes[lineStart:lineEnd]))
	}
	return strings.Join(out, "\n")
}

func (m *Model) renderLines(streamBuf string) []renderedLine {
	items := make([]Item, 0, len(m.Items)+1)
	items = append(items, m.Items...)
	if streamBuf != "" {
		items = append(items, Item{Kind: KindAssistant, Content: streamBuf})
	}

	lines := make([]renderedLine, 0, len(items)*4)
	for i := 0; i < len(items); i++ {
		item := items[i]
		if i > 0 && len(lines) > 0 && needsSeparator(items[i-1], item) {
			lines = append(lines, renderedLine{})
		}
		if item.Kind == KindToolCall && i+1 < len(items) && items[i+1].Kind == KindToolResult && items[i+1].ToolName == item.ToolName {
			lines = append(lines, renderToolPair(item, items[i+1], m.width)...)
			i++
			continue
		}
		lines = append(lines, item.RenderLines(m.width)...)
	}
	return lines
}

func needsSeparator(prev, next Item) bool {
	return !(isEventItem(prev.Kind) && isEventItem(next.Kind))
}

func isEventItem(kind Kind) bool {
	switch kind {
	case KindToolCall, KindToolResult, KindError:
		return true
	default:
		return false
	}
}

func (m *Model) clampPosition(x, y int) (position, bool) {
	if len(m.lines) == 0 || y < 0 {
		return position{}, false
	}
	if y >= len(m.lines) {
		y = len(m.lines) - 1
	}

	runes := []rune(m.lines[y].Plain)
	if x < 0 {
		x = 0
	}
	if x > len(runes) {
		x = len(runes)
	}

	return position{Line: y, Column: x}, true
}

func (m *Model) hasSelection() bool {
	return m.anchor != m.cursor
}

type lineSelection struct {
	start int
	end   int
}

func (m *Model) selectionRangeForLine(line int) (lineSelection, bool) {
	if !m.hasSelection() || line < 0 || line >= len(m.lines) {
		return lineSelection{}, false
	}

	start, end := normalizeRange(m.anchor, m.cursor)
	if line < start.Line || line > end.Line {
		return lineSelection{}, false
	}

	runes := []rune(m.lines[line].Plain)
	lineStart := 0
	lineEnd := len(runes)
	if line == start.Line {
		lineStart = clamp(start.Column, 0, len(runes))
	}
	if line == end.Line {
		lineEnd = clamp(end.Column, 0, len(runes))
	}
	if lineStart == lineEnd {
		return lineSelection{}, false
	}
	return lineSelection{start: lineStart, end: lineEnd}, true
}

func highlightPlainLine(text string, start, end int) string {
	runes := []rune(text)
	start = clamp(start, 0, len(runes))
	end = clamp(end, 0, len(runes))
	if start >= end {
		return text
	}

	var b strings.Builder
	b.WriteString(string(runes[:start]))
	b.WriteString(StyleSelection.Render(string(runes[start:end])))
	b.WriteString(string(runes[end:]))
	return b.String()
}

func normalizeRange(a, b position) (position, position) {
	if a.Line > b.Line || (a.Line == b.Line && a.Column > b.Column) {
		return b, a
	}
	return a, b
}

func clamp(v, low, high int) int {
	if v < low {
		return low
	}
	if v > high {
		return high
	}
	return v
}

func renderBlockLines(styled string) []renderedLine {
	rawLines := strings.Split(strings.TrimRight(styled, "\n"), "\n")
	if len(rawLines) == 1 && rawLines[0] == "" {
		return nil
	}

	lines := make([]renderedLine, 0, len(rawLines))
	for _, line := range rawLines {
		lines = append(lines, renderedLine{
			Styled: line,
			Plain:  xansi.Strip(line),
		})
	}
	return lines
}

var StyleSelection lipgloss.Style
