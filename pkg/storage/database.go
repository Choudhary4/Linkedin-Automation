package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// DB represents the database connection
type DB struct {
	conn *sql.DB
}

// ConnectionRequest represents a connection request record
type ConnectionRequest struct {
	ID          int
	ProfileURL  string
	ProfileName string
	Company     string
	Message     string
	SentAt      time.Time
	Status      string // pending, accepted, rejected
}

// Message represents a sent message record
type Message struct {
	ID         int
	ProfileURL string
	ProfileName string
	Content    string
	SentAt     time.Time
}

// SearchHistory represents a search history record
type SearchHistory struct {
	ID          int
	Query       string
	Location    string
	ResultCount int
	SearchedAt  time.Time
}

// NewDB creates a new database connection
func NewDB(path string) (*DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{conn: conn}

	// Initialize schema
	if err := db.initSchema(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return db, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// initSchema creates the database tables
func (db *DB) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS connection_requests (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		profile_url TEXT NOT NULL UNIQUE,
		profile_name TEXT,
		company TEXT,
		message TEXT,
		sent_at DATETIME NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending'
	);

	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		profile_url TEXT NOT NULL,
		profile_name TEXT,
		content TEXT NOT NULL,
		sent_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS search_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		query TEXT NOT NULL,
		location TEXT,
		result_count INTEGER,
		searched_at DATETIME NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_connection_requests_sent_at ON connection_requests(sent_at);
	CREATE INDEX IF NOT EXISTS idx_connection_requests_status ON connection_requests(status);
	CREATE INDEX IF NOT EXISTS idx_messages_sent_at ON messages(sent_at);
	`

	_, err := db.conn.Exec(schema)
	return err
}

// SaveConnectionRequest saves a connection request to the database
func (db *DB) SaveConnectionRequest(req *ConnectionRequest) error {
	query := `
		INSERT INTO connection_requests (profile_url, profile_name, company, message, sent_at, status)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := db.conn.Exec(query, req.ProfileURL, req.ProfileName, req.Company, req.Message, req.SentAt, req.Status)
	if err != nil {
		return fmt.Errorf("failed to save connection request: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	req.ID = int(id)
	return nil
}

// GetConnectionRequestCount returns the number of connection requests sent in a time period
func (db *DB) GetConnectionRequestCount(since time.Time) (int, error) {
	query := `SELECT COUNT(*) FROM connection_requests WHERE sent_at >= ?`

	var count int
	err := db.conn.QueryRow(query, since).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get connection request count: %w", err)
	}

	return count, nil
}

// GetConnectionRequestCountToday returns the number of connection requests sent today
func (db *DB) GetConnectionRequestCountToday() (int, error) {
	today := time.Now().Truncate(24 * time.Hour)
	return db.GetConnectionRequestCount(today)
}

// GetConnectionRequestCountThisHour returns the number of connection requests sent this hour
func (db *DB) GetConnectionRequestCountThisHour() (int, error) {
	thisHour := time.Now().Truncate(time.Hour)
	return db.GetConnectionRequestCount(thisHour)
}

// HasSentConnectionRequest checks if a connection request has been sent to a profile
func (db *DB) HasSentConnectionRequest(profileURL string) (bool, error) {
	query := `SELECT COUNT(*) FROM connection_requests WHERE profile_url = ?`

	var count int
	err := db.conn.QueryRow(query, profileURL).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check connection request: %w", err)
	}

	return count > 0, nil
}

// UpdateConnectionRequestStatus updates the status of a connection request
func (db *DB) UpdateConnectionRequestStatus(profileURL, status string) error {
	query := `UPDATE connection_requests SET status = ? WHERE profile_url = ?`

	_, err := db.conn.Exec(query, status, profileURL)
	if err != nil {
		return fmt.Errorf("failed to update connection request status: %w", err)
	}

	return nil
}

// SaveMessage saves a message to the database
func (db *DB) SaveMessage(msg *Message) error {
	query := `
		INSERT INTO messages (profile_url, profile_name, content, sent_at)
		VALUES (?, ?, ?, ?)
	`

	result, err := db.conn.Exec(query, msg.ProfileURL, msg.ProfileName, msg.Content, msg.SentAt)
	if err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	msg.ID = int(id)
	return nil
}

// GetMessageCountToday returns the number of messages sent today
func (db *DB) GetMessageCountToday() (int, error) {
	today := time.Now().Truncate(24 * time.Hour)
	query := `SELECT COUNT(*) FROM messages WHERE sent_at >= ?`

	var count int
	err := db.conn.QueryRow(query, today).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get message count: %w", err)
	}

	return count, nil
}

// HasSentMessage checks if a message has been sent to a profile
func (db *DB) HasSentMessage(profileURL string) (bool, error) {
	query := `SELECT COUNT(*) FROM messages WHERE profile_url = ?`

	var count int
	err := db.conn.QueryRow(query, profileURL).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check message: %w", err)
	}

	return count > 0, nil
}

// SaveSearchHistory saves search history to the database
func (db *DB) SaveSearchHistory(search *SearchHistory) error {
	query := `
		INSERT INTO search_history (query, location, result_count, searched_at)
		VALUES (?, ?, ?, ?)
	`

	result, err := db.conn.Exec(query, search.Query, search.Location, search.ResultCount, search.SearchedAt)
	if err != nil {
		return fmt.Errorf("failed to save search history: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	search.ID = int(id)
	return nil
}

// GetRecentConnectionRequests returns recent connection requests
func (db *DB) GetRecentConnectionRequests(limit int) ([]*ConnectionRequest, error) {
	query := `
		SELECT id, profile_url, profile_name, company, message, sent_at, status
		FROM connection_requests
		ORDER BY sent_at DESC
		LIMIT ?
	`

	rows, err := db.conn.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent connection requests: %w", err)
	}
	defer rows.Close()

	var requests []*ConnectionRequest
	for rows.Next() {
		var req ConnectionRequest
		err := rows.Scan(&req.ID, &req.ProfileURL, &req.ProfileName, &req.Company, &req.Message, &req.SentAt, &req.Status)
		if err != nil {
			return nil, err
		}
		requests = append(requests, &req)
	}

	return requests, nil
}

// GetPendingConnectionRequests returns all pending connection requests
func (db *DB) GetPendingConnectionRequests() ([]*ConnectionRequest, error) {
	query := `
		SELECT id, profile_url, profile_name, company, message, sent_at, status
		FROM connection_requests
		WHERE status = 'pending'
		ORDER BY sent_at DESC
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending connection requests: %w", err)
	}
	defer rows.Close()

	var requests []*ConnectionRequest
	for rows.Next() {
		var req ConnectionRequest
		err := rows.Scan(&req.ID, &req.ProfileURL, &req.ProfileName, &req.Company, &req.Message, &req.SentAt, &req.Status)
		if err != nil {
			return nil, err
		}
		requests = append(requests, &req)
	}

	return requests, nil
}
