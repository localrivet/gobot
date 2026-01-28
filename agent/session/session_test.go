package session

import (
	"path/filepath"
	"testing"
)

func TestSessionManager(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create manager
	manager, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	defer manager.Close()

	// Test GetOrCreate
	sess, err := manager.GetOrCreate("test-session")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	if sess.SessionKey != "test-session" {
		t.Errorf("expected session key 'test-session', got %q", sess.SessionKey)
	}

	// Test getting the same session
	sess2, err := manager.GetOrCreate("test-session")
	if err != nil {
		t.Fatalf("failed to get session: %v", err)
	}

	if sess.ID != sess2.ID {
		t.Error("expected same session ID")
	}

	// Test AppendMessage
	err = manager.AppendMessage(sess.ID, Message{
		SessionID: sess.ID,
		Role:      "user",
		Content:   "hello",
	})
	if err != nil {
		t.Fatalf("failed to append message: %v", err)
	}

	// Test GetMessages
	messages, err := manager.GetMessages(sess.ID, 0)
	if err != nil {
		t.Fatalf("failed to get messages: %v", err)
	}

	if len(messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(messages))
	}

	if messages[0].Content != "hello" {
		t.Errorf("expected content 'hello', got %q", messages[0].Content)
	}

	// Test Reset
	err = manager.Reset(sess.ID)
	if err != nil {
		t.Fatalf("failed to reset session: %v", err)
	}

	messages, _ = manager.GetMessages(sess.ID, 0)
	if len(messages) != 0 {
		t.Errorf("expected 0 messages after reset, got %d", len(messages))
	}
}

func TestSessionManagerWithLimit(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	manager, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	defer manager.Close()

	sess, _ := manager.GetOrCreate("limit-test")

	// Add 10 messages
	for i := 0; i < 10; i++ {
		manager.AppendMessage(sess.ID, Message{
			SessionID: sess.ID,
			Role:      "user",
			Content:   "message",
		})
	}

	// Get with limit of 5
	messages, _ := manager.GetMessages(sess.ID, 5)
	if len(messages) != 5 {
		t.Errorf("expected 5 messages with limit, got %d", len(messages))
	}
}

func TestListSessions(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	manager, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	defer manager.Close()

	// Create some sessions
	manager.GetOrCreate("session-1")
	manager.GetOrCreate("session-2")
	manager.GetOrCreate("session-3")

	sessions, err := manager.ListSessions()
	if err != nil {
		t.Fatalf("failed to list sessions: %v", err)
	}

	if len(sessions) != 3 {
		t.Errorf("expected 3 sessions, got %d", len(sessions))
	}
}
