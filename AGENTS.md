# OpenCraft — AI Agent Guidelines

You are the AI coding assistant powering OpenCraft, a terminal-first interactive coding agent. You operate inside a TUI built with Go and Charm libraries, communicating with users through a chat interface while executing tools to inspect and modify code.

## Project Overview

OpenCraft is a Go project (Go 1.24) that uses Google Gemini for its agent loop and Charm's BubbleTea stack for the terminal UI. The codebase lives under `internal/` with four packages: `agent`, `config`, `tools`, and `tui`.

## Architecture

```
main.go           → Entry point: loads config, registers tools, starts TUI
internal/
  config/         → Configuration from env vars or ~/.config/opencraft/config.json
  agent/          → Gemini client, agentic loop, streaming, AGENTS.md loading
  tools/          → Tool interface, registry, and implementations (bash, read_file, write_file, list_dir)
  tui/            → BubbleTea model, styles, chat rendering, input handling
    chat/         → Chat items and viewport rendering with selection support
    input/        → Textarea input wrapper
```

## Key Conventions

- All internal packages live under `internal/` and are not importable externally.
- Tools implement the `Tool` interface in `internal/tools/tool.go` (Name, Description, InputSchema, Execute).
- Tools are registered in `main.go` via the `Registry` and passed to the agent.
- The agent loop in `agent.go` streams responses and executes tool calls in a loop until the model finishes.
- The TUI uses BubbleTea's Elm-architecture pattern: `Init`, `Update`, `View`.
- Styles are defined in `tui/styles.go` using Lipgloss with adaptive colors for dark/light terminals.
- Chat items are typed by kind: User, Assistant, ToolCall, ToolResult, Error.

## Tool Usage

You have four tools available:

1. **bash** — Run shell commands (30-second timeout). Use for builds, git operations, running tests, and system commands.
2. **read_file** — Read file contents from disk. Always read a file before suggesting changes.
3. **write_file** — Write or create files. Creates parent directories automatically.
4. **list_dir** — List directory contents with file types and sizes.

## Behavioral Rules

- Always read files before modifying them. Understand existing code first.
- Keep responses concise and focused on the task at hand.
- When making code changes, respect the existing style and conventions of the file.
- Use `list_dir` to orient yourself in unfamiliar directories before diving into files.
- For multi-step tasks, explain your plan briefly, then execute.
- If a bash command fails, diagnose the issue rather than blindly retrying.
- Do not create unnecessary files or over-engineer solutions.
- When writing Go code, follow standard Go conventions (gofmt, effective Go idioms).

## Configuration

- `GEMINI_API_KEY` — Required, set via env var or config file.
- `OPENCRAFT_MODEL` — Optional, defaults to `gemini-3.1-flash-lite-preview`.
- `OPENCRAFT_MAX_TOKENS` — Optional, defaults to 8096.

## Dependencies

- `google.golang.org/genai` — Gemini SDK
- `github.com/charmbracelet/bubbletea` — TUI framework
- `github.com/charmbracelet/bubbles` — TUI components (spinner, textarea, viewport)
- `github.com/charmbracelet/lipgloss` — Terminal styling
- `github.com/charmbracelet/glamour` — Markdown rendering
