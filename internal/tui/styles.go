package tui

import "github.com/charmbracelet/lipgloss"

var (
	colorBgBase       = lipgloss.AdaptiveColor{Dark: "#11131A", Light: "#F8F6FC"}
	colorBgRaised     = lipgloss.AdaptiveColor{Dark: "#171B24", Light: "#F0EDF8"}
	colorBgPanel      = lipgloss.AdaptiveColor{Dark: "#0D1421", Light: "#FBFAFE"}
	colorEventSurface = lipgloss.AdaptiveColor{Dark: "#1F2431", Light: "#ECE8F4"}
	colorLine         = lipgloss.AdaptiveColor{Dark: "#343B4E", Light: "#D4CEE5"}
	colorText         = lipgloss.AdaptiveColor{Dark: "#F3F2F8", Light: "#17151F"}
	colorMuted        = lipgloss.AdaptiveColor{Dark: "#A1A9BC", Light: "#6F7382"}
	colorSubtle       = lipgloss.AdaptiveColor{Dark: "#677089", Light: "#AEB4C2"}
	colorAccentBlue   = lipgloss.AdaptiveColor{Dark: "#63D5FF", Light: "#005F99"}
	colorAccentIndigo = lipgloss.AdaptiveColor{Dark: "#7C6CFF", Light: "#4D41D9"}
	colorAccentPink   = lipgloss.AdaptiveColor{Dark: "#FF70C9", Light: "#B83E90"}
	colorAssistant    = lipgloss.AdaptiveColor{Dark: "#FFD479", Light: "#9A6200"}
	colorSuccess      = lipgloss.AdaptiveColor{Dark: "#68E1B8", Light: "#1F8A63"}
	colorWarning      = lipgloss.AdaptiveColor{Dark: "#FFD166", Light: "#9A5D00"}
	colorError        = lipgloss.AdaptiveColor{Dark: "#FF8B7B", Light: "#B33D2E"}
	colorSelection    = lipgloss.AdaptiveColor{Dark: "#3158A6", Light: "#D8E5FF"}
	colorSelectionFg  = lipgloss.AdaptiveColor{Dark: "#FDFDFF", Light: "#111827"}
)

var AppStyle = lipgloss.NewStyle().
	Background(colorBgBase).
	Foreground(colorText)

var HeaderStyle = lipgloss.NewStyle().
	Background(colorBgRaised).
	BorderBottom(true).
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(colorAccentIndigo).
	Padding(0, 2)

var HeaderCompactStyle = lipgloss.NewStyle().
	Background(colorBgRaised).
	BorderBottom(true).
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(colorAccentIndigo).
	Padding(0, 1)

var HeaderBrandStyle = lipgloss.NewStyle().
	Foreground(colorAccentBlue).
	Bold(true)

var HeaderModelStyle = lipgloss.NewStyle().
	Foreground(colorMuted)

var HeaderMetaStyle = lipgloss.NewStyle().
	Foreground(colorAccentPink).
	Bold(true)

var HeaderChipStyle = lipgloss.NewStyle().
	Background(colorAccentPink).
	Foreground(colorBgBase).
	Bold(true).
	Padding(0, 1)

var ChatViewportStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(colorLine).
	Padding(1, 2)

var ChatViewportCompactStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(colorLine).
	Padding(0, 1)

var WindowTooSmallStyle = lipgloss.NewStyle().
	Background(colorBgRaised).
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(colorAccentIndigo).
	Foreground(colorMuted).
	Padding(1, 2)

var EmptyStateStyle = lipgloss.NewStyle().
	Foreground(colorMuted)

var EmptyEyebrowStyle = lipgloss.NewStyle().
	Foreground(colorAccentPink).
	Bold(true)

var EmptyTitleStyle = lipgloss.NewStyle().
	Foreground(colorText).
	Bold(true)

var EmptyBodyStyle = lipgloss.NewStyle().
	Foreground(colorMuted)

var EmptyHintStyle = lipgloss.NewStyle().
	Foreground(colorSubtle)

var EmptyActionStyle = lipgloss.NewStyle().
	Foreground(colorAccentBlue)

var UserNameStyle = lipgloss.NewStyle().
	Foreground(colorAccentBlue).
	Bold(true)

var UserBorderStyle = lipgloss.NewStyle().
	BorderLeft(true).
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(colorAccentIndigo).
	PaddingLeft(1).
	Foreground(colorText)

var TimestampStyle = lipgloss.NewStyle().
	Foreground(colorSubtle)

var AssistantNameStyle = lipgloss.NewStyle().
	Foreground(colorAssistant).
	Bold(true)

var ToolCallBoxStyle = lipgloss.NewStyle().
	Background(colorEventSurface).
	Foreground(colorMuted).
	Padding(0, 1).
	MarginLeft(1)

var ToolCallTitleStyle = lipgloss.NewStyle().
	Foreground(colorWarning).
	Bold(true)

var ToolResultBoxStyle = lipgloss.NewStyle().
	Background(colorEventSurface).
	Foreground(colorText).
	Padding(0, 1).
	MarginLeft(1)

var ToolResultTitleStyle = lipgloss.NewStyle().
	Foreground(colorSuccess).
	Bold(true)

var ToolErrorBoxStyle = lipgloss.NewStyle().
	Background(colorEventSurface).
	Foreground(colorError).
	Padding(0, 1).
	MarginLeft(1)

var ToolErrorTitleStyle = lipgloss.NewStyle().
	Foreground(colorError).
	Bold(true)

var ErrorStyle = lipgloss.NewStyle().
	Foreground(colorError).
	Bold(true)

var EventPrefixStyle = lipgloss.NewStyle().
	Foreground(colorSubtle)

var EventMetaStyle = lipgloss.NewStyle().
	Foreground(colorMuted)

var EventContentStyle = lipgloss.NewStyle().
	Foreground(colorText)

var EventTruncateStyle = lipgloss.NewStyle().
	Foreground(colorSubtle).
	Italic(true)

var StatusBarStyle = lipgloss.NewStyle().
	Background(colorBgRaised).
	Foreground(colorMuted).
	Padding(0, 1)

var StatusBarActiveStyle = lipgloss.NewStyle().
	Foreground(colorText)

var StatusBarHintStyle = lipgloss.NewStyle().
	Foreground(colorSubtle)

var SpinnerStyle = lipgloss.NewStyle().
	Foreground(colorAccentIndigo)

var InputBoxStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(colorLine).
	Padding(0, 1)

var InputBoxFocusedStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(colorAccentBlue).
	Padding(0, 1)

var InputPlaceholderStyle = lipgloss.NewStyle().
	Foreground(colorSubtle)

var InputCursorStyle = lipgloss.NewStyle().
	Foreground(colorAccentPink).
	Bold(true)

var SelectionStyle = lipgloss.NewStyle().
	Background(colorSelection).
	Foreground(colorSelectionFg)

var CommandPaletteBoxStyle = lipgloss.NewStyle().
	Background(colorBgRaised).
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(colorAccentIndigo).
	Padding(1, 2)

var CommandPaletteTitleStyle = lipgloss.NewStyle().
	Foreground(colorAccentBlue).
	Bold(true)

var CommandPaletteIndexStyle = lipgloss.NewStyle().
	Foreground(colorMuted)

var CommandPaletteMutedStyle = lipgloss.NewStyle().
	Foreground(colorSubtle)

var CommandPaletteItemStyle = lipgloss.NewStyle().
	Foreground(colorText).
	Padding(0, 1)

var CommandPaletteSelectedStyle = lipgloss.NewStyle().
	Background(colorAccentIndigo).
	Foreground(colorText).
	Bold(true).
	Padding(0, 1)

var CommandPaletteShortcutStyle = lipgloss.NewStyle().
	Foreground(colorMuted)

var CommandPaletteDescStyle = lipgloss.NewStyle().
	Foreground(colorMuted)

var CommandPaletteScrollStyle = lipgloss.NewStyle().
	Foreground(colorSubtle).
	Align(lipgloss.Center)

var CommandPaletteHintStyle = lipgloss.NewStyle().
	Foreground(colorSubtle)
