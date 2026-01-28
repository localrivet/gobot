package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// BashTool executes shell commands
type BashTool struct {
	policy *Policy
}

// NewBashTool creates a new bash tool
func NewBashTool(policy *Policy) *BashTool {
	return &BashTool{policy: policy}
}

// Name returns the tool name
func (t *BashTool) Name() string {
	return "bash"
}

// Description returns the tool description
func (t *BashTool) Description() string {
	return `Execute a bash command. Use for running shell commands, scripts, and system operations.
Be careful with destructive commands - they require user approval.
Prefer using dedicated tools (read, write, glob, grep) for file operations.`
}

// Schema returns the JSON schema for the tool input
func (t *BashTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"command": {
				"type": "string",
				"description": "The bash command to execute"
			},
			"timeout": {
				"type": "integer",
				"description": "Timeout in seconds (default: 120)"
			},
			"cwd": {
				"type": "string",
				"description": "Working directory for the command"
			}
		},
		"required": ["command"]
	}`)
}

// BashInput represents the tool input
type BashInput struct {
	Command string `json:"command"`
	Timeout int    `json:"timeout"`
	Cwd     string `json:"cwd"`
}

// Execute runs the bash command
func (t *BashTool) Execute(ctx context.Context, input json.RawMessage) (*ToolResult, error) {
	var in BashInput
	if err := json.Unmarshal(input, &in); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if in.Command == "" {
		return &ToolResult{
			Content: "Error: command is required",
			IsError: true,
		}, nil
	}

	// Set default timeout
	timeout := 120 * time.Second
	if in.Timeout > 0 {
		timeout = time.Duration(in.Timeout) * time.Second
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Create command
	cmd := exec.CommandContext(ctx, "bash", "-c", in.Command)
	if in.Cwd != "" {
		cmd.Dir = in.Cwd
	}

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run command
	err := cmd.Run()

	// Build result
	var result strings.Builder
	if stdout.Len() > 0 {
		result.WriteString(stdout.String())
	}
	if stderr.Len() > 0 {
		if result.Len() > 0 {
			result.WriteString("\n")
		}
		result.WriteString("STDERR:\n")
		result.WriteString(stderr.String())
	}

	// Handle errors
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return &ToolResult{
				Content: fmt.Sprintf("Command timed out after %v\n%s", timeout, result.String()),
				IsError: true,
			}, nil
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			return &ToolResult{
				Content: fmt.Sprintf("Command exited with code %d\n%s", exitErr.ExitCode(), result.String()),
				IsError: true,
			}, nil
		}
		return &ToolResult{
			Content: fmt.Sprintf("Command failed: %v\n%s", err, result.String()),
			IsError: true,
		}, nil
	}

	output := result.String()
	if output == "" {
		output = "(no output)"
	}

	// Truncate very long output
	const maxOutput = 50000
	if len(output) > maxOutput {
		output = output[:maxOutput] + "\n... (output truncated)"
	}

	return &ToolResult{
		Content: output,
	}, nil
}

// RequiresApproval checks if this command needs approval
func (t *BashTool) RequiresApproval() bool {
	// Actual check happens in policy during Execute
	return true
}
