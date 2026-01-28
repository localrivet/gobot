package ai

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"gobot/agent/session"
)

// CLIProvider wraps an official CLI tool (claude, gemini, codex) as a provider
type CLIProvider struct {
	name    string
	command string
	args    []string
}

// NewCLIProvider creates a new CLI-based provider
func NewCLIProvider(name, command string, args []string) *CLIProvider {
	return &CLIProvider{
		name:    name,
		command: command,
		args:    args,
	}
}

// NewClaudeCodeProvider creates a provider that wraps the Claude Code CLI
// Claude Code: brew install claude-code or npm i -g @anthropic-ai/claude-code
// Uses ~/.claude/ for auth, supports extended thinking, MCP, agentic tools
func NewClaudeCodeProvider() *CLIProvider {
	return &CLIProvider{
		name:    "claude-code",
		command: "claude",
		args:    []string{"--print"}, // Output-only mode
	}
}

// NewGeminiCLIProvider creates a provider that wraps the Google Gemini CLI
// Gemini CLI: npm i -g @google/gemini-cli
// FREE: 1000 requests/day, 1M context window, Google Search grounding
func NewGeminiCLIProvider() *CLIProvider {
	return &CLIProvider{
		name:    "gemini-cli",
		command: "gemini",
		args:    []string{}, // Gemini CLI reads from stdin
	}
}

// NewCodexCLIProvider creates a provider that wraps the OpenAI Codex CLI
// Codex CLI: brew install --cask codex or npm i -g @openai/codex
// Uses ChatGPT account or API key, supports --full-auto mode
func NewCodexCLIProvider() *CLIProvider {
	return &CLIProvider{
		name:    "codex-cli",
		command: "codex",
		args:    []string{"--full-auto"}, // Autonomous mode
	}
}

// ID returns the provider identifier
func (p *CLIProvider) ID() string {
	return p.name
}

// Stream sends a request to the CLI and streams the response
func (p *CLIProvider) Stream(ctx context.Context, req *ChatRequest) (<-chan StreamEvent, error) {
	resultCh := make(chan StreamEvent, 100)

	go func() {
		defer close(resultCh)

		// Build the prompt from messages
		prompt := buildPromptFromMessages(req.Messages)

		// Build command args
		args := append([]string{}, p.args...)
		args = append(args, prompt)

		// Create command
		cmd := exec.CommandContext(ctx, p.command, args...)

		// Get stdout pipe for streaming
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			resultCh <- StreamEvent{
				Type:  EventTypeError,
				Error: fmt.Errorf("failed to create stdout pipe: %w", err),
			}
			return
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			resultCh <- StreamEvent{
				Type:  EventTypeError,
				Error: fmt.Errorf("failed to create stderr pipe: %w", err),
			}
			return
		}

		// Start the command
		if err := cmd.Start(); err != nil {
			resultCh <- StreamEvent{
				Type:  EventTypeError,
				Error: fmt.Errorf("failed to start %s: %w", p.command, err),
			}
			return
		}

		// Read stderr in background for error messages
		go func() {
			errBytes, _ := io.ReadAll(stderr)
			if len(errBytes) > 0 {
				errMsg := strings.TrimSpace(string(errBytes))
				if errMsg != "" && !strings.Contains(errMsg, "ANTHROPIC_API_KEY") {
					// Don't emit API key warnings as errors
					resultCh <- StreamEvent{
						Type:  EventTypeError,
						Error: fmt.Errorf("%s stderr: %s", p.command, errMsg),
					}
				}
			}
		}()

		// Stream stdout line by line
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB buffer

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				cmd.Process.Kill()
				return
			default:
				line := scanner.Text()

				// Try to parse as JSON (Claude Code may output structured data)
				event := p.parseLine(line)
				resultCh <- event
			}
		}

		if err := scanner.Err(); err != nil {
			resultCh <- StreamEvent{
				Type:  EventTypeError,
				Error: fmt.Errorf("error reading output: %w", err),
			}
		}

		// Wait for command to finish
		if err := cmd.Wait(); err != nil {
			// Don't emit error if context was cancelled
			if ctx.Err() == nil {
				resultCh <- StreamEvent{
					Type:  EventTypeError,
					Error: fmt.Errorf("%s exited with error: %w", p.command, err),
				}
			}
		}

		resultCh <- StreamEvent{Type: EventTypeDone}
	}()

	return resultCh, nil
}

// parseLine parses a line of output from the CLI
func (p *CLIProvider) parseLine(line string) StreamEvent {
	// Try to parse as JSON first (some CLIs output structured data)
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(line), &jsonData); err == nil {
		// Check for known event types
		if eventType, ok := jsonData["type"].(string); ok {
			switch eventType {
			case "tool_use", "tool_call":
				if name, ok := jsonData["name"].(string); ok {
					input, _ := json.Marshal(jsonData["input"])
					return StreamEvent{
						Type: EventTypeToolCall,
						ToolCall: &ToolCall{
							ID:    fmt.Sprintf("%v", jsonData["id"]),
							Name:  name,
							Input: input,
						},
					}
				}
			case "thinking":
				if text, ok := jsonData["text"].(string); ok {
					return StreamEvent{
						Type: EventTypeThinking,
						Text: text,
					}
				}
			case "text":
				if text, ok := jsonData["text"].(string); ok {
					return StreamEvent{
						Type: EventTypeText,
						Text: text,
					}
				}
			case "error":
				msg := fmt.Sprintf("%v", jsonData["message"])
				return StreamEvent{
					Type:  EventTypeError,
					Error: &ProviderError{Message: msg},
				}
			}
		}

		// Check for content field (common in Claude output)
		if content, ok := jsonData["content"].(string); ok {
			return StreamEvent{
				Type: EventTypeText,
				Text: content,
			}
		}
	}

	// Plain text output
	return StreamEvent{
		Type: EventTypeText,
		Text: line + "\n",
	}
}

// buildPromptFromMessages converts session messages to a single prompt string
func buildPromptFromMessages(messages []session.Message) string {
	var parts []string

	for _, msg := range messages {
		switch msg.Role {
		case "system":
			parts = append(parts, fmt.Sprintf("[System]\n%s", msg.Content))
		case "user":
			parts = append(parts, fmt.Sprintf("[User]\n%s", msg.Content))
		case "assistant":
			parts = append(parts, fmt.Sprintf("[Assistant]\n%s", msg.Content))
		case "tool":
			// Include tool results in context
			if len(msg.ToolResults) > 0 {
				var results []session.ToolResult
				json.Unmarshal(msg.ToolResults, &results)
				for _, r := range results {
					parts = append(parts, fmt.Sprintf("[Tool Result: %s]\n%s", r.ToolCallID, r.Content))
				}
			}
		}
	}

	return strings.Join(parts, "\n\n")
}

// CheckCLIAvailable checks if a CLI command is available in PATH
func CheckCLIAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// GetAvailableCLIProviders returns a list of available CLI providers
func GetAvailableCLIProviders() []string {
	var available []string

	clis := []string{"claude", "gemini", "codex"}
	for _, cli := range clis {
		if CheckCLIAvailable(cli) {
			available = append(available, cli)
		}
	}

	return available
}
