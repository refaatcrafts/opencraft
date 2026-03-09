package chat

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	xansi "github.com/charmbracelet/x/ansi"
	"github.com/muesli/reflow/wrap"
)

// Kind describes what type a chat item is.
type Kind int

const (
	KindUser Kind = iota
	KindAssistant
	KindToolCall
	KindToolResult
	KindError
)

// Item is a single entry in the chat history.
type Item struct {
	Kind      Kind
	Content   string
	ToolName  string
	Timestamp time.Time
	IsError   bool
}

// Styles are the lipgloss styles injected from the tui package (to avoid
// circular imports). They are set once at program start.
var (
	StyleUserName      lipgloss.Style
	StyleUserBorder    lipgloss.Style
	StyleTimestamp     lipgloss.Style
	StyleAssistant     lipgloss.Style
	StyleToolCallBox   lipgloss.Style
	StyleToolCallTitle lipgloss.Style
	StyleResultBox     lipgloss.Style
	StyleResultTitle   lipgloss.Style
	StyleErrorBox      lipgloss.Style
	StyleErrorTitle    lipgloss.Style
	StyleError         lipgloss.Style
	StyleEventPrefix   lipgloss.Style
	StyleEventMeta     lipgloss.Style
	StyleEventContent  lipgloss.Style
	StyleEventTruncate lipgloss.Style
)

var (
	glamourRenderer *glamour.TermRenderer
	glamourDark     = true
)

// SetGlamourDark sets the dark/light mode for glamour.
func SetGlamourDark(dark bool) {
	glamourDark = dark
}

// SetGlamourWidth rebuilds the renderer for the current content width.
func SetGlamourWidth(width int) {
	if width < 12 {
		width = 12
	}

	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("notty"),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return
	}
	glamourRenderer = renderer
}

// RenderLines converts the item to line-oriented output at the given width.
func (it Item) RenderLines(width int) []renderedLine {
	switch it.Kind {
	case KindUser:
		return renderUser(it, width)
	case KindAssistant:
		return renderAssistant(it, width)
	case KindToolCall:
		return renderToolCall(it, width)
	case KindToolResult:
		return renderToolResult(it, width)
	case KindError:
		return renderError(it, width)
	default:
		return nil
	}
}

func renderUser(it Item, width int) []renderedLine {
	contentWidth := max(8, width)
	timestamp := it.Timestamp.Format("3:04 PM")
	ts := StyleTimestamp.Render(timestamp)
	header := lipgloss.JoinHorizontal(
		lipgloss.Top,
		StyleUserName.Render("You"),
		"  ",
		ts,
	)

	bodyWidth := max(8, contentWidth-StyleUserBorder.GetHorizontalFrameSize())
	body := StyleUserBorder.Render(wrap.String(it.Content, bodyWidth))
	return append(renderBlockLines(header), renderBlockLines(body)...)
}

func renderAssistant(it Item, width int) []renderedLine {
	body := wrap.String(it.Content, max(12, width))
	if glamourRenderer != nil {
		rendered, err := glamourRenderer.Render(it.Content)
		if err == nil {
			body = strings.TrimRight(rendered, "\n")
		}
	}

	lines := renderBlockLines(StyleAssistant.Render("OpenCraft"))
	lines = append(lines, renderBlockLines(body)...)
	return lines
}

func renderToolCall(it Item, width int) []renderedLine {
	return renderToolCard(it, Item{}, width, false)
}

func renderToolResult(it Item, width int) []renderedLine {
	return renderToolCard(Item{}, it, width, false)
}

func renderError(it Item, width int) []renderedLine {
	lines := renderBlockLines(StyleErrorTitle.Render("error"))
	lines = append(lines, renderSurfaceBlock(previewText(it.Content, eventInnerWidth(width, StyleErrorBox), 6), width, StyleErrorBox)...)
	return lines
}

func renderToolPair(call Item, result Item, width int) []renderedLine {
	return renderToolCard(call, result, width, true)
}

func summarizeToolInput(input string, width int) string {
	if width < 1 {
		return ""
	}
	width = max(8, width)

	var params map[string]any
	if err := json.Unmarshal([]byte(input), &params); err != nil || len(params) == 0 {
		return StyleEventMeta.Render(truncatePlain(singleLine(input), width))
	}

	parts := make([]string, 0, len(params))
	for _, key := range orderedKeys(params) {
		parts = append(parts, fmt.Sprintf("%s=%s", key, compactValue(params[key])))
	}
	return StyleEventMeta.Render(truncatePlain(strings.Join(parts, " "), width))
}

func renderToolCard(call Item, result Item, width int, paired bool) []renderedLine {
	name := result.ToolName
	if name == "" {
		name = call.ToolName
	}

	style := StyleToolCallBox
	status := StyleEventMeta.Render("running")
	title := StyleToolCallTitle.Render(name)
	if result.ToolName != "" {
		if result.IsError {
			style = StyleErrorBox
			status = StyleErrorTitle.Render("error")
			title = StyleErrorTitle.Render(name)
		} else {
			style = StyleResultBox
			status = StyleResultTitle.Render("completed")
			title = StyleResultTitle.Render(name)
		}
	}

	innerWidth := eventInnerWidth(width, style)
	fixedWidth := plainWidth("• ") + len([]rune(name)) + plainWidth(" running")
	metaWidth := max(0, innerWidth-fixedWidth)
	meta := ""
	if call.ToolName != "" {
		meta = summarizeToolInput(call.Content, metaWidth)
	}

	headerParts := []string{
		StyleEventPrefix.Render("•"),
		title,
	}
	if meta != "" {
		headerParts = append(headerParts, meta)
	}
	headerParts = append(headerParts, status)

	lines := []string{strings.TrimSpace(strings.Join(headerParts, "  "))}
	if result.ToolName != "" {
		lines = append(lines, previewText(result.Content, eventInnerWidth(width, style), 5)...)
	}
	if paired {
		return renderSurfaceBlock(lines, width, style)
	}
	return renderSurfaceBlock(lines, width, style)
}

func renderSurfaceBlock(lines []string, width int, style lipgloss.Style) []renderedLine {
	if len(lines) == 0 {
		return nil
	}

	innerWidth := eventInnerWidth(width, style)
	contentWidth := 1
	for _, line := range lines {
		if lineWidth := plainWidth(line); lineWidth > contentWidth {
			contentWidth = lineWidth
		}
	}
	if contentWidth > innerWidth {
		contentWidth = innerWidth
	}

	block := style.Width(contentWidth).MaxWidth(contentWidth).Render(strings.Join(lines, "\n"))
	return renderBlockLines(block)
}

func eventInnerWidth(width int, style lipgloss.Style) int {
	width = min(width, 72)
	return max(12, width-style.GetHorizontalFrameSize())
}

func plainWidth(s string) int {
	return xansi.StringWidth(xansi.Strip(s))
}

func orderedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[j] < keys[i] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}

func compactValue(v any) string {
	switch vv := v.(type) {
	case string:
		if vv == "" {
			return `""`
		}
		return vv
	case []any:
		return fmt.Sprintf("[%d items]", len(vv))
	case map[string]any:
		return "{...}"
	default:
		return fmt.Sprint(v)
	}
}

func previewText(content string, width, maxLines int) []string {
	width = max(8, width)
	if maxLines < 1 {
		maxLines = 1
	}

	wrapped := wrap.String(strings.TrimSpace(content), width)
	raw := strings.Split(wrapped, "\n")
	lines := make([]string, 0, len(raw))
	for _, line := range raw {
		line = strings.TrimRight(line, " ")
		if line == "" && len(lines) == 0 {
			continue
		}
		lines = append(lines, line)
	}
	if len(lines) == 0 {
		return []string{StyleEventMeta.Render("(no output)")}
	}
	if len(lines) > maxLines {
		hidden := len(lines) - maxLines
		lines = append(lines[:maxLines], StyleEventTruncate.Render(fmt.Sprintf("... %d more lines", hidden)))
	}
	return lines
}

func singleLine(s string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(s, "\n", " ")), " ")
}

func truncatePlain(s string, width int) string {
	runes := []rune(s)
	if len(runes) <= width {
		return s
	}
	if width <= 3 {
		return string(runes[:width])
	}
	return string(runes[:width-3]) + "..."
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
