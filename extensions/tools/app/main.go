// Application control plugin for macOS.
// Provides: list, launch, quit, activate, hide, info, menu
// Build with: go build -o app
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/rpc"
	"os/exec"
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

type AppTool struct{}

func (t *AppTool) Name() string {
	return "app"
}

func (t *AppTool) Description() string {
	return "Control macOS applications: list running apps, launch apps, quit apps, activate/bring to front, hide, get app info, and interact with menu bar."
}

func (t *AppTool) Schema() json.RawMessage {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"action": map[string]any{
				"type":        "string",
				"description": "Action: list (running apps), launch, quit, quit_all (quit all apps), activate, hide, info, menu (click menu), frontmost",
				"enum":        []string{"list", "launch", "quit", "quit_all", "activate", "hide", "info", "menu", "frontmost"},
			},
			"name": map[string]any{
				"type":        "string",
				"description": "Application name (e.g., 'Safari', 'Google Chrome', 'Terminal')",
			},
			"path": map[string]any{
				"type":        "string",
				"description": "Application path for launch (e.g., '/Applications/Safari.app'). Optional if name is provided.",
			},
			"menu_path": map[string]any{
				"type":        "string",
				"description": "Menu path for menu action, separated by > (e.g., 'File > New Window', 'Edit > Copy')",
			},
			"force": map[string]any{
				"type":        "boolean",
				"description": "Force quit without saving (for quit action)",
			},
		},
		"required": []string{"action"},
	}
	data, _ := json.Marshal(schema)
	return data
}

type AppInput struct {
	Action   string `json:"action"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	MenuPath string `json:"menu_path"`
	Force    bool   `json:"force"`
}

func (t *AppTool) Execute(ctx context.Context, input json.RawMessage) (*ToolResult, error) {
	var in AppInput
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
		result, err = t.listApps()
	case "launch":
		result, err = t.launchApp(in.Name, in.Path)
	case "quit":
		result, err = t.quitApp(in.Name, in.Force)
	case "quit_all":
		result, err = t.quitAllApps(in.Force)
	case "activate":
		result, err = t.activateApp(in.Name)
	case "hide":
		result, err = t.hideApp(in.Name)
	case "info":
		result, err = t.getAppInfo(in.Name)
	case "menu":
		result, err = t.clickMenu(in.Name, in.MenuPath)
	case "frontmost":
		result, err = t.getFrontmostApp()
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

func (t *AppTool) listApps() (string, error) {
	script := `
		tell application "System Events"
			set appInfo to ""
			repeat with proc in (every process whose background only is false)
				set appName to name of proc
				set appFront to frontmost of proc
				set winCount to 0
				try
					set winCount to count of windows of proc
				end try
				set frontMark to ""
				if appFront then set frontMark to " (frontmost)"
				set appInfo to appInfo & appName & frontMark & " - " & winCount & " windows
"
			end repeat
		end tell
		return appInfo
	`

	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return "", fmt.Errorf("failed to list apps: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Running Applications (%d):\n\n", len(lines)))
	for _, line := range lines {
		if line != "" {
			sb.WriteString(fmt.Sprintf("• %s\n", line))
		}
	}

	return sb.String(), nil
}

func (t *AppTool) launchApp(name, path string) (string, error) {
	if name == "" && path == "" {
		return "", fmt.Errorf("name or path is required")
	}

	var script string
	if path != "" {
		script = fmt.Sprintf(`
			tell application "Finder"
				open POSIX file "%s"
			end tell
			return "launched"
		`, path)
	} else {
		script = fmt.Sprintf(`
			tell application "%s"
				activate
			end tell
			return "launched"
		`, name)
	}

	_, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		// Try open command as fallback
		var cmd *exec.Cmd
		if path != "" {
			cmd = exec.Command("open", path)
		} else {
			cmd = exec.Command("open", "-a", name)
		}
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("failed to launch app: %v", err)
		}
	}

	appName := name
	if appName == "" {
		appName = path
	}
	return fmt.Sprintf("Launched %s", appName), nil
}

func (t *AppTool) quitApp(name string, force bool) (string, error) {
	if name == "" {
		return "", fmt.Errorf("name is required")
	}

	var script string
	if force {
		script = fmt.Sprintf(`
			tell application "System Events"
				set targetProcess to first process whose name is "%s"
				do shell script "kill -9 " & (unix id of targetProcess)
			end tell
			return "force quit"
		`, name)
	} else {
		script = fmt.Sprintf(`
			tell application "%s"
				quit
			end tell
			return "quit"
		`, name)
	}

	_, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return "", fmt.Errorf("failed to quit %s: %v", name, err)
	}

	if force {
		return fmt.Sprintf("Force quit %s", name), nil
	}
	return fmt.Sprintf("Quit %s", name), nil
}

func (t *AppTool) quitAllApps(force bool) (string, error) {
	script := `
		tell application "System Events"
			set appList to name of every process whose background only is false
			set quitCount to 0
			repeat with appName in appList
				-- Don't quit Finder or the script itself
				if appName is not "Finder" and appName is not "osascript" then
					try
						tell application appName to quit
						set quitCount to quitCount + 1
					end try
				end if
			end repeat
			return quitCount as string
		end tell
	`

	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return "", fmt.Errorf("failed to quit apps: %v", err)
	}

	count := strings.TrimSpace(string(out))
	return fmt.Sprintf("Requested quit for %s applications", count), nil
}

func (t *AppTool) activateApp(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("name is required")
	}

	script := fmt.Sprintf(`
		tell application "%s"
			activate
		end tell
		return "activated"
	`, name)

	_, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return "", fmt.Errorf("failed to activate %s: %v", name, err)
	}

	return fmt.Sprintf("Activated %s", name), nil
}

func (t *AppTool) hideApp(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("name is required")
	}

	script := fmt.Sprintf(`
		tell application "System Events"
			set visible of process "%s" to false
		end tell
		return "hidden"
	`, name)

	_, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return "", fmt.Errorf("failed to hide %s: %v", name, err)
	}

	return fmt.Sprintf("Hidden %s", name), nil
}

func (t *AppTool) getAppInfo(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("name is required")
	}

	script := fmt.Sprintf(`
		tell application "System Events"
			set proc to first process whose name is "%s"
			set appName to name of proc
			set appPath to ""
			try
				set appPath to POSIX path of (file of proc as alias)
			end try
			set appPID to unix id of proc
			set appFront to frontmost of proc
			set winCount to 0
			try
				set winCount to count of windows of proc
			end try
			set appVisible to visible of proc
		end tell

		set info to "Name: " & appName & "
"
		set info to info & "Path: " & appPath & "
"
		set info to info & "PID: " & appPID & "
"
		set info to info & "Frontmost: " & appFront & "
"
		set info to info & "Visible: " & appVisible & "
"
		set info to info & "Windows: " & winCount

		return info
	`, name)

	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return "", fmt.Errorf("failed to get info for %s: %v", name, err)
	}

	return fmt.Sprintf("Application Info:\n\n%s", strings.TrimSpace(string(out))), nil
}

func (t *AppTool) clickMenu(name, menuPath string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("name is required")
	}
	if menuPath == "" {
		return "", fmt.Errorf("menu_path is required (e.g., 'File > New Window')")
	}

	// Parse menu path
	parts := strings.Split(menuPath, ">")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	if len(parts) < 2 {
		return "", fmt.Errorf("menu_path must have at least menu and item (e.g., 'File > New')")
	}

	// Build AppleScript for nested menu navigation
	var menuScript strings.Builder
	menuScript.WriteString(fmt.Sprintf(`
		tell application "%s"
			activate
		end tell
		delay 0.2
		tell application "System Events"
			tell process "%s"
	`, name, name))

	// First level is the menu bar menu
	menuScript.WriteString(fmt.Sprintf(`				tell menu bar 1
					tell menu bar item "%s"
						click
						delay 0.1
						tell menu 1
	`, parts[0]))

	// Navigate through submenus
	for i := 1; i < len(parts)-1; i++ {
		menuScript.WriteString(fmt.Sprintf(`							tell menu item "%s"
								click
								delay 0.1
								tell menu 1
	`, parts[i]))
	}

	// Click the final menu item
	menuScript.WriteString(fmt.Sprintf(`							click menu item "%s"
	`, parts[len(parts)-1]))

	// Close all the nested tells
	for i := 1; i < len(parts); i++ {
		menuScript.WriteString(`						end tell
	`)
		if i < len(parts)-1 {
			menuScript.WriteString(`					end tell
	`)
		}
	}

	menuScript.WriteString(`					end tell
				end tell
			end tell
		end tell
		return "clicked"
	`)

	out, err := exec.Command("osascript", "-e", menuScript.String()).Output()
	if err != nil {
		// Try simpler approach for basic menus
		simpleScript := fmt.Sprintf(`
			tell application "%s"
				activate
			end tell
			delay 0.2
			tell application "System Events"
				tell process "%s"
					click menu item "%s" of menu 1 of menu bar item "%s" of menu bar 1
				end tell
			end tell
			return "clicked"
		`, name, name, parts[len(parts)-1], parts[0])

		out, err = exec.Command("osascript", "-e", simpleScript).Output()
		if err != nil {
			return "", fmt.Errorf("failed to click menu '%s': %v", menuPath, err)
		}
	}

	if strings.TrimSpace(string(out)) == "clicked" {
		return fmt.Sprintf("Clicked menu: %s > %s", name, menuPath), nil
	}

	return "", fmt.Errorf("menu click may have failed")
}

func (t *AppTool) getFrontmostApp() (string, error) {
	script := `
		tell application "System Events"
			set frontApp to first process whose frontmost is true
			set appName to name of frontApp
			set winName to ""
			try
				set winName to name of window 1 of frontApp
			end try
		end tell
		return appName & "|||" & winName
	`

	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return "", fmt.Errorf("failed to get frontmost app: %v", err)
	}

	parts := strings.Split(strings.TrimSpace(string(out)), "|||")
	appName := parts[0]
	winName := ""
	if len(parts) > 1 {
		winName = parts[1]
	}

	result := fmt.Sprintf("Frontmost application: %s", appName)
	if winName != "" {
		result += fmt.Sprintf("\nActive window: %s", winName)
	}

	return result, nil
}

func (t *AppTool) RequiresApproval() bool {
	return true // App control is sensitive
}

// RPC Server implementation
type ToolRPCServer struct {
	Impl *AppTool
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
	Impl *AppTool
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
			"tool": &ToolPlugin{Impl: &AppTool{}},
		},
	})
}
