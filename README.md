# OpenCraft

Terminal-first AI coding agent built with Go, Gemini, and Charm's TUI stack.

OpenCraft is an interactive coding assistant that runs entirely in the terminal. It combines a conversational interface with tool execution so the agent can inspect files, list directories, run shell commands, and write code while keeping the full workflow visible inside a clean text UI.

## Why OpenCraft

- Keeps the coding loop inside the terminal instead of moving to a browser chat
- Makes agent actions visible through explicit tool calls and results
- Gives you a faster path for exploring a codebase, understanding files, and making edits
- Uses a native-feeling TUI for prompting, streaming output, and reviewing agent activity

## Features

- Chat-based coding workflow in a terminal UI
- Gemini-powered agent loop with streaming responses
- Built-in tools for reading files, writing files, listing directories, and running bash commands
- Visible tool activity during agent execution
- Mouse-based text selection inside the chat area
- Responsive Bubble Tea interface styled with Charm libraries

## Built With

- Go
- Google Gemini via `google.golang.org/genai`
- `bubbletea`, `bubbles`, `lipgloss`, and `glamour` from Charm

This stack fits OpenCraft well: Go keeps the runtime simple and fast, Gemini provides the model and tool-calling loop, and Charm's terminal UI libraries make it possible to build a structured, interactive coding experience without leaving the shell.

## Quick Start

Set your Gemini API key:

```bash
export GEMINI_API_KEY=your_api_key_here
```

Optional configuration:

```bash
export OPENCRAFT_MODEL=gemini-3-flash-preview
export OPENCRAFT_MAX_TOKENS=8096
```

Run the app:

```bash
go run main.go
```

## Current Tools

- `bash`
- `read_file`
- `write_file`
- `list_dir`
