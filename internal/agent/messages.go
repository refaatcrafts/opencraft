package agent

// StreamChunkMsg carries a streamed text delta from the assistant.
type StreamChunkMsg struct{ Delta string }

// GuidelinesLoadedMsg indicates repository AGENTS.md guidelines were loaded.
type GuidelinesLoadedMsg struct {
	Path string
}

// ToolCallStartMsg signals that a tool call is beginning.
type ToolCallStartMsg struct {
	ToolName string
	Input    string
}

// ToolCallResultMsg carries the result of a completed tool call.
type ToolCallResultMsg struct {
	ToolName string
	Output   string
	IsError  bool
}

// AgentDoneMsg signals the agent has finished its response.
type AgentDoneMsg struct{}

// AgentErrMsg signals an unrecoverable error in the agent loop.
type AgentErrMsg struct{ Err error }
