// Example tool plugin demonstrating how to create GoBot tool plugins.
// Build with: go build -o ../example-tool
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/rpc"
	"time"

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

// ExampleTool implements a simple tool that returns the current time
type ExampleTool struct{}

func (t *ExampleTool) Name() string {
	return "example"
}

func (t *ExampleTool) Description() string {
	return "An example tool plugin that returns the current time and optional message"
}

func (t *ExampleTool) Schema() json.RawMessage {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"message": map[string]any{
				"type":        "string",
				"description": "Optional message to include in the response",
			},
			"format": map[string]any{
				"type":        "string",
				"description": "Time format: 'short', 'long', or 'unix'",
				"default":     "short",
			},
		},
	}
	data, _ := json.Marshal(schema)
	return data
}

type ExampleInput struct {
	Message string `json:"message"`
	Format  string `json:"format"`
}

func (t *ExampleTool) Execute(ctx context.Context, input json.RawMessage) (*ToolResult, error) {
	var in ExampleInput
	if err := json.Unmarshal(input, &in); err != nil {
		return &ToolResult{
			Content: fmt.Sprintf("Failed to parse input: %v", err),
			IsError: true,
		}, nil
	}

	// Format the time
	now := time.Now()
	var timeStr string
	switch in.Format {
	case "long":
		timeStr = now.Format(time.RFC1123)
	case "unix":
		timeStr = fmt.Sprintf("%d", now.Unix())
	default:
		timeStr = now.Format("15:04:05")
	}

	// Build response
	response := fmt.Sprintf("Current time: %s", timeStr)
	if in.Message != "" {
		response = fmt.Sprintf("%s\nMessage: %s", response, in.Message)
	}

	return &ToolResult{
		Content: response,
		IsError: false,
	}, nil
}

func (t *ExampleTool) RequiresApproval() bool {
	return false // Safe tool, no approval needed
}

// RPC Server implementation
type ToolRPCServer struct {
	Impl *ExampleTool
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
	Impl *ExampleTool
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
			"tool": &ToolPlugin{Impl: &ExampleTool{}},
		},
	})
}
