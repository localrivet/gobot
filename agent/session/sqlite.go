package session

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Message represents a conversation message
type Message struct {
	ID          int64           `json:"id,omitempty"`
	SessionID   string          `json:"session_id"`
	Role        string          `json:"role"` // user, assistant, system, tool
	Content     string          `json:"content,omitempty"`
	ToolCalls   json.RawMessage `json:"tool_calls,omitempty"`
	ToolResults json.RawMessage `json:"tool_results,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

// ToolCall represents a tool invocation
type ToolCall struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	ToolCallID string `json:"tool_call_id"`
	Content    string `json:"content"`
	IsError    bool   `json:"is_error,omitempty"`
}

// Session represents a conversation session
type Session struct {
	ID         string    `json:"id"`
	SessionKey string    `json:"session_key"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Manager handles session persistence
type Manager struct {
	db *sql.DB
}

// New creates a new session manager
func New(dbPath string) (*Manager, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	m := &Manager{db: db}
	if err := m.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return m, nil
}

// migrate creates the required tables
func (m *Manager) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		session_key TEXT NOT NULL UNIQUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT NOT NULL,
		role TEXT NOT NULL,
		content TEXT,
		tool_calls TEXT,
		tool_results TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (session_id) REFERENCES sessions(id)
	);

	CREATE INDEX IF NOT EXISTS idx_messages_session ON messages(session_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_key ON sessions(session_key);
	`

	_, err := m.db.Exec(schema)
	return err
}

// Close closes the database connection
func (m *Manager) Close() error {
	return m.db.Close()
}

// GetOrCreate returns an existing session or creates a new one
func (m *Manager) GetOrCreate(sessionKey string) (*Session, error) {
	// Try to get existing session
	session, err := m.getByKey(sessionKey)
	if err == nil {
		return session, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	// Create new session
	id := generateID()
	now := time.Now()
	_, err = m.db.Exec(
		"INSERT INTO sessions (id, session_key, created_at, updated_at) VALUES (?, ?, ?, ?)",
		id, sessionKey, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &Session{
		ID:         id,
		SessionKey: sessionKey,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

// getByKey retrieves a session by its key
func (m *Manager) getByKey(sessionKey string) (*Session, error) {
	var s Session
	err := m.db.QueryRow(
		"SELECT id, session_key, created_at, updated_at FROM sessions WHERE session_key = ?",
		sessionKey,
	).Scan(&s.ID, &s.SessionKey, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// GetMessages retrieves messages for a session with an optional limit
func (m *Manager) GetMessages(sessionID string, limit int) ([]Message, error) {
	query := `
		SELECT id, session_id, role, content, tool_calls, tool_results, created_at
		FROM messages
		WHERE session_id = ?
		ORDER BY created_at ASC
	`
	if limit > 0 {
		// Get the last N messages
		query = `
			SELECT id, session_id, role, content, tool_calls, tool_results, created_at
			FROM (
				SELECT * FROM messages
				WHERE session_id = ?
				ORDER BY created_at DESC
				LIMIT ?
			) ORDER BY created_at ASC
		`
	}

	var rows *sql.Rows
	var err error
	if limit > 0 {
		rows, err = m.db.Query(query, sessionID, limit)
	} else {
		rows, err = m.db.Query(query, sessionID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		var toolCalls, toolResults sql.NullString
		err := rows.Scan(
			&msg.ID, &msg.SessionID, &msg.Role, &msg.Content,
			&toolCalls, &toolResults, &msg.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		if toolCalls.Valid {
			msg.ToolCalls = json.RawMessage(toolCalls.String)
		}
		if toolResults.Valid {
			msg.ToolResults = json.RawMessage(toolResults.String)
		}
		messages = append(messages, msg)
	}

	return messages, rows.Err()
}

// AppendMessage adds a message to a session
func (m *Manager) AppendMessage(sessionID string, msg Message) error {
	var toolCalls, toolResults sql.NullString
	if len(msg.ToolCalls) > 0 {
		toolCalls = sql.NullString{String: string(msg.ToolCalls), Valid: true}
	}
	if len(msg.ToolResults) > 0 {
		toolResults = sql.NullString{String: string(msg.ToolResults), Valid: true}
	}

	_, err := m.db.Exec(
		"INSERT INTO messages (session_id, role, content, tool_calls, tool_results, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		sessionID, msg.Role, msg.Content, toolCalls, toolResults, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to append message: %w", err)
	}

	// Update session timestamp
	_, err = m.db.Exec(
		"UPDATE sessions SET updated_at = ? WHERE id = ?",
		time.Now(), sessionID,
	)
	return err
}

// Compact summarizes old messages to reduce context size
// This should be called when context window is getting full
func (m *Manager) Compact(sessionID string, summaryContent string) error {
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get message count
	var count int
	err = tx.QueryRow("SELECT COUNT(*) FROM messages WHERE session_id = ?", sessionID).Scan(&count)
	if err != nil {
		return err
	}

	// Keep the last 10 messages, summarize the rest
	keepCount := 10
	if count <= keepCount {
		return nil // Nothing to compact
	}

	// Delete old messages (keeping the newest ones)
	_, err = tx.Exec(`
		DELETE FROM messages
		WHERE session_id = ? AND id NOT IN (
			SELECT id FROM messages
			WHERE session_id = ?
			ORDER BY created_at DESC
			LIMIT ?
		)
	`, sessionID, sessionID, keepCount)
	if err != nil {
		return err
	}

	// Insert summary as a system message at the beginning
	_, err = tx.Exec(
		"INSERT INTO messages (session_id, role, content, created_at) VALUES (?, 'system', ?, ?)",
		sessionID, summaryContent, time.Now().Add(-time.Hour), // Slightly in the past
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Reset clears all messages from a session
func (m *Manager) Reset(sessionID string) error {
	_, err := m.db.Exec("DELETE FROM messages WHERE session_id = ?", sessionID)
	return err
}

// ListSessions returns all sessions
func (m *Manager) ListSessions() ([]Session, error) {
	rows, err := m.db.Query(
		"SELECT id, session_key, created_at, updated_at FROM sessions ORDER BY updated_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var s Session
		if err := rows.Scan(&s.ID, &s.SessionKey, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

// DeleteSession removes a session and all its messages
func (m *Manager) DeleteSession(sessionID string) error {
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM messages WHERE session_id = ?", sessionID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM sessions WHERE id = ?", sessionID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// generateID creates a simple unique ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
