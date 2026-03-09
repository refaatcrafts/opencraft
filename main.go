package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"opencraft/internal/agent"
	"opencraft/internal/config"
	"opencraft/internal/tools"
	"opencraft/internal/tui"
	"opencraft/internal/tui/chat"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "opencraft:", err)
		os.Exit(1)
	}

	// Detect terminal background colour BEFORE p.Run() while the terminal is
	// still in normal (cooked) mode. This avoids the glamour renderer blocking
	// inside the TUI event loop on the OSC-11 colour query.
	chat.SetGlamourDark(lipgloss.HasDarkBackground())

	registry := tools.NewRegistry()
	registry.Register(tools.NewBashTool())
	registry.Register(tools.NewReadFileTool())
	registry.Register(tools.NewWriteFileTool())
	registry.Register(tools.NewListDirTool())

	ag, err := agent.New(cfg, registry)
	if err != nil {
		fmt.Fprintln(os.Stderr, "opencraft:", err)
		os.Exit(1)
	}
	model := tui.New(cfg, ag)

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "opencraft:", err)
		os.Exit(1)
	}
}
