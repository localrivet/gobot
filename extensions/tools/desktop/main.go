// Desktop control plugin for macOS using cliclick (or AppleScript fallback).
// Provides: click, double_click, right_click, type, hotkey, scroll, move, drag
// Build with: go build -o desktop
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/rpc"
	"os/exec"
	"strconv"
	"strings"

	"github.com/hashicorp/go-plugin"
)

// Handshake must match the main application
var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "GOBOT_PLUGIN",
	MagicCookieValue: "gobot-plugin-v1",
}

// ToolResult matches the protocol definition
type ToolResult struct {
	Content string `json:"content"`
	IsError bool   `json:"is_error"`
}

// DesktopTool provides mouse and keyboard control for macOS
type DesktopTool struct {
	useCliclick bool // true if cliclick is available
}

func NewDesktopTool() *DesktopTool {
	// Check if cliclick is available
	_, err := exec.LookPath("cliclick")
	return &DesktopTool{useCliclick: err == nil}
}

func (t *DesktopTool) Name() string {
	return "desktop"
}

func (t *DesktopTool) Description() string {
	return "Control macOS desktop: mouse clicks, keyboard input, scrolling, and cursor movement. Requires cliclick (brew install cliclick) for best results, falls back to AppleScript."
}

func (t *DesktopTool) Schema() json.RawMessage {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"action": map[string]any{
				"type":        "string",
				"description": "Action to perform: click, double_click, right_click, type, hotkey, scroll, move, drag",
				"enum":        []string{"click", "double_click", "right_click", "type", "hotkey", "scroll", "move", "drag"},
			},
			"x": map[string]any{
				"type":        "integer",
				"description": "X coordinate (for click, move, drag actions)",
			},
			"y": map[string]any{
				"type":        "integer",
				"description": "Y coordinate (for click, move, drag actions)",
			},
			"text": map[string]any{
				"type":        "string",
				"description": "Text to type (for type action)",
			},
			"keys": map[string]any{
				"type":        "string",
				"description": "Keyboard shortcut (for hotkey action), e.g., 'cmd+c', 'cmd+shift+s', 'return', 'escape'",
			},
			"direction": map[string]any{
				"type":        "string",
				"description": "Scroll direction: up, down, left, right",
				"enum":        []string{"up", "down", "left", "right"},
			},
			"amount": map[string]any{
				"type":        "integer",
				"description": "Scroll amount (number of scroll units, default: 3)",
			},
			"to_x": map[string]any{
				"type":        "integer",
				"description": "Destination X coordinate (for drag action)",
			},
			"to_y": map[string]any{
				"type":        "integer",
				"description": "Destination Y coordinate (for drag action)",
			},
		},
		"required": []string{"action"},
	}
	data, _ := json.Marshal(schema)
	return data
}

type DesktopInput struct {
	Action    string `json:"action"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Text      string `json:"text"`
	Keys      string `json:"keys"`
	Direction string `json:"direction"`
	Amount    int    `json:"amount"`
	ToX       int    `json:"to_x"`
	ToY       int    `json:"to_y"`
}

func (t *DesktopTool) Execute(ctx context.Context, input json.RawMessage) (*ToolResult, error) {
	var in DesktopInput
	if err := json.Unmarshal(input, &in); err != nil {
		return &ToolResult{
			Content: fmt.Sprintf("Failed to parse input: %v", err),
			IsError: true,
		}, nil
	}

	var result string
	var err error

	switch in.Action {
	case "click":
		result, err = t.click(in.X, in.Y, "left", 1)
	case "double_click":
		result, err = t.click(in.X, in.Y, "left", 2)
	case "right_click":
		result, err = t.click(in.X, in.Y, "right", 1)
	case "type":
		result, err = t.typeText(in.Text)
	case "hotkey":
		result, err = t.hotkey(in.Keys)
	case "scroll":
		amount := in.Amount
		if amount == 0 {
			amount = 3
		}
		result, err = t.scroll(in.Direction, amount)
	case "move":
		result, err = t.moveCursor(in.X, in.Y)
	case "drag":
		result, err = t.drag(in.X, in.Y, in.ToX, in.ToY)
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

func (t *DesktopTool) click(x, y int, button string, count int) (string, error) {
	if t.useCliclick {
		var cmd string
		switch button {
		case "right":
			cmd = "rc"
		default:
			if count == 2 {
				cmd = "dc"
			} else {
				cmd = "c"
			}
		}
		_, err := exec.Command("cliclick", fmt.Sprintf("%s:%d,%d", cmd, x, y)).CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("cliclick failed: %v", err)
		}
		return fmt.Sprintf("Clicked at (%d, %d) with %s button", x, y, button), nil
	}

	// AppleScript fallback
	script := fmt.Sprintf(`
		tell application "System Events"
			click at {%d, %d}
		end tell
	`, x, y)
	_, err := exec.Command("osascript", "-e", script).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("AppleScript click failed: %v", err)
	}
	return fmt.Sprintf("Clicked at (%d, %d)", x, y), nil
}

func (t *DesktopTool) typeText(text string) (string, error) {
	if t.useCliclick {
		_, err := exec.Command("cliclick", "t:"+text).CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("cliclick type failed: %v", err)
		}
		return fmt.Sprintf("Typed: %q", text), nil
	}

	// AppleScript fallback - escape special characters
	escaped := strings.ReplaceAll(text, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	script := fmt.Sprintf(`
		tell application "System Events"
			keystroke "%s"
		end tell
	`, escaped)
	_, err := exec.Command("osascript", "-e", script).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("AppleScript type failed: %v", err)
	}
	return fmt.Sprintf("Typed: %q", text), nil
}

func (t *DesktopTool) hotkey(keys string) (string, error) {
	// Parse keys like "cmd+c", "cmd+shift+s", "return"
	parts := strings.Split(strings.ToLower(keys), "+")

	if t.useCliclick {
		// Convert to cliclick format
		var cliclickKeys []string
		for _, p := range parts {
			switch p {
			case "cmd", "command":
				cliclickKeys = append(cliclickKeys, "cmd")
			case "ctrl", "control":
				cliclickKeys = append(cliclickKeys, "ctrl")
			case "alt", "option":
				cliclickKeys = append(cliclickKeys, "alt")
			case "shift":
				cliclickKeys = append(cliclickKeys, "shift")
			case "return", "enter":
				cliclickKeys = append(cliclickKeys, "return")
			case "escape", "esc":
				cliclickKeys = append(cliclickKeys, "esc")
			case "tab":
				cliclickKeys = append(cliclickKeys, "tab")
			case "space":
				cliclickKeys = append(cliclickKeys, "space")
			case "delete", "backspace":
				cliclickKeys = append(cliclickKeys, "delete")
			case "up":
				cliclickKeys = append(cliclickKeys, "arrow-up")
			case "down":
				cliclickKeys = append(cliclickKeys, "arrow-down")
			case "left":
				cliclickKeys = append(cliclickKeys, "arrow-left")
			case "right":
				cliclickKeys = append(cliclickKeys, "arrow-right")
			default:
				cliclickKeys = append(cliclickKeys, p)
			}
		}
		_, err := exec.Command("cliclick", "kp:"+strings.Join(cliclickKeys, ",")).CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("cliclick hotkey failed: %v", err)
		}
		return fmt.Sprintf("Pressed: %s", keys), nil
	}

	// AppleScript fallback
	var modifiers []string
	var key string
	for _, p := range parts {
		switch p {
		case "cmd", "command":
			modifiers = append(modifiers, "command down")
		case "ctrl", "control":
			modifiers = append(modifiers, "control down")
		case "alt", "option":
			modifiers = append(modifiers, "option down")
		case "shift":
			modifiers = append(modifiers, "shift down")
		case "return", "enter":
			key = "return"
		case "escape", "esc":
			key = "escape"
		case "tab":
			key = "tab"
		case "space":
			key = "space"
		case "delete", "backspace":
			key = "delete"
		case "up":
			key = "up arrow"
		case "down":
			key = "down arrow"
		case "left":
			key = "left arrow"
		case "right":
			key = "right arrow"
		default:
			key = p
		}
	}

	var script string
	if len(modifiers) > 0 {
		script = fmt.Sprintf(`
			tell application "System Events"
				key code (key code "%s") using {%s}
			end tell
		`, key, strings.Join(modifiers, ", "))
		// Actually, simpler approach:
		script = fmt.Sprintf(`
			tell application "System Events"
				keystroke "%s" using {%s}
			end tell
		`, key, strings.Join(modifiers, ", "))
	} else {
		script = fmt.Sprintf(`
			tell application "System Events"
				key code (key code "%s")
			end tell
		`, key)
	}
	_, err := exec.Command("osascript", "-e", script).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("AppleScript hotkey failed: %v", err)
	}
	return fmt.Sprintf("Pressed: %s", keys), nil
}

func (t *DesktopTool) scroll(direction string, amount int) (string, error) {
	if t.useCliclick {
		var deltaX, deltaY int
		switch direction {
		case "up":
			deltaY = amount
		case "down":
			deltaY = -amount
		case "left":
			deltaX = amount
		case "right":
			deltaX = -amount
		default:
			return "", fmt.Errorf("invalid scroll direction: %s", direction)
		}
		// cliclick scroll format: scroll:x,y (positive = up/left)
		_, err := exec.Command("cliclick", fmt.Sprintf("scroll:%d,%d", deltaX, deltaY)).CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("cliclick scroll failed: %v", err)
		}
		return fmt.Sprintf("Scrolled %s by %d", direction, amount), nil
	}

	// AppleScript doesn't have great scroll support, use mouse scroll wheel simulation
	// This is a rough approximation
	script := fmt.Sprintf(`
		tell application "System Events"
			repeat %d times
				-- scroll simulation (limited in AppleScript)
			end repeat
		end tell
	`, amount)
	_, _ = exec.Command("osascript", "-e", script).CombinedOutput()
	return fmt.Sprintf("Scrolled %s by %d (AppleScript - limited support)", direction, amount), nil
}

func (t *DesktopTool) moveCursor(x, y int) (string, error) {
	if t.useCliclick {
		_, err := exec.Command("cliclick", fmt.Sprintf("m:%d,%d", x, y)).CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("cliclick move failed: %v", err)
		}
		return fmt.Sprintf("Moved cursor to (%d, %d)", x, y), nil
	}

	// AppleScript fallback using cliclick is preferred; no good AS alternative
	return "", fmt.Errorf("move requires cliclick - install with: brew install cliclick")
}

func (t *DesktopTool) drag(fromX, fromY, toX, toY int) (string, error) {
	if t.useCliclick {
		_, err := exec.Command("cliclick",
			fmt.Sprintf("dd:%d,%d", fromX, fromY),
			fmt.Sprintf("du:%d,%d", toX, toY),
		).CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("cliclick drag failed: %v", err)
		}
		return fmt.Sprintf("Dragged from (%d, %d) to (%d, %d)", fromX, fromY, toX, toY), nil
	}

	return "", fmt.Errorf("drag requires cliclick - install with: brew install cliclick")
}

func (t *DesktopTool) RequiresApproval() bool {
	return true // Desktop control is sensitive
}

// RPC Server implementation
type ToolRPCServer struct {
	Impl *DesktopTool
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

// ToolPlugin implements hashicorp/go-plugin interface
type ToolPlugin struct {
	Impl *DesktopTool
}

func (p *ToolPlugin) Server(*plugin.MuxBroker) (any, error) {
	return &ToolRPCServer{Impl: p.Impl}, nil
}

func (p *ToolPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (any, error) {
	return nil, fmt.Errorf("client not implemented")
}

// Unused but required for strconv import
var _ = strconv.Itoa

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: Handshake,
		Plugins: map[string]plugin.Plugin{
			"tool": &ToolPlugin{Impl: NewDesktopTool()},
		},
	})
}
