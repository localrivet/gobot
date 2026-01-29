// Window management plugin for macOS.
// Provides: list, focus, move, resize, minimize, maximize, close
// Build with: go build -o window
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

var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "GOBOT_PLUGIN",
	MagicCookieValue: "gobot-plugin-v1",
}

type ToolResult struct {
	Content string `json:"content"`
	IsError bool   `json:"is_error"`
}

type WindowTool struct{}

func (t *WindowTool) Name() string {
	return "window"
}

func (t *WindowTool) Description() string {
	return "Manage macOS windows: list all windows, focus/bring to front, move, resize, minimize, maximize, or close windows by app name or window title."
}

func (t *WindowTool) Schema() json.RawMessage {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"action": map[string]any{
				"type":        "string",
				"description": "Action to perform: list, focus, move, resize, minimize, maximize, close",
				"enum":        []string{"list", "focus", "move", "resize", "minimize", "maximize", "close"},
			},
			"app": map[string]any{
				"type":        "string",
				"description": "Application name (e.g., 'Google Chrome', 'Finder', 'Terminal')",
			},
			"title": map[string]any{
				"type":        "string",
				"description": "Window title to match (partial match supported)",
			},
			"x": map[string]any{
				"type":        "integer",
				"description": "X position for move action",
			},
			"y": map[string]any{
				"type":        "integer",
				"description": "Y position for move action",
			},
			"width": map[string]any{
				"type":        "integer",
				"description": "Width for resize action",
			},
			"height": map[string]any{
				"type":        "integer",
				"description": "Height for resize action",
			},
		},
		"required": []string{"action"},
	}
	data, _ := json.Marshal(schema)
	return data
}

type WindowInput struct {
	Action string `json:"action"`
	App    string `json:"app"`
	Title  string `json:"title"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type WindowInfo struct {
	App    string `json:"app"`
	Title  string `json:"title"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Index  int    `json:"index"`
}

func (t *WindowTool) Execute(ctx context.Context, input json.RawMessage) (*ToolResult, error) {
	var in WindowInput
	if err := json.Unmarshal(input, &in); err != nil {
		return &ToolResult{
			Content: fmt.Sprintf("Failed to parse input: %v", err),
			IsError: true,
		}, nil
	}

	var result string
	var err error

	switch in.Action {
	case "list":
		result, err = t.listWindows()
	case "focus":
		result, err = t.focusWindow(in.App, in.Title)
	case "move":
		result, err = t.moveWindow(in.App, in.Title, in.X, in.Y)
	case "resize":
		result, err = t.resizeWindow(in.App, in.Title, in.Width, in.Height)
	case "minimize":
		result, err = t.minimizeWindow(in.App, in.Title)
	case "maximize":
		result, err = t.maximizeWindow(in.App, in.Title)
	case "close":
		result, err = t.closeWindow(in.App, in.Title)
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

func (t *WindowTool) listWindows() (string, error) {
	// AppleScript to list all windows from all apps
	script := `
		set windowList to ""
		tell application "System Events"
			set appList to (name of every process whose visible is true)
			repeat with appName in appList
				try
					tell process appName
						set winCount to count of windows
						if winCount > 0 then
							repeat with i from 1 to winCount
								set win to window i
								set winTitle to name of win
								set winPos to position of win
								set winSize to size of win
								set windowList to windowList & appName & "|||" & winTitle & "|||" & (item 1 of winPos) & "|||" & (item 2 of winPos) & "|||" & (item 1 of winSize) & "|||" & (item 2 of winSize) & "|||" & i & "
"
							end repeat
						end if
					end tell
				end try
			end repeat
		end tell
		return windowList
	`

	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return "", fmt.Errorf("failed to list windows: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var windows []WindowInfo

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|||")
		if len(parts) >= 7 {
			x, _ := strconv.Atoi(parts[2])
			y, _ := strconv.Atoi(parts[3])
			w, _ := strconv.Atoi(parts[4])
			h, _ := strconv.Atoi(parts[5])
			idx, _ := strconv.Atoi(parts[6])
			windows = append(windows, WindowInfo{
				App:    parts[0],
				Title:  parts[1],
				X:      x,
				Y:      y,
				Width:  w,
				Height: h,
				Index:  idx,
			})
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d windows:\n\n", len(windows)))
	for _, win := range windows {
		title := win.Title
		if len(title) > 50 {
			title = title[:47] + "..."
		}
		sb.WriteString(fmt.Sprintf("• %s\n", win.App))
		sb.WriteString(fmt.Sprintf("  Title: %s\n", title))
		sb.WriteString(fmt.Sprintf("  Position: (%d, %d), Size: %dx%d\n\n", win.X, win.Y, win.Width, win.Height))
	}

	return sb.String(), nil
}

func (t *WindowTool) focusWindow(app, title string) (string, error) {
	if app == "" {
		return "", fmt.Errorf("app name is required")
	}

	var script string
	if title != "" {
		script = fmt.Sprintf(`
			tell application "System Events"
				tell process "%s"
					set frontmost to true
					repeat with win in windows
						if name of win contains "%s" then
							perform action "AXRaise" of win
							return "focused"
						end if
					end repeat
				end tell
			end tell
			return "not found"
		`, app, title)
	} else {
		script = fmt.Sprintf(`
			tell application "%s"
				activate
			end tell
			return "focused"
		`, app)
	}

	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return "", fmt.Errorf("failed to focus window: %v", err)
	}

	result := strings.TrimSpace(string(out))
	if result == "not found" {
		return "", fmt.Errorf("window with title containing '%s' not found in %s", title, app)
	}

	if title != "" {
		return fmt.Sprintf("Focused window '%s' in %s", title, app), nil
	}
	return fmt.Sprintf("Focused %s", app), nil
}

func (t *WindowTool) moveWindow(app, title string, x, y int) (string, error) {
	if app == "" {
		return "", fmt.Errorf("app name is required")
	}

	var script string
	if title != "" {
		script = fmt.Sprintf(`
			tell application "System Events"
				tell process "%s"
					repeat with win in windows
						if name of win contains "%s" then
							set position of win to {%d, %d}
							return "moved"
						end if
					end repeat
				end tell
			end tell
			return "not found"
		`, app, title, x, y)
	} else {
		script = fmt.Sprintf(`
			tell application "System Events"
				tell process "%s"
					if (count of windows) > 0 then
						set position of window 1 to {%d, %d}
						return "moved"
					end if
				end tell
			end tell
			return "no windows"
		`, app, x, y)
	}

	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return "", fmt.Errorf("failed to move window: %v", err)
	}

	result := strings.TrimSpace(string(out))
	if result == "not found" || result == "no windows" {
		return "", fmt.Errorf("window not found")
	}

	return fmt.Sprintf("Moved window to (%d, %d)", x, y), nil
}

func (t *WindowTool) resizeWindow(app, title string, width, height int) (string, error) {
	if app == "" {
		return "", fmt.Errorf("app name is required")
	}
	if width <= 0 || height <= 0 {
		return "", fmt.Errorf("width and height must be positive")
	}

	var script string
	if title != "" {
		script = fmt.Sprintf(`
			tell application "System Events"
				tell process "%s"
					repeat with win in windows
						if name of win contains "%s" then
							set size of win to {%d, %d}
							return "resized"
						end if
					end repeat
				end tell
			end tell
			return "not found"
		`, app, title, width, height)
	} else {
		script = fmt.Sprintf(`
			tell application "System Events"
				tell process "%s"
					if (count of windows) > 0 then
						set size of window 1 to {%d, %d}
						return "resized"
					end if
				end tell
			end tell
			return "no windows"
		`, app, width, height)
	}

	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return "", fmt.Errorf("failed to resize window: %v", err)
	}

	result := strings.TrimSpace(string(out))
	if result == "not found" || result == "no windows" {
		return "", fmt.Errorf("window not found")
	}

	return fmt.Sprintf("Resized window to %dx%d", width, height), nil
}

func (t *WindowTool) minimizeWindow(app, title string) (string, error) {
	if app == "" {
		return "", fmt.Errorf("app name is required")
	}

	var script string
	if title != "" {
		script = fmt.Sprintf(`
			tell application "System Events"
				tell process "%s"
					repeat with win in windows
						if name of win contains "%s" then
							set value of attribute "AXMinimized" of win to true
							return "minimized"
						end if
					end repeat
				end tell
			end tell
			return "not found"
		`, app, title)
	} else {
		script = fmt.Sprintf(`
			tell application "System Events"
				tell process "%s"
					if (count of windows) > 0 then
						set value of attribute "AXMinimized" of window 1 to true
						return "minimized"
					end if
				end tell
			end tell
			return "no windows"
		`, app)
	}

	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return "", fmt.Errorf("failed to minimize window: %v", err)
	}

	result := strings.TrimSpace(string(out))
	if result == "not found" || result == "no windows" {
		return "", fmt.Errorf("window not found")
	}

	return "Minimized window", nil
}

func (t *WindowTool) maximizeWindow(app, title string) (string, error) {
	if app == "" {
		return "", fmt.Errorf("app name is required")
	}

	// Get screen size first
	sizeScript := `
		tell application "Finder"
			set screenBounds to bounds of window of desktop
			return (item 3 of screenBounds) & "," & (item 4 of screenBounds)
		end tell
	`
	sizeOut, _ := exec.Command("osascript", "-e", sizeScript).Output()
	parts := strings.Split(strings.TrimSpace(string(sizeOut)), ",")
	screenWidth := 1920
	screenHeight := 1080
	if len(parts) >= 2 {
		if w, err := strconv.Atoi(parts[0]); err == nil {
			screenWidth = w
		}
		if h, err := strconv.Atoi(parts[1]); err == nil {
			screenHeight = h
		}
	}

	// Account for menu bar (roughly 25 pixels)
	screenHeight -= 25

	var script string
	if title != "" {
		script = fmt.Sprintf(`
			tell application "System Events"
				tell process "%s"
					repeat with win in windows
						if name of win contains "%s" then
							set position of win to {0, 25}
							set size of win to {%d, %d}
							return "maximized"
						end if
					end repeat
				end tell
			end tell
			return "not found"
		`, app, title, screenWidth, screenHeight)
	} else {
		script = fmt.Sprintf(`
			tell application "System Events"
				tell process "%s"
					if (count of windows) > 0 then
						set position of window 1 to {0, 25}
						set size of window 1 to {%d, %d}
						return "maximized"
					end if
				end tell
			end tell
			return "no windows"
		`, app, screenWidth, screenHeight)
	}

	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return "", fmt.Errorf("failed to maximize window: %v", err)
	}

	result := strings.TrimSpace(string(out))
	if result == "not found" || result == "no windows" {
		return "", fmt.Errorf("window not found")
	}

	return fmt.Sprintf("Maximized window to %dx%d", screenWidth, screenHeight), nil
}

func (t *WindowTool) closeWindow(app, title string) (string, error) {
	if app == "" {
		return "", fmt.Errorf("app name is required")
	}

	var script string
	if title != "" {
		script = fmt.Sprintf(`
			tell application "System Events"
				tell process "%s"
					repeat with win in windows
						if name of win contains "%s" then
							click button 1 of win
							return "closed"
						end if
					end repeat
				end tell
			end tell
			return "not found"
		`, app, title)
	} else {
		script = fmt.Sprintf(`
			tell application "System Events"
				tell process "%s"
					if (count of windows) > 0 then
						click button 1 of window 1
						return "closed"
					end if
				end tell
			end tell
			return "no windows"
		`, app)
	}

	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return "", fmt.Errorf("failed to close window: %v", err)
	}

	result := strings.TrimSpace(string(out))
	if result == "not found" || result == "no windows" {
		return "", fmt.Errorf("window not found")
	}

	return "Closed window", nil
}

func (t *WindowTool) RequiresApproval() bool {
	return true // Window control is sensitive
}

// RPC Server implementation
type ToolRPCServer struct {
	Impl *WindowTool
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
	Impl *WindowTool
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
			"tool": &ToolPlugin{Impl: &WindowTool{}},
		},
	})
}
