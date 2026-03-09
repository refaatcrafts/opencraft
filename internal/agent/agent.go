package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/genai"

	"opencraft/internal/config"
	"opencraft/internal/tools"
)

// Agent drives the conversation with Gemini and executes tools.
type Agent struct {
	client   *genai.Client
	cfg      *config.Config
	registry *tools.Registry
	history  []*genai.Content
	genTools []*genai.Tool
}

func New(cfg *config.Config, registry *tools.Registry) (*Agent, error) {
	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  cfg.GeminiAPIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("create gemini client: %w", err)
	}
	return &Agent{
		client:   client,
		cfg:      cfg,
		registry: registry,
		genTools: registry.ToGeminiTools(),
	}, nil
}

// Run executes the full agentic loop for a user message, calling send() for
// each intermediate event. Blocks until done; run in a goroutine.
func (a *Agent) Run(ctx context.Context, userMsg string, send func(any)) {
	a.history = append(a.history, &genai.Content{
		Role:  genai.RoleUser,
		Parts: []*genai.Part{{Text: userMsg}},
	})

	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{{Text: "You are opencraft, a helpful AI coding assistant. You have access to tools to read files, write files, list directories, and run bash commands. Be concise and helpful."}},
		},
		MaxOutputTokens: int32(a.cfg.MaxTokens),
		Tools:           a.genTools,
	}

	for {
		funcCalls, err := a.streamTurn(ctx, config, send)
		if err != nil {
			send(AgentErrMsg{Err: err})
			return
		}

		if len(funcCalls) == 0 {
			send(AgentDoneMsg{})
			return
		}

		var funcResponseParts []*genai.Part
		for _, fc := range funcCalls {
			inputJSON, _ := json.Marshal(fc.Args)
			send(ToolCallStartMsg{ToolName: fc.Name, Input: string(inputJSON)})

			t := a.registry.Get(fc.Name)
			var output string
			var isError bool

			if t == nil {
				output = fmt.Sprintf("unknown tool: %s", fc.Name)
				isError = true
			} else {
				var execErr error
				output, execErr = t.Execute(ctx, inputJSON)
				if execErr != nil {
					if output == "" {
						output = execErr.Error()
					} else {
						output += "\n" + execErr.Error()
					}
					isError = true
				}
			}

			send(ToolCallResultMsg{ToolName: fc.Name, Output: output, IsError: isError})

			responseMap := map[string]any{"output": output}
			if isError {
				responseMap["error"] = output
			}
			funcResponseParts = append(funcResponseParts, &genai.Part{
				FunctionResponse: &genai.FunctionResponse{
					ID:       fc.ID,
					Name:     fc.Name,
					Response: responseMap,
				},
			})
		}

		a.history = append(a.history, &genai.Content{
			Role:  genai.RoleUser,
			Parts: funcResponseParts,
		})
	}
}

// streamTurn runs one streaming API call, emits StreamChunkMsg deltas, appends
// the assistant message to history, and returns any function calls found.
func (a *Agent) streamTurn(ctx context.Context, cfg *genai.GenerateContentConfig, send func(any)) ([]*genai.FunctionCall, error) {
	var fullText strings.Builder
	var funcCalls []*genai.FunctionCall
	var modelParts []*genai.Part

	for result, err := range a.client.Models.GenerateContentStream(ctx, a.cfg.Model, a.history, cfg) {
		if err != nil {
			return nil, err
		}
		if len(result.Candidates) == 0 || result.Candidates[0].Content == nil {
			continue
		}
		for _, part := range result.Candidates[0].Content.Parts {
			if part.Text != "" && part.FunctionCall == nil && !part.Thought && len(part.ThoughtSignature) == 0 {
				send(StreamChunkMsg{Delta: part.Text})
				fullText.WriteString(part.Text)
			}
			if part.FunctionCall != nil {
				funcCalls = append(funcCalls, part.FunctionCall)
				if cloned := clonePart(part); cloned != nil {
					modelParts = append(modelParts, cloned)
				}
				continue
			}
			if part.Thought || len(part.ThoughtSignature) > 0 {
				if cloned := clonePart(part); cloned != nil {
					modelParts = append(modelParts, cloned)
				}
			}
		}
	}

	// Build assistant content for history.
	var assistantParts []*genai.Part
	if fullText.Len() > 0 {
		assistantParts = append(assistantParts, &genai.Part{Text: fullText.String()})
	}
	assistantParts = append(assistantParts, modelParts...)
	if len(assistantParts) > 0 {
		a.history = append(a.history, &genai.Content{
			Role:  genai.RoleModel,
			Parts: assistantParts,
		})
	}

	return funcCalls, nil
}

func clonePart(part *genai.Part) *genai.Part {
	if part == nil {
		return nil
	}

	data, err := json.Marshal(part)
	if err != nil {
		return nil
	}

	var cloned genai.Part
	if err := json.Unmarshal(data, &cloned); err != nil {
		return nil
	}

	return &cloned
}
