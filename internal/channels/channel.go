package channels

import (
	"context"
)

// Channel represents a messaging channel adapter
type Channel interface {
	// ID returns the unique identifier for this channel type
	ID() string

	// Connect establishes connection to the channel
	Connect(ctx context.Context, cfg ChannelConfig) error

	// Disconnect closes the connection to the channel
	Disconnect() error

	// Send sends a message to the channel
	Send(ctx context.Context, msg OutboundMessage) error

	// SetHandler sets the callback for incoming messages
	SetHandler(fn func(InboundMessage))
}

// ChannelConfig holds configuration for a channel
type ChannelConfig struct {
	// Common fields
	Token string `json:"token"` // Bot token or API key
	OrgID string `json:"org_id"`

	// Telegram-specific
	TelegramBotUsername string `json:"telegram_bot_username,omitempty"`

	// Discord-specific
	DiscordGuildID string `json:"discord_guild_id,omitempty"`

	// Slack-specific
	SlackBotID  string `json:"slack_bot_id,omitempty"`
	SlackTeamID string `json:"slack_team_id,omitempty"`
}

// InboundMessage represents a message received from a channel
type InboundMessage struct {
	// Channel info
	ChannelType string `json:"channel_type"` // telegram, discord, slack
	ChannelID   string `json:"channel_id"`   // Chat ID, channel ID, etc.

	// Message content
	MessageID string `json:"message_id"`
	Text      string `json:"text"`

	// Sender info
	SenderID   string `json:"sender_id"`
	SenderName string `json:"sender_name"`

	// Optional fields
	ReplyToID string `json:"reply_to_id,omitempty"`
	ThreadID  string `json:"thread_id,omitempty"`

	// Raw message for channel-specific handling
	Raw any `json:"-"`
}

// OutboundMessage represents a message to send to a channel
type OutboundMessage struct {
	// Target
	ChannelID string `json:"channel_id"`

	// Content
	Text string `json:"text"`

	// Optional fields
	ReplyToID string `json:"reply_to_id,omitempty"`
	ThreadID  string `json:"thread_id,omitempty"`

	// Formatting
	ParseMode string `json:"parse_mode,omitempty"` // markdown, html
}

// Manager manages multiple channel connections
type Manager struct {
	channels map[string]Channel
}

// NewManager creates a new channel manager
func NewManager() *Manager {
	return &Manager{
		channels: make(map[string]Channel),
	}
}

// Register adds a channel to the manager
func (m *Manager) Register(channel Channel) {
	m.channels[channel.ID()] = channel
}

// Get returns a channel by ID
func (m *Manager) Get(id string) (Channel, bool) {
	ch, ok := m.channels[id]
	return ch, ok
}

// ConnectAll connects all registered channels
func (m *Manager) ConnectAll(ctx context.Context) error {
	for _, ch := range m.channels {
		// Channels will need their own config, skip for now
		_ = ch
	}
	return nil
}

// DisconnectAll disconnects all channels
func (m *Manager) DisconnectAll() {
	for _, ch := range m.channels {
		ch.Disconnect()
	}
}

// List returns all registered channel IDs
func (m *Manager) List() []string {
	ids := make([]string, 0, len(m.channels))
	for id := range m.channels {
		ids = append(ids, id)
	}
	return ids
}
