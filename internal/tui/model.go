package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	xansi "github.com/charmbracelet/x/ansi"
	"github.com/muesli/reflow/wrap"

	"opencraft/internal/agent"
	"opencraft/internal/config"
	"opencraft/internal/tui/chat"
)

type appState int

const (
	stateIdle appState = iota
	stateThinking
	stateStreaming
	stateToolRunning
)

const (
	compactWidthBreakpoint  = 96
	compactHeightBreakpoint = 28
	minWidth                = 44
	minHeight               = 12
	mouseWheelDelta         = 3
)

const (
	commandInitAGENTS = "init_agents"
	commandClearChat  = "clear_chat"
	commandClearInput = "clear_input"
	commandCopySelect = "copy_selection"
)

type paletteCommand struct {
	id          string
	label       string
	description string
	shortcut    string
}

type layoutState struct {
	marginX        int
	panelWidth     int
	headerY        int
	headerHeight   int
	chatY          int
	chatHeight     int
	statusY        int
	statusHeight   int
	inputY         int
	inputHeight    int
	textareaHeight int
	compact        bool
	tooSmall       bool
}

// Model is the top-level bubbletea model.
type Model struct {
	cfg       *config.Config
	ag        *agent.Agent
	ctx       context.Context
	cancel    context.CancelFunc
	agentChan chan tea.Msg
	cwd       string

	state     appState
	chat      chat.Model
	streamBuf string
	viewport  viewport.Model
	textarea  textarea.Model
	spinner   spinner.Model

	activeToolName string
	width          int
	height         int
	ready          bool
	layout         layoutState
	statusMessage  string
	commandStatus  string
	guidelinesPath string
	paletteOpen    bool
	paletteIndex   int
	paletteScroll  int
	paletteItems   []paletteCommand
}

// New creates a fully initialised TUI model.
func New(cfg *config.Config, ag *agent.Agent) Model {
	ctx, cancel := context.WithCancel(context.Background())
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	sp.Style = SpinnerStyle

	ta := textarea.New()
	ta.Placeholder = "Ask OpenCraft to inspect, edit, or run code"
	ta.Prompt = ""
	ta.ShowLineNumbers = false
	ta.CharLimit = 0
	ta.SetHeight(3)
	ta.FocusedStyle.Base = lipgloss.NewStyle()
	ta.BlurredStyle.Base = lipgloss.NewStyle()
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.BlurredStyle.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(colorSubtle)
	ta.BlurredStyle.Placeholder = lipgloss.NewStyle().Foreground(colorSubtle)
	ta.FocusedStyle.Prompt = lipgloss.NewStyle().Foreground(colorAccentBlue)
	ta.BlurredStyle.Prompt = lipgloss.NewStyle().Foreground(colorMuted)
	ta.Focus()

	chatModel := chat.New()

	chat.StyleUserName = UserNameStyle
	chat.StyleUserBorder = UserBorderStyle
	chat.StyleTimestamp = TimestampStyle
	chat.StyleAssistant = AssistantNameStyle
	chat.StyleToolCallBox = ToolCallBoxStyle
	chat.StyleToolCallTitle = ToolCallTitleStyle
	chat.StyleResultBox = ToolResultBoxStyle
	chat.StyleResultTitle = ToolResultTitleStyle
	chat.StyleErrorBox = ToolErrorBoxStyle
	chat.StyleErrorTitle = ToolErrorTitleStyle
	chat.StyleError = ErrorStyle
	chat.StyleEventPrefix = EventPrefixStyle
	chat.StyleEventMeta = EventMetaStyle
	chat.StyleEventContent = EventContentStyle
	chat.StyleEventTruncate = EventTruncateStyle
	chat.StyleSelection = SelectionStyle

	vp := viewport.New(0, 0)
	vp.MouseWheelEnabled = false

	return Model{
		cfg:       cfg,
		ag:        ag,
		ctx:       ctx,
		cancel:    cancel,
		agentChan: make(chan tea.Msg, 64),
		cwd:       cwd,
		spinner:   sp,
		textarea:  ta,
		chat:      chatModel,
		viewport:  vp,
		paletteItems: []paletteCommand{
			{
				id:          commandInitAGENTS,
				label:       "Initialize AGENTS.md",
				description: "Use AI to generate a project-specific AGENTS.md from repository context.",
				shortcut:    "enter",
			},
			{
				id:          commandClearChat,
				label:       "Clear Chat History",
				description: "Remove all chat messages from the current session view.",
				shortcut:    "enter",
			},
			{
				id:          commandClearInput,
				label:       "Clear Input",
				description: "Reset the message composer.",
				shortcut:    "enter",
			},
			{
				id:          commandCopySelect,
				label:       "Copy Selection",
				description: "Copy currently selected chat text to clipboard.",
				shortcut:    "enter",
			},
		},
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		m.spinner.Tick,
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayoutAndSize()
		m.ready = true
		m.refreshViewport(false)

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			m.cancel()
			return m, tea.Quit
		}

		if m.paletteOpen {
			handled, cmd := m.handlePaletteKey(msg)
			if handled {
				m.refreshViewport(false)
			}
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			break
		}

		if m.state == stateIdle && msg.Type == tea.KeyRunes && string(msg.Runes) == "/" && m.textarea.Value() == "" {
			m.openPalette()
			break
		}

		switch msg.Type {
		case tea.KeyEsc:
			if m.chat.ClearSelection() {
				m.setStatusMessage("selection cleared")
				m.refreshViewport(false)
			}
		case tea.KeyPgUp:
			m.viewport.ScrollUp(m.viewport.Height)
		case tea.KeyPgDown:
			m.viewport.ScrollDown(m.viewport.Height)
		case tea.KeyHome:
			m.viewport.GotoTop()
		case tea.KeyEnd:
			m.viewport.GotoBottom()
		case tea.KeyEnter:
			if msg.Alt {
				var taCmd tea.Cmd
				m.textarea, taCmd = m.textarea.Update(msg)
				cmds = append(cmds, taCmd)
				break
			}
			if m.state == stateIdle {
				text := strings.TrimSpace(m.textarea.Value())
				if text != "" {
					m.chat.ClearSelection()
					m.statusMessage = ""
					m.textarea.Reset()
					cmds = append(cmds, func() tea.Msg { return userSubmitMsg{text: text} })
				}
			}
		default:
			var taCmd tea.Cmd
			m.textarea, taCmd = m.textarea.Update(msg)
			cmds = append(cmds, taCmd)
		}

	case tea.MouseMsg:
		if m.paletteOpen || !m.ready || m.layout.tooSmall {
			break
		}
		if m.handleMouse(msg) {
			m.refreshViewport(false)
		}

	case userSubmitMsg:
		cmds = append(cmds, m.beginAgentRun(internalCommandRunMsg{
			prompt: msg.text,
			hidden: false,
		})...)

	case internalCommandRunMsg:
		cmds = append(cmds, m.beginAgentRun(msg)...)

	case agent.GuidelinesLoadedMsg:
		m.guidelinesPath = m.guidelineDisplayPath(msg.Path)
		meta := map[string]string{
			"event": "loaded",
			"path":  m.guidelinesPath,
		}
		metaJSON, _ := json.Marshal(meta)
		m.chat.AddToolCall("AGENTS.md", string(metaJSON))
		m.chat.AddToolResult("AGENTS.md", "Loaded repository instructions from "+m.guidelinesPath, false)
		m.refreshViewport(false)
		cmds = append(cmds, m.waitForAgent())

	case agent.StreamChunkMsg:
		if m.state != stateStreaming {
			m.state = stateStreaming
			m.activeToolName = ""
		}
		m.streamBuf += msg.Delta
		m.refreshViewport(true)
		cmds = append(cmds, m.waitForAgent())

	case agent.ToolCallStartMsg:
		if buf := m.streamBuf; buf != "" {
			m.chat.AddAssistant(buf)
			m.streamBuf = ""
		}
		m.chat.AddToolCall(msg.ToolName, msg.Input)
		m.state = stateToolRunning
		m.activeToolName = msg.ToolName
		m.refreshViewport(true)
		cmds = append(cmds, m.waitForAgent())

	case agent.ToolCallResultMsg:
		m.chat.AddToolResult(msg.ToolName, msg.Output, msg.IsError)
		m.state = stateThinking
		m.activeToolName = ""
		m.refreshViewport(true)
		cmds = append(cmds, m.waitForAgent())

	case agent.AgentDoneMsg:
		if buf := m.streamBuf; buf != "" {
			m.chat.AddAssistant(buf)
			m.streamBuf = ""
		}
		m.state = stateIdle
		m.activeToolName = ""
		m.commandStatus = ""
		m.refreshViewport(true)

	case agent.AgentErrMsg:
		if buf := m.streamBuf; buf != "" {
			m.chat.AddAssistant(buf)
			m.streamBuf = ""
		}
		m.chat.AddError(msg.Err.Error())
		m.state = stateIdle
		m.activeToolName = ""
		m.commandStatus = ""
		m.refreshViewport(true)

	case spinner.TickMsg:
		if m.state != stateIdle {
			var spCmd tea.Cmd
			m.spinner, spCmd = m.spinner.Update(msg)
			cmds = append(cmds, spCmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if !m.ready {
		return "\n  Loading OpenCraft..."
	}
	if m.layout.tooSmall {
		return m.renderTooSmall()
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		m.renderHeader(),
		m.renderChatPanel(),
		m.renderStatusBar(),
		m.renderInput(),
	)
	content = lipgloss.NewStyle().Margin(0, m.layout.marginX).Render(content)
	root := lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Left,
		lipgloss.Top,
		content,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceBackground(AppStyle.GetBackground()),
	)
	if m.paletteOpen {
		return m.renderPaletteOverlay(root)
	}
	return root
}

func (m *Model) updateLayoutAndSize() {
	layout := layoutState{
		marginX:      0,
		panelWidth:   m.width,
		headerY:      0,
		statusHeight: 1,
		compact:      m.width < compactWidthBreakpoint || m.height < compactHeightBreakpoint,
		tooSmall:     m.width < minWidth || m.height < minHeight,
	}
	if m.width >= 72 {
		layout.marginX = 1
		layout.panelWidth = m.width - 2
	}

	headerStyle := HeaderStyle
	if layout.compact {
		headerStyle = HeaderCompactStyle
	}
	layout.headerHeight = headerStyle.GetVerticalFrameSize() + 1

	layout.textareaHeight = 3
	if !layout.compact && m.height >= 34 {
		layout.textareaHeight = 4
	}
	inputStyle := InputBoxFocusedStyle
	layout.inputHeight = inputStyle.GetVerticalFrameSize() + layout.textareaHeight

	layout.chatY = layout.headerHeight
	layout.chatHeight = m.height - layout.headerHeight - layout.statusHeight - layout.inputHeight
	layout.statusY = layout.chatY + layout.chatHeight
	layout.inputY = layout.statusY + layout.statusHeight
	if layout.chatHeight < 4 {
		layout.tooSmall = true
	}
	m.layout = layout

	if layout.tooSmall {
		return
	}

	chatStyle := m.chatPanelStyle()
	chatWidth := layout.panelWidth - chatStyle.GetHorizontalFrameSize()
	chatHeight := layout.chatHeight - chatStyle.GetVerticalFrameSize()
	m.viewport.Width = chatWidth
	m.viewport.Height = chatHeight
	m.viewport.Style = lipgloss.NewStyle()
	m.chat.SetSize(chatWidth, chatHeight)

	inputStyle = InputBoxStyle
	if m.state == stateIdle {
		inputStyle = InputBoxFocusedStyle
	}
	textareaWidth := layout.panelWidth - inputStyle.GetHorizontalFrameSize()
	if textareaWidth < 8 {
		textareaWidth = 8
	}
	m.textarea.SetWidth(textareaWidth)
	m.textarea.SetHeight(layout.textareaHeight)
}

func (m *Model) refreshViewport(anchorBottom bool) {
	if m.layout.tooSmall {
		return
	}

	wasBottom := anchorBottom || m.viewport.AtBottom()
	prevOffset := m.viewport.YOffset
	m.viewport.SetContent(m.chat.Render(m.streamBuf))
	if wasBottom {
		m.viewport.GotoBottom()
		return
	}
	m.viewport.SetYOffset(prevOffset)
}

func (m Model) renderHeader() string {
	style := m.headerStyle()
	innerWidth := m.layout.panelWidth - style.GetHorizontalFrameSize()

	brand := HeaderBrandStyle.Render("OpenCraft")
	chip := HeaderChipStyle.Render("code agent")
	left := lipgloss.JoinHorizontal(lipgloss.Center, brand, " ", chip)

	meta := m.headerStateLabel()
	modelWidth := max(8, innerWidth/2)
	model := ansiTruncateStyled(HeaderModelStyle.Render(m.cfg.Model), modelWidth)
	right := lipgloss.JoinHorizontal(lipgloss.Center, model, "  ", HeaderMetaStyle.Render(meta))

	rightWidth := lipgloss.Width(xansi.Strip(right))
	leftWidth := max(0, innerWidth-rightWidth-1)
	row := lipgloss.JoinHorizontal(
		lipgloss.Center,
		lipgloss.NewStyle().Width(leftWidth).MaxWidth(leftWidth).Render(left),
		lipgloss.NewStyle().MaxWidth(innerWidth-leftWidth).Render(right),
	)
	return style.Render(fillArea(innerWidth, 1, row, style.GetBackground(), lipgloss.Left, lipgloss.Center))
}

func (m Model) renderChatPanel() string {
	style := m.chatPanelStyle()
	innerWidth := m.layout.panelWidth - style.GetHorizontalFrameSize()
	innerHeight := m.layout.chatHeight - style.GetVerticalFrameSize()

	var content string
	if len(m.chat.Items) == 0 && m.streamBuf == "" {
		content = m.renderEmptyState(innerWidth, innerHeight)
	} else {
		content = placeArea(innerWidth, innerHeight, m.viewport.View(), lipgloss.Left, lipgloss.Top)
	}

	return style.Render(content)
}

func (m Model) renderStatusBar() string {
	innerWidth := m.layout.panelWidth - StatusBarStyle.GetHorizontalFrameSize()
	leftText := m.statusSummary()
	rightText := "/ commands  enter send  alt+enter newline  drag select  esc clear"
	if m.layout.compact {
		rightText = "/ commands  enter send  drag select"
	}

	left := StatusBarActiveStyle.Render(leftText)
	right := StatusBarHintStyle.Render(rightText)

	rightWidth := lipgloss.Width(xansi.Strip(rightText))
	leftWidth := max(0, innerWidth-rightWidth-1)
	if lipgloss.Width(xansi.Strip(leftText)) > leftWidth {
		left = StatusBarActiveStyle.Render(xansi.Truncate(leftText, leftWidth, "..."))
	}

	row := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Width(leftWidth).MaxWidth(leftWidth).Render(left),
		lipgloss.NewStyle().MaxWidth(innerWidth-leftWidth).Render(right),
	)
	return StatusBarStyle.Render(fillArea(innerWidth, 1, row, StatusBarStyle.GetBackground(), lipgloss.Left, lipgloss.Center))
}

func (m Model) renderInput() string {
	style := m.inputStyle()
	innerWidth := m.layout.panelWidth - style.GetHorizontalFrameSize()
	innerHeight := m.layout.inputHeight - style.GetVerticalFrameSize()
	content := m.textarea.View()
	if m.textarea.Value() == "" {
		content = InputCursorStyle.Render("│") + " " + InputPlaceholderStyle.Render(m.textarea.Placeholder)
	}
	return style.Render(placeArea(innerWidth, innerHeight, content, lipgloss.Left, lipgloss.Top))
}

func (m Model) renderTooSmall() string {
	body := fmt.Sprintf(
		"OpenCraft needs at least %dx%d cells.\nCurrent terminal: %dx%d\n\nResize the terminal to restore the full chat UI.",
		minWidth,
		minHeight,
		m.width,
		m.height,
	)

	contentWidth := max(20, m.width-WindowTooSmallStyle.GetHorizontalFrameSize()-2)
	box := WindowTooSmallStyle.Width(contentWidth).Render(body)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m Model) statusSummary() string {
	if m.statusMessage != "" {
		return m.statusMessage
	}

	suffix := m.guidelinesStatusSuffix()
	if m.commandStatus != "" && m.state != stateIdle {
		return m.spinner.View() + " " + m.commandStatus + suffix
	}

	switch m.state {
	case stateThinking:
		return m.spinner.View() + " Thinking" + suffix
	case stateStreaming:
		return m.spinner.View() + " Writing response" + suffix
	case stateToolRunning:
		tool := m.activeToolName
		if tool == "" {
			tool = "tool"
		}
		return m.spinner.View() + " Running " + tool + suffix
	default:
		if m.chat.HasSelection() {
			return "Selection active"
		}
		return "Ready"
	}
}

func (m Model) headerStateLabel() string {
	switch m.state {
	case stateThinking:
		return "thinking"
	case stateStreaming:
		return "writing"
	case stateToolRunning:
		return "tool"
	default:
		if m.layout.compact {
			return "compact"
		}
		return "ready"
	}
}

func (m Model) renderEmptyState(width, height int) string {
	cardWidth := min(max(24, width-12), 56)
	bodyWidth := max(18, cardWidth)
	lines := []string{
		EmptyBodyStyle.Render(wrap.String("Ask OpenCraft to inspect code, explain a file, or make a change.", bodyWidth)),
		"",
		EmptyHintStyle.Render("Try one of these"),
		EmptyActionStyle.Render("summarize this repository"),
		EmptyActionStyle.Render("explain the main entrypoint"),
		EmptyActionStyle.Render("trace the tool call flow"),
	}
	if m.layout.compact {
		lines = []string{
			EmptyHintStyle.Render("Try one of these"),
			EmptyActionStyle.Render("summarize this repository"),
			EmptyActionStyle.Render("trace the tool call flow"),
		}
	}
	return placeArea(width, height, strings.Join(lines, "\n"), lipgloss.Center, lipgloss.Center)
}

func (m *Model) handleMouse(msg tea.MouseMsg) bool {
	if m.isInChatContent(msg.X, msg.Y) {
		contentX, contentY := m.chatContentPoint(msg.X, msg.Y)
		contentY += m.viewport.YOffset

		switch {
		case msg.Button == tea.MouseButtonWheelUp && msg.Action == tea.MouseActionPress:
			m.viewport.ScrollUp(mouseWheelDelta)
			m.statusMessage = ""
			return true
		case msg.Button == tea.MouseButtonWheelDown && msg.Action == tea.MouseActionPress:
			m.viewport.ScrollDown(mouseWheelDelta)
			m.statusMessage = ""
			return true
		case msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress:
			m.statusMessage = ""
			if m.chat.StartSelection(contentX, contentY) {
				return true
			}
		case msg.Action == tea.MouseActionMotion && m.chat.UpdateSelection(contentX, contentY):
			return true
		case msg.Action == tea.MouseActionRelease:
			selected := m.chat.FinishSelection()
			if selected == "" {
				return false
			}
			m.copySelection(selected)
			return true
		}
		return false
	}

	if m.chat.ClearSelection() {
		m.setStatusMessage("selection cleared")
		return true
	}
	return false
}

func (m Model) isInChatContent(x, y int) bool {
	contentX, contentY := m.chatContentOrigin()
	if x < contentX || y < contentY {
		return false
	}
	if x >= contentX+m.chatContentWidth() {
		return false
	}
	if y >= contentY+m.chatContentHeight() {
		return false
	}
	return true
}

func (m Model) chatContentOrigin() (int, int) {
	style := m.chatPanelStyle()
	x := m.layout.marginX + style.GetBorderLeftSize() + style.GetPaddingLeft()
	y := m.layout.chatY + style.GetBorderTopSize() + style.GetPaddingTop()
	return x, y
}

func (m Model) chatContentPoint(x, y int) (int, int) {
	originX, originY := m.chatContentOrigin()
	return x - originX, y - originY
}

func (m Model) chatContentWidth() int {
	return max(1, m.layout.panelWidth-m.chatPanelStyle().GetHorizontalFrameSize())
}

func (m Model) chatContentHeight() int {
	return max(1, m.layout.chatHeight-m.chatPanelStyle().GetVerticalFrameSize())
}

func (m *Model) copySelection(selected string) {
	if err := clipboard.WriteAll(selected); err != nil {
		m.setStatusMessage(fmt.Sprintf("selected %d chars (clipboard unavailable)", len([]rune(selected))))
		return
	}
	m.setStatusMessage(fmt.Sprintf("copied %d chars", len([]rune(selected))))
}

func (m *Model) setStatusMessage(message string) {
	m.statusMessage = message
}

func (m *Model) beginAgentRun(req internalCommandRunMsg) []tea.Cmd {
	if strings.TrimSpace(req.prompt) == "" {
		return nil
	}

	if !req.hidden {
		m.chat.AddUser(req.prompt)
	}
	m.chat.ClearSelection()
	m.statusMessage = ""
	m.commandStatus = req.statusLabel
	m.state = stateThinking
	m.activeToolName = ""
	m.streamBuf = ""
	m.guidelinesPath = ""
	m.refreshViewport(true)

	return []tea.Cmd{
		m.startAgent(req.prompt),
		m.waitForAgent(),
		m.spinner.Tick,
	}
}

func (m Model) guidelinesStatusSuffix() string {
	if m.guidelinesPath == "" {
		return ""
	}
	return " • AGENTS.md loaded (" + m.guidelinesPath + ")"
}

func (m Model) guidelineDisplayPath(path string) string {
	rel, err := filepath.Rel(m.cwd, path)
	if err != nil || rel == "" {
		return path
	}
	if strings.HasPrefix(rel, "..") {
		return path
	}
	return rel
}

func (m *Model) openPalette() {
	if len(m.paletteItems) == 0 || m.state != stateIdle {
		return
	}
	m.paletteOpen = true
	m.paletteIndex = 0
	m.paletteScroll = 0
	m.statusMessage = ""
}

func (m *Model) closePalette() {
	m.paletteOpen = false
}

func (m *Model) handlePaletteKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.closePalette()
		return true, nil
	case tea.KeyEnter:
		cmd := m.runSelectedPaletteCommand()
		m.closePalette()
		return true, cmd
	case tea.KeyUp:
		m.movePaletteSelection(-1)
		return true, nil
	case tea.KeyDown:
		m.movePaletteSelection(1)
		return true, nil
	case tea.KeyPgUp:
		m.movePaletteSelection(-m.paletteVisibleCount())
		return true, nil
	case tea.KeyPgDown:
		m.movePaletteSelection(m.paletteVisibleCount())
		return true, nil
	case tea.KeyHome:
		m.paletteIndex = 0
		m.ensurePaletteSelectionVisible()
		return true, nil
	case tea.KeyEnd:
		if n := len(m.paletteItems); n > 0 {
			m.paletteIndex = n - 1
			m.ensurePaletteSelectionVisible()
		}
		return true, nil
	}

	if msg.Type != tea.KeyRunes {
		return false, nil
	}

	switch string(msg.Runes) {
	case "j":
		m.movePaletteSelection(1)
		return true, nil
	case "k":
		m.movePaletteSelection(-1)
		return true, nil
	}
	return false, nil
}

func (m *Model) movePaletteSelection(delta int) {
	n := len(m.paletteItems)
	if n == 0 {
		return
	}
	m.paletteIndex += delta
	if m.paletteIndex < 0 {
		m.paletteIndex = n - 1
	}
	if m.paletteIndex >= n {
		m.paletteIndex = 0
	}
	m.ensurePaletteSelectionVisible()
}

func (m *Model) ensurePaletteSelectionVisible() {
	visible := m.paletteVisibleCount()
	if m.paletteIndex < m.paletteScroll {
		m.paletteScroll = m.paletteIndex
	}
	if m.paletteIndex >= m.paletteScroll+visible {
		m.paletteScroll = m.paletteIndex - visible + 1
	}
	if m.paletteScroll < 0 {
		m.paletteScroll = 0
	}
}

func (m Model) paletteVisibleCount() int {
	visible := m.height - 18
	if m.layout.compact {
		visible = m.height - 14
	}
	if visible < 4 {
		visible = 4
	}
	if visible > 12 {
		visible = 12
	}
	return visible
}

func (m *Model) runSelectedPaletteCommand() tea.Cmd {
	if len(m.paletteItems) == 0 || m.paletteIndex < 0 || m.paletteIndex >= len(m.paletteItems) {
		return nil
	}

	switch m.paletteItems[m.paletteIndex].id {
	case commandInitAGENTS:
		return func() tea.Msg {
			return internalCommandRunMsg{
				prompt:      m.buildAGENTSInitPrompt(),
				hidden:      true,
				statusLabel: "Generating AGENTS.md with AI",
			}
		}
	case commandClearChat:
		m.chat = chat.New()
		m.chat.SetSize(m.viewport.Width, m.viewport.Height)
		m.streamBuf = ""
		m.commandStatus = ""
		m.setStatusMessage("chat history cleared")
		m.refreshViewport(true)
	case commandClearInput:
		m.textarea.Reset()
		m.setStatusMessage("input cleared")
	case commandCopySelect:
		selected := m.chat.SelectedText()
		if selected == "" {
			m.setStatusMessage("no active selection to copy")
		} else {
			m.copySelection(selected)
		}
	}
	return nil
}

func (m Model) buildAGENTSInitPrompt() string {
	projectName := filepath.Base(m.cwd)
	if projectName == "" || projectName == "." || projectName == string(filepath.Separator) {
		projectName = "project"
	}

	return strings.TrimSpace(fmt.Sprintf(`
Initialize AGENTS.md for this repository.

Requirements:
1) Inspect repository context before writing by reading README.md and a few key project files/directories.
2) Write AGENTS.md at exactly this path: AGENTS.md.
3) Overwrite AGENTS.md if it already exists. Do not create backups.
4) File format must be:
   - First line: # AGENTS.md instructions for %s
   - Then a single <INSTRUCTIONS>...</INSTRUCTIONS> block.
5) Inside <INSTRUCTIONS>, include practical, repo-specific guidance:
   - Role and responsibilities for the coding agent
   - Workflow expectations (planning, edits, verification)
   - Code quality/style rules aligned with this repo
   - Testing/validation expectations with concrete commands when known
   - Tool usage and safety boundaries relevant to this codebase
6) Keep instructions concise, actionable, and specific to this repository (not generic boilerplate).
7) Use tools to complete the work, and finish after AGENTS.md is written.
`, projectName))
}

func (m Model) renderPaletteOverlay(_ string) string {
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		m.renderCommandPalette(),
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceBackground(AppStyle.GetBackground()),
	)
}

func (m Model) renderCommandPalette() string {
	if len(m.paletteItems) == 0 {
		return CommandPaletteBoxStyle.Render("No commands")
	}

	width := min(max(54, m.layout.panelWidth-14), 88)
	if m.layout.compact {
		width = m.layout.panelWidth - 2
	}
	if width < 40 {
		width = 40
	}

	boxStyle := CommandPaletteBoxStyle.Width(width)
	innerWidth := width - boxStyle.GetHorizontalFrameSize()
	visible := min(len(m.paletteItems), m.paletteVisibleCount())
	if visible < 1 {
		visible = 1
	}

	if maxScroll := max(0, len(m.paletteItems)-visible); m.paletteScroll > maxScroll {
		m.paletteScroll = maxScroll
	}

	start := m.paletteScroll
	end := min(len(m.paletteItems), start+visible)

	var rows []string
	if start > 0 {
		rows = append(rows, CommandPaletteScrollStyle.Width(innerWidth).Render("..."))
	}
	for i := start; i < end; i++ {
		rows = append(rows, m.renderPaletteCommandRow(innerWidth, m.paletteItems[i], i == m.paletteIndex))
	}
	if end < len(m.paletteItems) {
		rows = append(rows, CommandPaletteScrollStyle.Width(innerWidth).Render("..."))
	}

	selected := m.paletteItems[m.paletteIndex]
	index := fmt.Sprintf("%d/%d", m.paletteIndex+1, len(m.paletteItems))
	title := lipgloss.JoinHorizontal(
		lipgloss.Center,
		CommandPaletteTitleStyle.Render("Commands"),
		"  ",
		CommandPaletteIndexStyle.Render(index),
	)

	body := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		CommandPaletteMutedStyle.Render("Use up/down to choose, enter to run, esc to close."),
		"",
		strings.Join(rows, "\n"),
		"",
		CommandPaletteDescStyle.Render(wrap.String(selected.description, innerWidth)),
		"",
		CommandPaletteHintStyle.Render("↑/↓ choose  pgup/pgdn scroll  enter confirm  esc cancel"),
	)
	return boxStyle.Render(body)
}

func (m Model) renderPaletteCommandRow(width int, item paletteCommand, selected bool) string {
	shortcut := CommandPaletteShortcutStyle.Render(item.shortcut)
	shortcutWidth := lipgloss.Width(xansi.Strip(item.shortcut))
	if shortcutWidth < 0 {
		shortcutWidth = 0
	}

	leftWidth := max(1, width-shortcutWidth-2)
	label := xansi.Truncate(item.label, leftWidth, "...")

	row := lipgloss.JoinHorizontal(
		lipgloss.Center,
		lipgloss.NewStyle().Width(leftWidth).Render(label),
		"  ",
		lipgloss.NewStyle().Width(shortcutWidth).Align(lipgloss.Right).Render(shortcut),
	)

	if selected {
		return CommandPaletteSelectedStyle.Width(width).Render(row)
	}
	return CommandPaletteItemStyle.Width(width).Render(row)
}

func (m Model) startAgent(text string) tea.Cmd {
	return func() tea.Msg {
		go m.ag.Run(m.ctx, text, func(msg any) {
			m.agentChan <- msg.(tea.Msg)
		})
		return nil
	}
}

func (m Model) waitForAgent() tea.Cmd {
	return func() tea.Msg {
		return <-m.agentChan
	}
}

func ansiTruncateStyled(text string, width int) string {
	if width <= 0 {
		return ""
	}
	return xansi.Truncate(text, width, "...")
}

func (m Model) headerStyle() lipgloss.Style {
	if m.layout.compact {
		return HeaderCompactStyle
	}
	return HeaderStyle
}

func (m Model) chatPanelStyle() lipgloss.Style {
	if m.layout.compact {
		return ChatViewportCompactStyle
	}
	return ChatViewportStyle
}

func (m Model) inputStyle() lipgloss.Style {
	if m.state == stateIdle {
		return InputBoxFocusedStyle
	}
	return InputBoxStyle
}

func fillArea(width, height int, content string, bg lipgloss.TerminalColor, hPos, vPos lipgloss.Position) string {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	return lipgloss.Place(
		width,
		height,
		hPos,
		vPos,
		content,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceBackground(bg),
	)
}

func placeArea(width, height int, content string, hPos, vPos lipgloss.Position) string {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	return lipgloss.Place(width, height, hPos, vPos, content)
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
