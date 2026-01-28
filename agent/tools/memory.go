package tools

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// MemoryTool provides persistent fact storage across sessions
type MemoryTool struct {
	db     *sql.DB
	dbPath string
}

type memoryInput struct {
	Action    string            `json:"action"`    // store, recall, search, list, delete, clear
	Key       string            `json:"key"`       // Fact key/identifier
	Value     string            `json:"value"`     // Fact content (for store)
	Tags      []string          `json:"tags"`      // Tags for categorization
	Query     string            `json:"query"`     // Search query (for search action)
	Namespace string            `json:"namespace"` // Namespace for organization (default: "default")
	Metadata  map[string]string `json:"metadata"`  // Additional metadata
}

// MemoryConfig configures the memory tool
type MemoryConfig struct {
	DBPath string // Path to memory database (default: ~/.gobot/memory.db)
}

func NewMemoryTool(cfg MemoryConfig) (*MemoryTool, error) {
	if cfg.DBPath == "" {
		homeDir, _ := os.UserHomeDir()
		cfg.DBPath = filepath.Join(homeDir, ".gobot", "memory.db")
	}

	// Ensure directory exists
	dir := filepath.Dir(cfg.DBPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create memory directory: %w", err)
	}

	db, err := sql.Open("sqlite", cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open memory database: %w", err)
	}

	tool := &MemoryTool{
		db:     db,
		dbPath: cfg.DBPath,
	}

	if err := tool.initSchema(); err != nil {
		db.Close()
		return nil, err
	}

	return tool, nil
}

func (t *MemoryTool) initSchema() error {
	schema := `
		CREATE TABLE IF NOT EXISTS memories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			namespace TEXT NOT NULL DEFAULT 'default',
			key TEXT NOT NULL,
			value TEXT NOT NULL,
			tags TEXT,
			metadata TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			accessed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			access_count INTEGER DEFAULT 0,
			UNIQUE(namespace, key)
		);

		CREATE INDEX IF NOT EXISTS idx_memories_namespace ON memories(namespace);
		CREATE INDEX IF NOT EXISTS idx_memories_key ON memories(key);
		CREATE INDEX IF NOT EXISTS idx_memories_tags ON memories(tags);

		CREATE VIRTUAL TABLE IF NOT EXISTS memories_fts USING fts5(
			key, value, tags,
			content='memories',
			content_rowid='id'
		);

		CREATE TRIGGER IF NOT EXISTS memories_ai AFTER INSERT ON memories BEGIN
			INSERT INTO memories_fts(rowid, key, value, tags)
			VALUES (new.id, new.key, new.value, new.tags);
		END;

		CREATE TRIGGER IF NOT EXISTS memories_ad AFTER DELETE ON memories BEGIN
			INSERT INTO memories_fts(memories_fts, rowid, key, value, tags)
			VALUES ('delete', old.id, old.key, old.value, old.tags);
		END;

		CREATE TRIGGER IF NOT EXISTS memories_au AFTER UPDATE ON memories BEGIN
			INSERT INTO memories_fts(memories_fts, rowid, key, value, tags)
			VALUES ('delete', old.id, old.key, old.value, old.tags);
			INSERT INTO memories_fts(rowid, key, value, tags)
			VALUES (new.id, new.key, new.value, new.tags);
		END;
	`
	_, err := t.db.Exec(schema)
	return err
}

func (t *MemoryTool) Close() error {
	if t.db != nil {
		return t.db.Close()
	}
	return nil
}

func (t *MemoryTool) Name() string {
	return "memory"
}

func (t *MemoryTool) Description() string {
	return "Store and recall facts persistently across sessions. Use for remembering user preferences, project context, learned information, and important notes."
}

func (t *MemoryTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"action": {
				"type": "string",
				"enum": ["store", "recall", "search", "list", "delete", "clear"],
				"description": "Memory action: store (save fact), recall (get by key), search (full-text search), list (list keys), delete (remove fact), clear (remove all in namespace)"
			},
			"key": {
				"type": "string",
				"description": "Unique identifier for the fact (required for store, recall, delete)"
			},
			"value": {
				"type": "string",
				"description": "The fact content to store (required for store action)"
			},
			"tags": {
				"type": "array",
				"items": {"type": "string"},
				"description": "Tags for categorization (e.g., ['preference', 'user'])"
			},
			"query": {
				"type": "string",
				"description": "Search query for full-text search (required for search action)"
			},
			"namespace": {
				"type": "string",
				"description": "Namespace for organization (default: 'default'). Use different namespaces for different projects/contexts.",
				"default": "default"
			},
			"metadata": {
				"type": "object",
				"additionalProperties": {"type": "string"},
				"description": "Additional metadata as key-value pairs"
			}
		},
		"required": ["action"]
	}`)
}

func (t *MemoryTool) RequiresApproval() bool {
	return false
}

func (t *MemoryTool) Execute(ctx context.Context, input json.RawMessage) (*ToolResult, error) {
	var params memoryInput
	if err := json.Unmarshal(input, &params); err != nil {
		return &ToolResult{
			Content: fmt.Sprintf("Failed to parse input: %v", err),
			IsError: true,
		}, nil
	}

	if params.Namespace == "" {
		params.Namespace = "default"
	}

	var result string
	var err error

	switch params.Action {
	case "store":
		result, err = t.store(params)
	case "recall":
		result, err = t.recall(params)
	case "search":
		result, err = t.search(params)
	case "list":
		result, err = t.list(params)
	case "delete":
		result, err = t.delete(params)
	case "clear":
		result, err = t.clear(params)
	default:
		return &ToolResult{
			Content: fmt.Sprintf("Unknown action: %s", params.Action),
			IsError: true,
		}, nil
	}

	if err != nil {
		return &ToolResult{
			Content: fmt.Sprintf("Memory action failed: %v", err),
			IsError: true,
		}, nil
	}

	return &ToolResult{
		Content: result,
		IsError: false,
	}, nil
}

func (t *MemoryTool) store(params memoryInput) (string, error) {
	if params.Key == "" {
		return "", fmt.Errorf("key is required for store action")
	}
	if params.Value == "" {
		return "", fmt.Errorf("value is required for store action")
	}

	tagsJSON, _ := json.Marshal(params.Tags)
	metadataJSON, _ := json.Marshal(params.Metadata)

	// Upsert
	query := `
		INSERT INTO memories (namespace, key, value, tags, metadata, updated_at)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(namespace, key) DO UPDATE SET
			value = excluded.value,
			tags = excluded.tags,
			metadata = excluded.metadata,
			updated_at = CURRENT_TIMESTAMP
	`
	_, err := t.db.Exec(query, params.Namespace, params.Key, params.Value, string(tagsJSON), string(metadataJSON))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Stored memory: %s (namespace: %s)", params.Key, params.Namespace), nil
}

func (t *MemoryTool) recall(params memoryInput) (string, error) {
	if params.Key == "" {
		return "", fmt.Errorf("key is required for recall action")
	}

	var value, tags, metadata string
	var createdAt, updatedAt, accessedAt time.Time
	var accessCount int

	query := `
		SELECT value, tags, metadata, created_at, updated_at, accessed_at, access_count
		FROM memories
		WHERE namespace = ? AND key = ?
	`
	err := t.db.QueryRow(query, params.Namespace, params.Key).Scan(
		&value, &tags, &metadata, &createdAt, &updatedAt, &accessedAt, &accessCount,
	)
	if err == sql.ErrNoRows {
		return fmt.Sprintf("No memory found with key '%s' in namespace '%s'", params.Key, params.Namespace), nil
	}
	if err != nil {
		return "", err
	}

	// Update access stats
	t.db.Exec(`
		UPDATE memories SET accessed_at = CURRENT_TIMESTAMP, access_count = access_count + 1
		WHERE namespace = ? AND key = ?
	`, params.Namespace, params.Key)

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Key: %s\n", params.Key))
	result.WriteString(fmt.Sprintf("Value: %s\n", value))
	if tags != "" && tags != "null" {
		result.WriteString(fmt.Sprintf("Tags: %s\n", tags))
	}
	if metadata != "" && metadata != "null" && metadata != "{}" {
		result.WriteString(fmt.Sprintf("Metadata: %s\n", metadata))
	}
	result.WriteString(fmt.Sprintf("Created: %s\n", createdAt.Format(time.RFC3339)))
	result.WriteString(fmt.Sprintf("Accessed: %d times", accessCount+1))

	return result.String(), nil
}

func (t *MemoryTool) search(params memoryInput) (string, error) {
	if params.Query == "" {
		return "", fmt.Errorf("query is required for search action")
	}

	query := `
		SELECT m.key, m.value, m.tags
		FROM memories m
		JOIN memories_fts f ON m.id = f.rowid
		WHERE memories_fts MATCH ? AND m.namespace = ?
		ORDER BY rank
		LIMIT 10
	`
	rows, err := t.db.Query(query, params.Query, params.Namespace)
	if err != nil {
		// Try simple LIKE search as fallback
		query = `
			SELECT key, value, tags
			FROM memories
			WHERE namespace = ? AND (key LIKE ? OR value LIKE ?)
			LIMIT 10
		`
		likePattern := "%" + params.Query + "%"
		rows, err = t.db.Query(query, params.Namespace, likePattern, likePattern)
		if err != nil {
			return "", err
		}
	}
	defer rows.Close()

	var results []string
	for rows.Next() {
		var key, value, tags string
		if err := rows.Scan(&key, &value, &tags); err != nil {
			continue
		}
		// Truncate value if too long
		if len(value) > 200 {
			value = value[:200] + "..."
		}
		results = append(results, fmt.Sprintf("- %s: %s", key, value))
	}

	if len(results) == 0 {
		return fmt.Sprintf("No memories found matching '%s' in namespace '%s'", params.Query, params.Namespace), nil
	}

	return fmt.Sprintf("Found %d memories:\n%s", len(results), strings.Join(results, "\n")), nil
}

func (t *MemoryTool) list(params memoryInput) (string, error) {
	query := `
		SELECT key, substr(value, 1, 100) as preview, tags, access_count
		FROM memories
		WHERE namespace = ?
		ORDER BY access_count DESC, updated_at DESC
		LIMIT 50
	`
	rows, err := t.db.Query(query, params.Namespace)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var results []string
	for rows.Next() {
		var key, preview, tags string
		var accessCount int
		if err := rows.Scan(&key, &preview, &tags, &accessCount); err != nil {
			continue
		}
		if len(preview) > 80 {
			preview = preview[:80] + "..."
		}
		results = append(results, fmt.Sprintf("- %s: %s (accessed %d times)", key, preview, accessCount))
	}

	if len(results) == 0 {
		return fmt.Sprintf("No memories in namespace '%s'", params.Namespace), nil
	}

	return fmt.Sprintf("Memories in namespace '%s' (%d items):\n%s", params.Namespace, len(results), strings.Join(results, "\n")), nil
}

func (t *MemoryTool) delete(params memoryInput) (string, error) {
	if params.Key == "" {
		return "", fmt.Errorf("key is required for delete action")
	}

	result, err := t.db.Exec(`DELETE FROM memories WHERE namespace = ? AND key = ?`, params.Namespace, params.Key)
	if err != nil {
		return "", err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Sprintf("No memory found with key '%s' in namespace '%s'", params.Key, params.Namespace), nil
	}

	return fmt.Sprintf("Deleted memory: %s (namespace: %s)", params.Key, params.Namespace), nil
}

func (t *MemoryTool) clear(params memoryInput) (string, error) {
	result, err := t.db.Exec(`DELETE FROM memories WHERE namespace = ?`, params.Namespace)
	if err != nil {
		return "", err
	}

	rows, _ := result.RowsAffected()
	return fmt.Sprintf("Cleared %d memories from namespace '%s'", rows, params.Namespace), nil
}
