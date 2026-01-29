// Accessibility plugin for macOS.
// Provides: tree (UI element tree), find (find element), click_element, get_value, set_value
// Uses macOS Accessibility APIs via AppleScript/JXA
// Build with: go build -o accessibility
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

type AccessibilityTool struct{}

func (t *AccessibilityTool) Name() string {
	return "accessibility"
}

func (t *AccessibilityTool) Description() string {
	return "Access macOS UI elements via Accessibility APIs: list UI element tree, find elements by role/label, click buttons, read/set text fields. Requires Accessibility permissions in System Preferences."
}

func (t *AccessibilityTool) Schema() json.RawMessage {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"action": map[string]any{
				"type":        "string",
				"description": "Action: tree (get UI tree), find (find element), click (click element), get_value, set_value, list_apps",
				"enum":        []string{"tree", "find", "click", "get_value", "set_value", "list_apps"},
			},
			"app": map[string]any{
				"type":        "string",
				"description": "Application name (e.g., 'Safari', 'Finder'). Required for most actions.",
			},
			"role": map[string]any{
				"type":        "string",
				"description": "UI element role to find: button, textfield, checkbox, menu, menuitem, window, etc.",
			},
			"label": map[string]any{
				"type":        "string",
				"description": "Element label/title to match (partial match supported)",
			},
			"value": map[string]any{
				"type":        "string",
				"description": "Value to set (for set_value action)",
			},
			"max_depth": map[string]any{
				"type":        "integer",
				"description": "Maximum depth to traverse for tree action (default: 3)",
			},
		},
		"required": []string{"action"},
	}
	data, _ := json.Marshal(schema)
	return data
}

type AccessibilityInput struct {
	Action   string `json:"action"`
	App      string `json:"app"`
	Role     string `json:"role"`
	Label    string `json:"label"`
	Value    string `json:"value"`
	MaxDepth int    `json:"max_depth"`
}

func (t *AccessibilityTool) Execute(ctx context.Context, input json.RawMessage) (*ToolResult, error) {
	var in AccessibilityInput
	if err := json.Unmarshal(input, &in); err != nil {
		return &ToolResult{
			Content: fmt.Sprintf("Failed to parse input: %v", err),
			IsError: true,
		}, nil
	}

	var result string
	var err error

	switch in.Action {
	case "tree":
		if in.MaxDepth <= 0 {
			in.MaxDepth = 3
		}
		result, err = t.getUITree(in.App, in.MaxDepth)
	case "find":
		result, err = t.findElement(in.App, in.Role, in.Label)
	case "click":
		result, err = t.clickElement(in.App, in.Role, in.Label)
	case "get_value":
		result, err = t.getValue(in.App, in.Role, in.Label)
	case "set_value":
		result, err = t.setValue(in.App, in.Role, in.Label, in.Value)
	case "list_apps":
		result, err = t.listApps()
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

func (t *AccessibilityTool) listApps() (string, error) {
	script := `
		tell application "System Events"
			set appList to ""
			repeat with proc in (every process whose visible is true)
				set appList to appList & name of proc & "
"
			end repeat
		end tell
		return appList
	`

	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return "", fmt.Errorf("failed to list apps: %v", err)
	}

	apps := strings.Split(strings.TrimSpace(string(out)), "\n")
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Visible Applications (%d):\n\n", len(apps)))
	for _, app := range apps {
		if app != "" {
			sb.WriteString(fmt.Sprintf("• %s\n", app))
		}
	}

	return sb.String(), nil
}

func (t *AccessibilityTool) getUITree(app string, maxDepth int) (string, error) {
	if app == "" {
		return "", fmt.Errorf("app name is required")
	}

	// JXA (JavaScript for Automation) gives better access to accessibility tree
	// But AppleScript works for basic cases
	script := fmt.Sprintf(`
		on getUITree(elem, depth, maxD, indent)
			if depth > maxD then return ""

			set result to ""
			try
				set elemRole to role of elem
				set elemDesc to ""
				try
					set elemDesc to description of elem
				end try
				set elemTitle to ""
				try
					set elemTitle to title of elem
				end try
				set elemValue to ""
				try
					set elemValue to value of elem
				end try

				set line to indent & "- " & elemRole
				if elemTitle is not "" then set line to line & " \"" & elemTitle & "\""
				if elemDesc is not "" and elemDesc is not elemTitle then set line to line & " [" & elemDesc & "]"
				if elemValue is not "" then set line to line & " = " & elemValue
				set result to result & line & "
"

				-- Recurse into children
				try
					set children to UI elements of elem
					repeat with child in children
						set result to result & getUITree(child, depth + 1, maxD, indent & "  ")
					end repeat
				end try
			on error errMsg
				-- Skip elements we can't access
			end try

			return result
		end getUITree

		tell application "System Events"
			tell process "%s"
				set uiTree to ""
				repeat with win in windows
					set uiTree to uiTree & "Window: " & (name of win) & "
"
					set uiTree to uiTree & my getUITree(win, 1, %d, "  ")
				end repeat
				return uiTree
			end tell
		end tell
	`, app, maxDepth)

	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return "", fmt.Errorf("failed to get UI tree for %s: %v (ensure app is running and Accessibility is enabled)", app, err)
	}

	result := strings.TrimSpace(string(out))
	if result == "" {
		return fmt.Sprintf("No UI elements found in %s (app may have no windows or Accessibility access denied)", app), nil
	}

	return fmt.Sprintf("UI Tree for %s:\n\n%s", app, result), nil
}

func (t *AccessibilityTool) findElement(app, role, label string) (string, error) {
	if app == "" {
		return "", fmt.Errorf("app name is required")
	}
	if role == "" && label == "" {
		return "", fmt.Errorf("role or label is required")
	}

	// Map friendly role names to AX roles
	axRole := t.mapRole(role)

	var script string
	if role != "" && label != "" {
		script = fmt.Sprintf(`
			tell application "System Events"
				tell process "%s"
					set foundElems to {}
					repeat with win in windows
						try
							set elems to every %s of win whose title contains "%s" or description contains "%s"
							repeat with elem in elems
								set end of foundElems to {role:role of elem, title:(title of elem), desc:(description of elem)}
							end repeat
						end try
						-- Also check nested elements
						try
							set elems to every %s of every group of win whose title contains "%s" or description contains "%s"
							repeat with elemList in elems
								repeat with elem in elemList
									set end of foundElems to {role:role of elem, title:(title of elem), desc:(description of elem)}
								end repeat
							end repeat
						end try
					end repeat
					return foundElems
				end tell
			end tell
		`, app, axRole, label, label, axRole, label, label)
	} else if role != "" {
		script = fmt.Sprintf(`
			tell application "System Events"
				tell process "%s"
					set foundElems to {}
					repeat with win in windows
						try
							set elems to every %s of win
							repeat with elem in elems
								try
									set end of foundElems to {role:role of elem, title:(title of elem), desc:(description of elem)}
								end try
							end repeat
						end try
					end repeat
					return foundElems
				end tell
			end tell
		`, app, axRole)
	} else {
		script = fmt.Sprintf(`
			tell application "System Events"
				tell process "%s"
					set foundElems to {}
					repeat with win in windows
						try
							set elems to every UI element of win whose title contains "%s" or description contains "%s"
							repeat with elem in elems
								set end of foundElems to {role:role of elem, title:(title of elem), desc:(description of elem)}
							end repeat
						end try
					end repeat
					return foundElems
				end tell
			end tell
		`, app, label, label)
	}

	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return "", fmt.Errorf("failed to find elements: %v", err)
	}

	result := strings.TrimSpace(string(out))
	if result == "" || result == "{}" {
		return "No matching elements found", nil
	}

	return fmt.Sprintf("Found elements in %s:\n%s", app, result), nil
}

func (t *AccessibilityTool) clickElement(app, role, label string) (string, error) {
	if app == "" {
		return "", fmt.Errorf("app name is required")
	}
	if label == "" {
		return "", fmt.Errorf("label is required to identify element to click")
	}

	axRole := t.mapRole(role)
	if axRole == "" {
		axRole = "UI element"
	}

	script := fmt.Sprintf(`
		tell application "System Events"
			tell process "%s"
				set frontmost to true
				repeat with win in windows
					try
						click (first %s of win whose title contains "%s" or description contains "%s")
						return "clicked"
					end try
					-- Try nested in groups
					try
						click (first %s of first group of win whose title contains "%s" or description contains "%s")
						return "clicked"
					end try
					-- Try in scroll areas
					try
						click (first %s of first scroll area of win whose title contains "%s" or description contains "%s")
						return "clicked"
					end try
				end repeat
			end tell
		end tell
		return "not found"
	`, app, axRole, label, label, axRole, label, label, axRole, label, label)

	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return "", fmt.Errorf("failed to click element: %v", err)
	}

	result := strings.TrimSpace(string(out))
	if result == "not found" {
		return "", fmt.Errorf("element with label '%s' not found in %s", label, app)
	}

	return fmt.Sprintf("Clicked '%s' in %s", label, app), nil
}

func (t *AccessibilityTool) getValue(app, role, label string) (string, error) {
	if app == "" {
		return "", fmt.Errorf("app name is required")
	}
	if label == "" {
		return "", fmt.Errorf("label is required")
	}

	axRole := t.mapRole(role)
	if axRole == "" {
		axRole = "UI element"
	}

	script := fmt.Sprintf(`
		tell application "System Events"
			tell process "%s"
				repeat with win in windows
					try
						set elem to (first %s of win whose title contains "%s" or description contains "%s")
						return value of elem
					end try
					try
						set elem to (first %s of first group of win whose title contains "%s" or description contains "%s")
						return value of elem
					end try
				end repeat
			end tell
		end tell
		return "not found"
	`, app, axRole, label, label, axRole, label, label)

	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return "", fmt.Errorf("failed to get value: %v", err)
	}

	result := strings.TrimSpace(string(out))
	if result == "not found" {
		return "", fmt.Errorf("element '%s' not found", label)
	}

	return fmt.Sprintf("Value of '%s': %s", label, result), nil
}

func (t *AccessibilityTool) setValue(app, role, label, value string) (string, error) {
	if app == "" {
		return "", fmt.Errorf("app name is required")
	}
	if label == "" {
		return "", fmt.Errorf("label is required")
	}
	if value == "" {
		return "", fmt.Errorf("value is required")
	}

	axRole := t.mapRole(role)
	if axRole == "" {
		axRole = "text field"
	}

	// Escape value for AppleScript
	escapedValue := strings.ReplaceAll(value, `"`, `\"`)

	script := fmt.Sprintf(`
		tell application "System Events"
			tell process "%s"
				set frontmost to true
				repeat with win in windows
					try
						set elem to (first %s of win whose title contains "%s" or description contains "%s")
						set value of elem to "%s"
						return "set"
					end try
					try
						set elem to (first %s of first group of win whose title contains "%s" or description contains "%s")
						set value of elem to "%s"
						return "set"
					end try
				end repeat
			end tell
		end tell
		return "not found"
	`, app, axRole, label, label, escapedValue, axRole, label, label, escapedValue)

	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return "", fmt.Errorf("failed to set value: %v", err)
	}

	result := strings.TrimSpace(string(out))
	if result == "not found" {
		return "", fmt.Errorf("element '%s' not found", label)
	}

	return fmt.Sprintf("Set '%s' to: %s", label, value), nil
}

func (t *AccessibilityTool) mapRole(role string) string {
	roleMap := map[string]string{
		"button":     "button",
		"textfield":  "text field",
		"text field": "text field",
		"checkbox":   "checkbox",
		"radio":      "radio button",
		"menu":       "menu",
		"menuitem":   "menu item",
		"menu item":  "menu item",
		"window":     "window",
		"tab":        "tab group",
		"list":       "list",
		"table":      "table",
		"row":        "row",
		"cell":       "cell",
		"image":      "image",
		"link":       "link",
		"group":      "group",
		"scroll":     "scroll area",
		"toolbar":    "toolbar",
		"popup":      "pop up button",
		"combo":      "combo box",
		"slider":     "slider",
		"static":     "static text",
		"text":       "static text",
	}

	if mapped, ok := roleMap[strings.ToLower(role)]; ok {
		return mapped
	}
	return role
}

func (t *AccessibilityTool) RequiresApproval() bool {
	return true // Accessibility control is sensitive
}

// RPC Server implementation
type ToolRPCServer struct {
	Impl *AccessibilityTool
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
	Impl *AccessibilityTool
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
			"tool": &ToolPlugin{Impl: &AccessibilityTool{}},
		},
	})
}
