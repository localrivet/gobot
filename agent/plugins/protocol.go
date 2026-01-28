// Package plugins provides a hot-loadable plugin system using hashicorp/go-plugin.
// Plugins run as separate processes and communicate via RPC, enabling hot-reload
// without recompiling the main binary.
package plugins

import (
	"context"
	"encoding/json"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

// Handshake is used to verify plugin compatibility
var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "GOBOT_PLUGIN",
	MagicCookieValue: "gobot-plugin-v1",
}

// PluginMap is the map of plugins we can dispense
var PluginMap = map[string]plugin.Plugin{
	"tool":    &ToolPluginRPC{},
	"channel": &ChannelPluginRPC{},
}

// =============================================================================
// Tool Plugin Interface
// =============================================================================

// ToolResult represents the result of a tool execution
type ToolResult struct {
	Content string `json:"content"`
	IsError bool   `json:"is_error"`
}

// ToolPlugin is the interface that tool plugins must implement
type ToolPlugin interface {
	// Name returns the unique name of the tool
	Name() string

	// Description returns a human-readable description
	Description() string

	// Schema returns the JSON Schema for the tool's input
	Schema() json.RawMessage

	// Execute runs the tool with the given input
	Execute(ctx context.Context, input json.RawMessage) (*ToolResult, error)

	// RequiresApproval indicates if this tool needs user approval
	RequiresApproval() bool
}

// ToolPluginRPC is the RPC implementation of the tool plugin
type ToolPluginRPC struct {
	Impl ToolPlugin
}

func (p *ToolPluginRPC) Server(*plugin.MuxBroker) (interface{}, error) {
	return &ToolRPCServer{Impl: p.Impl}, nil
}

func (p *ToolPluginRPC) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &ToolRPCClient{client: c}, nil
}

// ToolRPCServer is the server-side RPC handler
type ToolRPCServer struct {
	Impl ToolPlugin
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

// ToolRPCClient is the client-side RPC implementation
type ToolRPCClient struct {
	client *rpc.Client
}

func (c *ToolRPCClient) Name() string {
	var resp string
	_ = c.client.Call("Plugin.Name", struct{}{}, &resp)
	return resp
}

func (c *ToolRPCClient) Description() string {
	var resp string
	_ = c.client.Call("Plugin.Description", struct{}{}, &resp)
	return resp
}

func (c *ToolRPCClient) Schema() json.RawMessage {
	var resp json.RawMessage
	_ = c.client.Call("Plugin.Schema", struct{}{}, &resp)
	return resp
}

func (c *ToolRPCClient) Execute(ctx context.Context, input json.RawMessage) (*ToolResult, error) {
	var reply ExecuteReply
	err := c.client.Call("Plugin.Execute", ExecuteArgs{Input: input}, &reply)
	if err != nil {
		return nil, err
	}
	if reply.Error != "" {
		return reply.Result, &PluginError{Message: reply.Error}
	}
	return reply.Result, nil
}

func (c *ToolRPCClient) RequiresApproval() bool {
	var resp bool
	_ = c.client.Call("Plugin.RequiresApproval", struct{}{}, &resp)
	return resp
}

// =============================================================================
// Channel Plugin Interface
// =============================================================================

// InboundMessage represents a message received from a channel
type InboundMessage struct {
	ChannelID string `json:"channel_id"`
	UserID    string `json:"user_id"`
	Text      string `json:"text"`
	Metadata  string `json:"metadata"` // JSON-encoded metadata
}

// ChannelPlugin is the interface for channel plugins
type ChannelPlugin interface {
	// ID returns the unique identifier for this channel
	ID() string

	// Connect establishes connection to the channel
	Connect(ctx context.Context, config map[string]string) error

	// Disconnect closes the channel connection
	Disconnect(ctx context.Context) error

	// Send sends a message to the channel
	Send(ctx context.Context, channelID, text string) error

	// SetHandler sets the callback for incoming messages
	SetHandler(fn func(msg InboundMessage))
}

// ChannelPluginRPC is the RPC implementation of the channel plugin
type ChannelPluginRPC struct {
	Impl ChannelPlugin
}

func (p *ChannelPluginRPC) Server(*plugin.MuxBroker) (interface{}, error) {
	return &ChannelRPCServer{Impl: p.Impl}, nil
}

func (p *ChannelPluginRPC) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &ChannelRPCClient{client: c}, nil
}

// ChannelRPCServer is the server-side RPC handler
type ChannelRPCServer struct {
	Impl ChannelPlugin
}

func (s *ChannelRPCServer) ID(_ struct{}, resp *string) error {
	*resp = s.Impl.ID()
	return nil
}

type ConnectArgs struct {
	Config map[string]string
}

func (s *ChannelRPCServer) Connect(args ConnectArgs, reply *string) error {
	err := s.Impl.Connect(context.Background(), args.Config)
	if err != nil {
		*reply = err.Error()
	}
	return nil
}

func (s *ChannelRPCServer) Disconnect(_ struct{}, reply *string) error {
	err := s.Impl.Disconnect(context.Background())
	if err != nil {
		*reply = err.Error()
	}
	return nil
}

type SendArgs struct {
	ChannelID string
	Text      string
}

func (s *ChannelRPCServer) Send(args SendArgs, reply *string) error {
	err := s.Impl.Send(context.Background(), args.ChannelID, args.Text)
	if err != nil {
		*reply = err.Error()
	}
	return nil
}

// ChannelRPCClient is the client-side RPC implementation
type ChannelRPCClient struct {
	client  *rpc.Client
	handler func(msg InboundMessage)
}

func (c *ChannelRPCClient) ID() string {
	var resp string
	_ = c.client.Call("Plugin.ID", struct{}{}, &resp)
	return resp
}

func (c *ChannelRPCClient) Connect(ctx context.Context, config map[string]string) error {
	var reply string
	err := c.client.Call("Plugin.Connect", ConnectArgs{Config: config}, &reply)
	if err != nil {
		return err
	}
	if reply != "" {
		return &PluginError{Message: reply}
	}
	return nil
}

func (c *ChannelRPCClient) Disconnect(ctx context.Context) error {
	var reply string
	err := c.client.Call("Plugin.Disconnect", struct{}{}, &reply)
	if err != nil {
		return err
	}
	if reply != "" {
		return &PluginError{Message: reply}
	}
	return nil
}

func (c *ChannelRPCClient) Send(ctx context.Context, channelID, text string) error {
	var reply string
	err := c.client.Call("Plugin.Send", SendArgs{ChannelID: channelID, Text: text}, &reply)
	if err != nil {
		return err
	}
	if reply != "" {
		return &PluginError{Message: reply}
	}
	return nil
}

func (c *ChannelRPCClient) SetHandler(fn func(msg InboundMessage)) {
	c.handler = fn
}

// =============================================================================
// Errors
// =============================================================================

// PluginError wraps errors from plugins
type PluginError struct {
	Message string
}

func (e *PluginError) Error() string {
	return e.Message
}
