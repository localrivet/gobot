// Notification plugin for macOS.
// Provides: send (display notification), schedule, clear, do_not_disturb
// Build with: go build -o notification
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

type NotificationTool struct{}

func (t *NotificationTool) Name() string {
	return "notification"
}

func (t *NotificationTool) Description() string {
	return "Display macOS notifications: send notifications with title/body/sound, check Do Not Disturb status. Can also trigger system alerts and speak text."
}

func (t *NotificationTool) Schema() json.RawMessage {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"action": map[string]any{
				"type":        "string",
				"description": "Action: send (display notification), alert (system alert dialog), speak (text-to-speech), dnd_status (check Do Not Disturb)",
				"enum":        []string{"send", "alert", "speak", "dnd_status"},
			},
			"title": map[string]any{
				"type":        "string",
				"description": "Notification title",
			},
			"message": map[string]any{
				"type":        "string",
				"description": "Notification body text",
			},
			"subtitle": map[string]any{
				"type":        "string",
				"description": "Notification subtitle (optional)",
			},
			"sound": map[string]any{
				"type":        "string",
				"description": "Sound name: default, Basso, Blow, Bottle, Frog, Funk, Glass, Hero, Morse, Ping, Pop, Purr, Sosumi, Submarine, Tink",
			},
			"voice": map[string]any{
				"type":        "string",
				"description": "Voice for speak action (e.g., 'Alex', 'Samantha', 'Daniel'). Use 'list' to see available voices.",
			},
		},
		"required": []string{"action"},
	}
	data, _ := json.Marshal(schema)
	return data
}

type NotificationInput struct {
	Action   string `json:"action"`
	Title    string `json:"title"`
	Message  string `json:"message"`
	Subtitle string `json:"subtitle"`
	Sound    string `json:"sound"`
	Voice    string `json:"voice"`
}

func (t *NotificationTool) Execute(ctx context.Context, input json.RawMessage) (*ToolResult, error) {
	var in NotificationInput
	if err := json.Unmarshal(input, &in); err != nil {
		return &ToolResult{
			Content: fmt.Sprintf("Failed to parse input: %v", err),
			IsError: true,
		}, nil
	}

	var result string
	var err error

	switch in.Action {
	case "send":
		result, err = t.sendNotification(in)
	case "alert":
		result, err = t.showAlert(in)
	case "speak":
		result, err = t.speak(in)
	case "dnd_status":
		result, err = t.getDNDStatus()
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

func (t *NotificationTool) sendNotification(in NotificationInput) (string, error) {
	if in.Message == "" {
		return "", fmt.Errorf("message is required")
	}

	// Escape strings for AppleScript
	title := strings.ReplaceAll(in.Title, `"`, `\"`)
	message := strings.ReplaceAll(in.Message, `"`, `\"`)
	subtitle := strings.ReplaceAll(in.Subtitle, `"`, `\"`)

	if title == "" {
		title = "GoBot"
	}

	var script strings.Builder
	script.WriteString(fmt.Sprintf(`display notification "%s" with title "%s"`, message, title))

	if subtitle != "" {
		script.WriteString(fmt.Sprintf(` subtitle "%s"`, subtitle))
	}

	if in.Sound != "" {
		script.WriteString(fmt.Sprintf(` sound name "%s"`, in.Sound))
	}

	_, err := exec.Command("osascript", "-e", script.String()).Output()
	if err != nil {
		return "", fmt.Errorf("failed to send notification: %v", err)
	}

	return fmt.Sprintf("Notification sent: %s - %s", title, message), nil
}

func (t *NotificationTool) showAlert(in NotificationInput) (string, error) {
	if in.Message == "" {
		return "", fmt.Errorf("message is required")
	}

	title := strings.ReplaceAll(in.Title, `"`, `\"`)
	message := strings.ReplaceAll(in.Message, `"`, `\"`)

	if title == "" {
		title = "GoBot Alert"
	}

	script := fmt.Sprintf(`
		display alert "%s" message "%s" as informational
		return "shown"
	`, title, message)

	_, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return "", fmt.Errorf("failed to show alert: %v", err)
	}

	return fmt.Sprintf("Alert shown: %s", title), nil
}

func (t *NotificationTool) speak(in NotificationInput) (string, error) {
	// Handle list voices request
	if in.Voice == "list" {
		out, err := exec.Command("say", "-v", "?").Output()
		if err != nil {
			return "", fmt.Errorf("failed to list voices: %v", err)
		}

		lines := strings.Split(string(out), "\n")
		var voices []string
		for _, line := range lines {
			if line != "" {
				parts := strings.Fields(line)
				if len(parts) > 0 {
					voices = append(voices, parts[0])
				}
			}
		}

		// Return first 20 voices
		if len(voices) > 20 {
			voices = voices[:20]
		}

		return fmt.Sprintf("Available voices (first 20):\n%s", strings.Join(voices, ", ")), nil
	}

	if in.Message == "" {
		return "", fmt.Errorf("message is required")
	}

	args := []string{}
	if in.Voice != "" {
		args = append(args, "-v", in.Voice)
	}
	args = append(args, in.Message)

	cmd := exec.Command("say", args...)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to speak: %v", err)
	}

	voice := in.Voice
	if voice == "" {
		voice = "default"
	}

	return fmt.Sprintf("Spoke with voice '%s': %s", voice, in.Message), nil
}

func (t *NotificationTool) getDNDStatus() (string, error) {
	// Check Do Not Disturb status (Focus mode in newer macOS)
	// This is tricky as it changed in macOS Monterey+

	// Try the newer Focus mode approach first
	script := `
		try
			-- Check for Focus mode (macOS 12+)
			do shell script "plutil -extract data.0.storeAssertionRecords xml1 -o - ~/Library/DoNotDisturb/DB/Assertions.json 2>/dev/null | grep -c assertionDetails || echo 0"
		on error
			-- Fallback for older macOS
			do shell script "defaults read com.apple.controlcenter 'NSStatusItem Visible DoNotDisturb' 2>/dev/null || echo '0'"
		end try
	`

	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		// Another approach
		out2, _ := exec.Command("bash", "-c", "defaults -currentHost read ~/Library/Preferences/ByHost/com.apple.notificationcenterui doNotDisturb 2>/dev/null || echo 0").Output()
		result := strings.TrimSpace(string(out2))
		if result == "1" {
			return "Do Not Disturb: ON", nil
		}
		return "Do Not Disturb: OFF (or unable to determine)", nil
	}

	result := strings.TrimSpace(string(out))
	if result != "0" && result != "" {
		return "Focus/Do Not Disturb: ON", nil
	}

	return "Focus/Do Not Disturb: OFF", nil
}

func (t *NotificationTool) RequiresApproval() bool {
	return false // Notifications are relatively safe
}

// RPC Server implementation
type ToolRPCServer struct {
	Impl *NotificationTool
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
	Impl *NotificationTool
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
			"tool": &ToolPlugin{Impl: &NotificationTool{}},
		},
	})
}
