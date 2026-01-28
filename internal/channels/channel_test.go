package channels

import (
	"testing"
)

func TestChannelConfig(t *testing.T) {
	cfg := ChannelConfig{
		Token: "test-token",
		OrgID: "test-org",
		TelegramBotUsername: "test_bot",
	}

	if cfg.Token != "test-token" {
		t.Errorf("expected token test-token, got %s", cfg.Token)
	}
	if cfg.OrgID != "test-org" {
		t.Errorf("expected org ID test-org, got %s", cfg.OrgID)
	}
}

func TestInboundMessage(t *testing.T) {
	msg := InboundMessage{
		ChannelType: "telegram",
		ChannelID:   "12345",
		MessageID:   "msg-1",
		Text:        "Hello",
		SenderID:    "user-1",
		SenderName:  "Test User",
	}

	if msg.ChannelType != "telegram" {
		t.Errorf("expected channel type telegram, got %s", msg.ChannelType)
	}
	if msg.Text != "Hello" {
		t.Errorf("expected text Hello, got %s", msg.Text)
	}
}

func TestOutboundMessage(t *testing.T) {
	msg := OutboundMessage{
		ChannelID: "12345",
		Text:      "Response",
		ParseMode: "markdown",
	}

	if msg.ChannelID != "12345" {
		t.Errorf("expected channel ID 12345, got %s", msg.ChannelID)
	}
	if msg.ParseMode != "markdown" {
		t.Errorf("expected parse mode markdown, got %s", msg.ParseMode)
	}
}

func TestManager(t *testing.T) {
	manager := NewManager()

	ids := manager.List()
	if len(ids) != 0 {
		t.Errorf("expected 0 channels, got %d", len(ids))
	}
}
