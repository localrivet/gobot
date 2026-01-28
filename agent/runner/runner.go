package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"gobot/agent/ai"
	"gobot/agent/config"
	"gobot/agent/session"
	"gobot/agent/tools"
)

// DefaultSystemPrompt is the default system prompt for the agent
const DefaultSystemPrompt = `You are a helpful AI assistant with access to tools for file operations, shell commands, and more.

When working on tasks:
1. Break down complex tasks into smaller steps
2. Use tools to gather information and make changes
3. If you encounter errors, analyze them and try to fix them
4. When the task is complete, provide a summary of what was done

Important:
- Use the 'read' tool to read files instead of 'cat'
- Use the 'write' tool to create/modify files
- Use the 'glob' tool to find files by pattern
- Use the 'grep' tool to search for content in files
- Use the 'bash' tool for shell commands

Always verify your changes work before considering a task complete.`

// Runner executes the agentic loop
type Runner struct {
	sessions  *session.Manager
	providers []ai.Provider
	tools     *tools.Registry
	config    *config.Config
}

// RunRequest contains parameters for a run
type RunRequest struct {
	SessionKey string // Session identifier (uses "default" if empty)
	Prompt     string // User prompt
	System     string // Override system prompt
}

// New creates a new runner
func New(cfg *config.Config, sessions *session.Manager, providers []ai.Provider, toolRegistry *tools.Registry) *Runner {
	return &Runner{
		sessions:  sessions,
		providers: providers,
		tools:     toolRegistry,
		config:    cfg,
	}
}

// Run executes the agentic loop
func (r *Runner) Run(ctx context.Context, req *RunRequest) (<-chan ai.StreamEvent, error) {
	if len(r.providers) == 0 {
		return nil, fmt.Errorf("no providers configured")
	}

	if req.SessionKey == "" {
		req.SessionKey = "default"
	}

	// Get or create session
	sess, err := r.sessions.GetOrCreate(req.SessionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Add user message to session
	if req.Prompt != "" {
		err = r.sessions.AppendMessage(sess.ID, session.Message{
			SessionID: sess.ID,
			Role:      "user",
			Content:   req.Prompt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to save message: %w", err)
		}
	}

	resultCh := make(chan ai.StreamEvent, 100)
	go r.runLoop(ctx, sess.ID, req.System, resultCh)

	return resultCh, nil
}

// runLoop is the main agentic execution loop
func (r *Runner) runLoop(ctx context.Context, sessionID, systemPrompt string, resultCh chan<- ai.StreamEvent) {
	defer close(resultCh)

	if systemPrompt == "" {
		systemPrompt = DefaultSystemPrompt
	}

	iteration := 0
	maxIterations := r.config.MaxIterations
	if maxIterations <= 0 {
		maxIterations = 100
	}

	// OUTER LOOP: Provider fallback
	for providerIdx := 0; providerIdx < len(r.providers); providerIdx++ {
		provider := r.providers[providerIdx]
		compactionAttempted := false

		// MIDDLE LOOP: Context compaction
		for {
			// Get session messages
			messages, err := r.sessions.GetMessages(sessionID, r.config.MaxContext)
			if err != nil {
				resultCh <- ai.StreamEvent{Type: ai.EventTypeError, Error: err}
				return
			}

			// INNER LOOP: Agentic execution
			for iteration < maxIterations {
				iteration++

				// Stream to AI provider
				events, err := provider.Stream(ctx, &ai.ChatRequest{
					Messages: messages,
					Tools:    r.tools.List(),
					System:   systemPrompt,
				})

				if err != nil {
					if ai.IsContextOverflow(err) && !compactionAttempted {
						compactionAttempted = true
						// Compact session and retry
						summary := r.generateSummary(ctx, provider, messages)
						if compactErr := r.sessions.Compact(sessionID, summary); compactErr == nil {
							break // Break inner loop to retry with compacted session
						}
					}
					if ai.IsRateLimitOrAuth(err) {
						// Try next provider
						break
					}
					resultCh <- ai.StreamEvent{Type: ai.EventTypeError, Error: err}
					return
				}

				// Process streaming events
				hasToolCalls := false
				var assistantContent strings.Builder
				var toolCalls []session.ToolCall

				for event := range events {
					// Forward event to caller
					resultCh <- event

					switch event.Type {
					case ai.EventTypeText:
						assistantContent.WriteString(event.Text)

					case ai.EventTypeToolCall:
						hasToolCalls = true
						toolCalls = append(toolCalls, session.ToolCall{
							ID:    event.ToolCall.ID,
							Name:  event.ToolCall.Name,
							Input: event.ToolCall.Input,
						})

					case ai.EventTypeError:
						return
					}
				}

				// Save assistant message
				if assistantContent.Len() > 0 || len(toolCalls) > 0 {
					var toolCallsJSON json.RawMessage
					if len(toolCalls) > 0 {
						toolCallsJSON, _ = json.Marshal(toolCalls)
					}

					r.sessions.AppendMessage(sessionID, session.Message{
						SessionID: sessionID,
						Role:      "assistant",
						Content:   assistantContent.String(),
						ToolCalls: toolCallsJSON,
					})
				}

				// Execute tool calls
				if hasToolCalls {
					var toolResults []session.ToolResult

					for _, tc := range toolCalls {
						result := r.tools.Execute(ctx, &ai.ToolCall{
							ID:    tc.ID,
							Name:  tc.Name,
							Input: tc.Input,
						})

						// Send tool result event
						resultCh <- ai.StreamEvent{
							Type: ai.EventTypeToolResult,
							Text: result.Content,
						}

						toolResults = append(toolResults, session.ToolResult{
							ToolCallID: tc.ID,
							Content:    result.Content,
							IsError:    result.IsError,
						})
					}

					// Save tool results
					toolResultsJSON, _ := json.Marshal(toolResults)
					r.sessions.AppendMessage(sessionID, session.Message{
						SessionID:   sessionID,
						Role:        "tool",
						ToolResults: toolResultsJSON,
					})

					// Refresh messages for next iteration
					messages, _ = r.sessions.GetMessages(sessionID, r.config.MaxContext)
					continue
				}

				// No tool calls - LLM decided task is complete
				resultCh <- ai.StreamEvent{Type: ai.EventTypeDone}
				return
			}

			// Check if we've exhausted iterations
			if iteration >= maxIterations {
				resultCh <- ai.StreamEvent{
					Type:  ai.EventTypeError,
					Error: fmt.Errorf("reached maximum iterations (%d)", maxIterations),
				}
				return
			}

			// If we broke out to compact, continue middle loop
			if compactionAttempted {
				continue
			}

			// Otherwise, break to try next provider
			break
		}
	}

	// All providers exhausted
	resultCh <- ai.StreamEvent{
		Type:  ai.EventTypeError,
		Error: fmt.Errorf("all providers failed"),
	}
}

// generateSummary creates a summary of the conversation for compaction
func (r *Runner) generateSummary(ctx context.Context, provider ai.Provider, messages []session.Message) string {
	// Simple summary: just note that conversation was compacted
	var summary strings.Builder
	summary.WriteString("[Previous conversation summary]\n")

	// Extract key points from messages
	for _, msg := range messages {
		if msg.Role == "user" && msg.Content != "" {
			summary.WriteString("- User request: ")
			content := msg.Content
			if len(content) > 200 {
				content = content[:200] + "..."
			}
			summary.WriteString(content)
			summary.WriteString("\n")
		}
	}

	return summary.String()
}

// Chat is a convenience method for one-shot chat without tool use
func (r *Runner) Chat(ctx context.Context, prompt string) (string, error) {
	if len(r.providers) == 0 {
		return "", fmt.Errorf("no providers configured")
	}

	provider := r.providers[0]
	events, err := provider.Stream(ctx, &ai.ChatRequest{
		Messages: []session.Message{
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return "", err
	}

	var result strings.Builder
	for event := range events {
		if event.Type == ai.EventTypeText {
			result.WriteString(event.Text)
		}
		if event.Type == ai.EventTypeError {
			return result.String(), event.Error
		}
	}

	return result.String(), nil
}
