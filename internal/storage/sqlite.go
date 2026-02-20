// Package storage provides SQLite storage implementation
package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite" // Use modernc.org/sqlite for pure Go SQLite (CGO-free)
)

// SQLiteStore implements Store interface with SQLite backend
type SQLiteStore struct {
	db   *sql.DB
	path string
	mu   sync.RWMutex
}

// NewSQLiteStore creates a new SQLite store
func NewSQLiteStore(config *SQLiteConfig) (*SQLiteStore, error) {
	if config == nil {
		return nil, ErrMissingSQLiteConfig
	}

	// Ensure directory exists
	dir := filepath.Dir(config.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Open database connection
	db, err := sql.Open("sqlite", config.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	store := &SQLiteStore{
		db:   db,
		path: config.Path,
	}

	// Initialize schema
	if err := store.initSchema(config); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return store, nil
}

// initSchema creates the database schema
func (s *SQLiteStore) initSchema(config *SQLiteConfig) error {
	schema := `
	CREATE TABLE IF NOT EXISTS conversations (
		id TEXT PRIMARY KEY,
		model TEXT NOT NULL,
		title TEXT,
		system_prompt TEXT,
		message_count INTEGER DEFAULT 0,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		metadata TEXT
	);

	CREATE TABLE IF NOT EXISTS messages (
		id TEXT PRIMARY KEY,
		conversation_id TEXT NOT NULL,
		role TEXT NOT NULL,
		content TEXT NOT NULL,
		name TEXT,
		token_count INTEGER DEFAULT 0,
		created_at INTEGER NOT NULL,
		metadata TEXT,
		FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_messages_conversation ON messages(conversation_id);
	CREATE INDEX IF NOT EXISTS idx_conversations_created ON conversations(created_at);
	CREATE INDEX IF NOT EXISTS idx_conversations_updated ON conversations(updated_at);
	`

	if _, err := s.db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	// Apply pragmas
	pragmas := []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA journal_mode = WAL",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA cache_size = -64000", // 64MB cache
		"PRAGMA temp_store = memory",
	}

	// Apply custom pragmas from config
	if config.Pragmas != nil {
		for key, value := range config.Pragmas {
			pragmas = append(pragmas, fmt.Sprintf("PRAGMA %s = %s", key, value))
		}
	}

	for _, pragma := range pragmas {
		if _, err := s.db.Exec(pragma); err != nil {
			return fmt.Errorf("failed to set pragma %s: %w", pragma, err)
		}
	}

	return nil
}

// CreateConversation creates a new conversation
func (s *SQLiteStore) CreateConversation(ctx context.Context, conv *Conversation) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if conv.ID == "" {
		conv.ID = generateID("conv")
	}

	if conv.CreatedAt.IsZero() {
		conv.CreatedAt = timeNow()
	}

	if conv.UpdatedAt.IsZero() {
		conv.UpdatedAt = timeNow()
	}

	metadataJSON, _ := json.Marshal(conv.Metadata)

	query := `
	INSERT INTO conversations (id, model, title, system_prompt, message_count, created_at, updated_at, metadata)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query,
		conv.ID,
		conv.Model,
		conv.Title,
		conv.SystemPrompt,
		conv.MessageCount,
		conv.CreatedAt.Unix(),
		conv.UpdatedAt.Unix(),
		string(metadataJSON),
	)

	return err
}

// GetConversation retrieves a conversation by ID
func (s *SQLiteStore) GetConversation(ctx context.Context, id string) (*Conversation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `
	SELECT id, model, title, system_prompt, message_count, created_at, updated_at, metadata
	FROM conversations
	WHERE id = ?
	`

	var metadataJSON []byte
	var createdUnix, updatedUnix int64
	conv := &Conversation{}

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&conv.ID,
		&conv.Model,
		&conv.Title,
		&conv.SystemPrompt,
		&conv.MessageCount,
		&createdUnix,
		&updatedUnix,
		&metadataJSON,
	)

	if err == sql.ErrNoRows {
		return nil, ErrConversationNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	// Parse metadata JSON
	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &conv.Metadata)
	}

	// Convert Unix timestamps to time.Time
	conv.CreatedAt = time.Unix(createdUnix, 0).UTC()
	conv.UpdatedAt = time.Unix(updatedUnix, 0).UTC()

	return conv, nil
}

// ListConversations lists all conversations
func (s *SQLiteStore) ListConversations(ctx context.Context, limit, offset int) ([]*Conversation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `
	SELECT id, model, title, system_prompt, message_count, created_at, updated_at, metadata
	FROM conversations
	ORDER BY updated_at DESC
	LIMIT ? OFFSET ?
	`

	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list conversations: %w", err)
	}
	defer rows.Close()

	convs := []*Conversation{}

	for rows.Next() {
		var metadataJSON []byte
		var createdUnix, updatedUnix int64
		conv := &Conversation{}

		err := rows.Scan(
			&conv.ID,
			&conv.Model,
			&conv.Title,
			&conv.SystemPrompt,
			&conv.MessageCount,
			&createdUnix,
			&updatedUnix,
			&metadataJSON,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan conversation: %w", err)
		}

		// Parse metadata JSON
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &conv.Metadata)
		}

		// Convert Unix timestamps
		conv.CreatedAt = time.Unix(createdUnix, 0).UTC()
		conv.UpdatedAt = time.Unix(updatedUnix, 0).UTC()

		convs = append(convs, conv)
	}

	return convs, nil
}

// UpdateConversation updates an existing conversation
func (s *SQLiteStore) UpdateConversation(ctx context.Context, conv *Conversation) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	conv.UpdatedAt = timeNow()

	metadataJSON, _ := json.Marshal(conv.Metadata)

	query := `
	UPDATE conversations
	SET model = ?, title = ?, system_prompt = ?, message_count = ?, updated_at = ?, metadata = ?
	WHERE id = ?
	`

	result, err := s.db.ExecContext(ctx, query,
		conv.Model,
		conv.Title,
		conv.SystemPrompt,
		conv.MessageCount,
		conv.UpdatedAt.Unix(),
		string(metadataJSON),
		conv.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update conversation: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrConversationNotFound
	}

	return nil
}

// DeleteConversation deletes a conversation and its messages
func (s *SQLiteStore) DeleteConversation(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	result, err := s.db.ExecContext(ctx, "DELETE FROM conversations WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete conversation: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrConversationNotFound
	}

	return nil
}

// CreateMessage creates a new message
func (s *SQLiteStore) CreateMessage(ctx context.Context, msg *Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if msg.ID == "" {
		msg.ID = generateID("msg")
	}

	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = timeNow()
	}

	metadataJSON, _ := json.Marshal(msg.Metadata)

	query := `
	INSERT INTO messages (id, conversation_id, role, content, name, token_count, created_at, metadata)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query,
		msg.ID,
		msg.ConversationID,
		msg.Role,
		msg.Content,
		msg.Name,
		msg.TokenCount,
		msg.CreatedAt.Unix(),
		string(metadataJSON),
	)

	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	// Update conversation message count and timestamp
	_, err = s.db.ExecContext(ctx,
		"UPDATE conversations SET message_count = message_count + 1, updated_at = ? WHERE id = ?",
		timeNow().Unix(), msg.ConversationID,
	)

	return err
}

// GetMessages retrieves messages for a conversation
func (s *SQLiteStore) GetMessages(ctx context.Context, conversationID string, limit, offset int) ([]*Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `
	SELECT id, conversation_id, role, content, name, token_count, created_at, metadata
	FROM messages
	WHERE conversation_id = ?
	ORDER BY created_at ASC
	LIMIT ? OFFSET ?
	`

	rows, err := s.db.QueryContext(ctx, query, conversationID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	defer rows.Close()

	messages := []*Message{}

	for rows.Next() {
		var metadataJSON []byte
		var createdUnix int64
		msg := &Message{}

		err := rows.Scan(
			&msg.ID,
			&msg.ConversationID,
			&msg.Role,
			&msg.Content,
			&msg.Name,
			&msg.TokenCount,
			&createdUnix,
			&metadataJSON,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		// Parse metadata JSON
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &msg.Metadata)
		}

		// Convert Unix timestamp
		msg.CreatedAt = time.Unix(createdUnix, 0).UTC()

		messages = append(messages, msg)
	}

	return messages, nil
}

// DeleteMessages deletes all messages for a conversation
func (s *SQLiteStore) DeleteMessages(ctx context.Context, conversationID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.ExecContext(ctx, "DELETE FROM messages WHERE conversation_id = ?", conversationID)
	if err != nil {
		return fmt.Errorf("failed to delete messages: %w", err)
	}

	// Reset conversation message count
	_, err = s.db.ExecContext(ctx,
		"UPDATE conversations SET message_count = 0, updated_at = ? WHERE id = ?",
		timeNow().Unix(), conversationID,
	)

	return err
}

// Close closes the database connection
func (s *SQLiteStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.db.Close()
}

// Stats returns statistics about the database
func (s *SQLiteStore) Stats() (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := make(map[string]interface{})

	// Get table counts
	var convCount, msgCount int64

	s.db.QueryRow("SELECT COUNT(*) FROM conversations").Scan(&convCount)
	s.db.QueryRow("SELECT COUNT(*) FROM messages").Scan(&msgCount)

	stats["conversations"] = convCount
	stats["messages"] = msgCount
	stats["type"] = "sqlite"
	stats["path"] = s.path

	// Get database size
	if info, err := os.Stat(s.path); err == nil {
		stats["size_bytes"] = info.Size()
	}

	return stats, nil
}

// Helper functions for time handling

func timeNow() time.Time {
	return time.Now().UTC()
}
