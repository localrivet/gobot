// Clipboard plugin for macOS.
// Provides: get, set, clear, history (if supported), watch
// Build with: go build -o clipboard
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/rpc"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-plugin"
)

var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "GOBOT_PLUGIN",
	MagicCookieValue: "gobot-plugin-v1",
}

type ToolResult struct {
	Content string `json:"content"`
	IsError bool   `json:"is_error"`
}

type ClipboardTool struct {
	mu      sync.Mutex
	history []ClipboardEntry
	maxHist int
}

type ClipboardEntry struct {
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"` // text, image, file
}

func NewClipboardTool() *ClipboardTool {
	return &ClipboardTool{
		history: make([]ClipboardEntry, 0),
		maxHist: 20,
	}
}

func (t *ClipboardTool) Name() string {
	return "clipboard"
}

func (t *ClipboardTool) Description() string {
	return "Manage macOS clipboard: get current content, set new content, clear, view history, and get clipboard type (text/image/file)."
}

func (t *ClipboardTool) Schema() json.RawMessage {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"action": map[string]any{
				"type":        "string",
				"description": "Action to perform: get, set, clear, type, history",
				"enum":        []string{"get", "set", "clear", "type", "history"},
			},
			"content": map[string]any{
				"type":        "string",
				"description": "Content to set (for set action)",
			},
			"limit": map[string]any{
				"type":        "integer",
				"description": "Number of history entries to return (for history action, default: 10)",
			},
		},
		"required": []string{"action"},
	}
	data, _ := json.Marshal(schema)
	return data
}

type ClipboardInput struct {
	Action  string `json:"action"`
	Content string `json:"content"`
	Limit   int    `json:"limit"`
}

func (t *ClipboardTool) Execute(ctx context.Context, input json.RawMessage) (*ToolResult, error) {
	var in ClipboardInput
	if err := json.Unmarshal(input, &in); err != nil {
		return &ToolResult{
			Content: fmt.Sprintf("Failed to parse input: %v", err),
			IsError: true,
		}, nil
	}

	var result string
	var err error

	switch in.Action {
	case "get":
		result, err = t.getClipboard()
	case "set":
		result, err = t.setClipboard(in.Content)
	case "clear":
		result, err = t.clearClipboard()
	case "type":
		result, err = t.getClipboardType()
	case "history":
		limit := in.Limit
		if limit <= 0 {
			limit = 10
		}
		result, err = t.getHistory(limit)
	default:
		return &ToolResult{
			Content: fmt.Sprintf("Unknown action: %s", in.Action),
			IsError: true,
		}, nil
	}

	if err != nil {
		return &ToolResult{
			Content: fmt.Sprintf("Action failed: %v", err),
			IsError: true,
		}, nil
	}

	return &ToolResult{
		Content: result,
		IsError: false,
	}, nil
}

func (t *ClipboardTool) getClipboard() (string, error) {
	out, err := exec.Command("pbpaste").Output()
	if err != nil {
		return "", fmt.Errorf("pbpaste failed: %v", err)
	}

	content := string(out)

	// Store in history
	t.mu.Lock()
	t.history = append([]ClipboardEntry{{
		Content:   content,
		Timestamp: time.Now(),
		Type:      "text",
	}}, t.history...)
	if len(t.history) > t.maxHist {
		t.history = t.history[:t.maxHist]
	}
	t.mu.Unlock()

	if content == "" {
		return "Clipboard is empty (or contains non-text content)", nil
	}

	// Truncate if very long
	if len(content) > 5000 {
		return fmt.Sprintf("%s\n\n... (truncated, total %d characters)", content[:5000], len(content)), nil
	}

	return content, nil
}

func (t *ClipboardTool) setClipboard(content string) (string, error) {
	if content == "" {
		return "", fmt.Errorf("content is required")
	}

	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(content)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("pbcopy failed: %v", err)
	}

	// Store in history
	t.mu.Lock()
	t.history = append([]ClipboardEntry{{
		Content:   content,
		Timestamp: time.Now(),
		Type:      "text",
	}}, t.history...)
	if len(t.history) > t.maxHist {
		t.history = t.history[:t.maxHist]
	}
	t.mu.Unlock()

	preview := content
	if len(preview) > 100 {
		preview = preview[:100] + "..."
	}
	return fmt.Sprintf("Clipboard set to: %q", preview), nil
}

func (t *ClipboardTool) clearClipboard() (string, error) {
	// Set clipboard to empty string
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader("")
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to clear clipboard: %v", err)
	}
	return "Clipboard cleared", nil
}

func (t *ClipboardTool) getClipboardType() (string, error) {
	// Use AppleScript to check clipboard type
	script := `
		set clipTypes to ""
		try
			set clipInfo to (clipboard info)
			repeat with clipItem in clipInfo
				set clipTypes to clipTypes & (item 1 of clipItem as string) & ", "
			end repeat
		end try
		return clipTypes
	`

	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return "", fmt.Errorf("failed to get clipboard type: %v", err)
	}

	result := strings.TrimSpace(string(out))
	if result == "" {
		return "Clipboard is empty", nil
	}

	// Parse common types
	types := strings.TrimSuffix(result, ", ")
	if strings.Contains(types, "«class PNGf»") || strings.Contains(types, "TIFF") {
		return fmt.Sprintf("Clipboard contains: IMAGE\nRaw types: %s", types), nil
	}
	if strings.Contains(types, "furl") || strings.Contains(types, "«class furl»") {
		return fmt.Sprintf("Clipboard contains: FILE(S)\nRaw types: %s", types), nil
	}
	if strings.Contains(types, "utxt") || strings.Contains(types, "«class utf8»") {
		return fmt.Sprintf("Clipboard contains: TEXT\nRaw types: %s", types), nil
	}

	return fmt.Sprintf("Clipboard types: %s", types), nil
}

func (t *ClipboardTool) getHistory(limit int) (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if len(t.history) == 0 {
		return "No clipboard history available (history is session-only and starts empty)", nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Clipboard History (last %d entries):\n\n", min(limit, len(t.history))))

	count := min(limit, len(t.history))
	for i := 0; i < count; i++ {
		entry := t.history[i]
		preview := entry.Content
		if len(preview) > 80 {
			preview = preview[:80] + "..."
		}
		// Replace newlines for cleaner display
		preview = strings.ReplaceAll(preview, "\n", "\\n")
		sb.WriteString(fmt.Sprintf("%d. [%s] %s: %q\n",
			i+1,
			entry.Timestamp.Format("15:04:05"),
			entry.Type,
			preview,
		))
	}

	return sb.String(), nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (t *ClipboardTool) RequiresApproval() bool {
	return false // Clipboard access is relatively safe
}

// RPC Server implementation
type ToolRPCServer struct {
	Impl *ClipboardTool
}

func (s *ToolRPCServer) Name(_ struct{}, resp *string) error {
	*resp = s.Impl.Name()
	return nil
}

func (s *ToolRPCServer) Description(_ struct{}, resp *string) error {
	*resp = s.Impl.Description()
	return nil
}

func (s *ToolRPCServer) Schema(_ struct{}, resp *json.RawMessage) error {
	*resp = s.Impl.Schema()
	return nil
}

type ExecuteArgs struct {
	Input json.RawMessage
}

type ExecuteReply struct {
	Result *ToolResult
	Error  string
}

func (s *ToolRPCServer) Execute(args ExecuteArgs, reply *ExecuteReply) error {
	result, err := s.Impl.Execute(context.Background(), args.Input)
	reply.Result = result
	if err != nil {
		reply.Error = err.Error()
	}
	return nil
}

func (s *ToolRPCServer) RequiresApproval(_ struct{}, resp *bool) error {
	*resp = s.Impl.RequiresApproval()
	return nil
}

type ToolPlugin struct {
	Impl *ClipboardTool
}

func (p *ToolPlugin) Server(*plugin.MuxBroker) (any, error) {
	return &ToolRPCServer{Impl: p.Impl}, nil
}

func (p *ToolPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (any, error) {
	return nil, fmt.Errorf("client not implemented")
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: Handshake,
		Plugins: map[string]plugin.Plugin{
			"tool": &ToolPlugin{Impl: NewClipboardTool()},
		},
	})
}
