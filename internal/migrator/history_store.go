package migrator

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
	// Drivers should be imported by the main package or migrator.go,
	// but we need them for sql.Open to work if they haven't been registered.
	// Since migrator.go imports the migrate drivers which likely import the sql drivers,
	// we should be okay. If not, we might need to add blank imports here.
)

const (
	historyTableName = "janus_migrations_history"
)

// HistoryEntry represents a single execution of a migration
type HistoryEntry struct {
	ID        int64
	Version   uint
	Action    string // "up" or "down"
	Name      string
	StartTime time.Time
	Duration  time.Duration
	Success   bool
	Error     string
}

// HistoryStore manages the history table
type HistoryStore struct {
	db     *sql.DB
	driver string
}

// NewHistoryStore creates a new HistoryStore given a database URL
func NewHistoryStore(databaseURL string) (*HistoryStore, error) {
	driver, dsn, err := parseURL(databaseURL)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return &HistoryStore{
		db:     db,
		driver: driver,
	}, nil
}

// Close closes the database connection
func (h *HistoryStore) Close() error {
	return h.db.Close()
}

// EnsureTable creates the history table if it doesn't exist
func (h *HistoryStore) EnsureTable() error {
	var query string

	switch h.driver {
	case "postgres":
		query = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			id SERIAL PRIMARY KEY,
			version BIGINT NOT NULL,
			action VARCHAR(10) NOT NULL,
			name VARCHAR(255) NOT NULL,
			start_time TIMESTAMP WITH TIME ZONE NOT NULL,
			duration_ns BIGINT NOT NULL,
			success BOOLEAN NOT NULL,
			error TEXT
		)`, historyTableName)
	case "mysql":
		query = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			version BIGINT UNSIGNED NOT NULL,
			action VARCHAR(10) NOT NULL,
			name VARCHAR(255) NOT NULL,
			start_time TIMESTAMP NOT NULL,
			duration_ns BIGINT NOT NULL,
			success BOOLEAN NOT NULL,
			error TEXT
		)`, historyTableName)
	case "sqlite3":
		query = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			version INTEGER NOT NULL,
			action TEXT NOT NULL,
			name TEXT NOT NULL,
			start_time DATETIME NOT NULL,
			duration_ns INTEGER NOT NULL,
			success BOOLEAN NOT NULL,
			error TEXT
		)`, historyTableName)
	default:
		return fmt.Errorf("unsupported driver for history: %s", h.driver)
	}

	_, err := h.db.Exec(query)
	if err != nil {
		return fmt.Errorf("create history table: %w", err)
	}

	return nil
}

// Record adds a new history entry
func (h *HistoryStore) Record(entry HistoryEntry) error {
	query := fmt.Sprintf(`INSERT INTO %s (version, action, name, start_time, duration_ns, success, error) VALUES (?, ?, ?, ?, ?, ?, ?)`, historyTableName)

	// Adjust placeholder for Postgres
	if h.driver == "postgres" {
		query = fmt.Sprintf(`INSERT INTO %s (version, action, name, start_time, duration_ns, success, error) VALUES ($1, $2, $3, $4, $5, $6, $7)`, historyTableName)
	}

	_, err := h.db.Exec(query,
		entry.Version,
		entry.Action,
		entry.Name,
		entry.StartTime,
		entry.Duration.Nanoseconds(),
		entry.Success,
		entry.Error,
	)

	if err != nil {
		return fmt.Errorf("record history: %w", err)
	}
	return nil
}

// GetHistory returns the execution history for a specific version
// If version is 0, returns all history
func (h *HistoryStore) GetHistory(version uint) ([]HistoryEntry, error) {
	query := fmt.Sprintf(`SELECT id, version, action, name, start_time, duration_ns, success, error FROM %s`, historyTableName)
	var args []interface{}

	if version > 0 {
		if h.driver == "postgres" {
			query += " WHERE version = $1"
		} else {
			query += " WHERE version = ?"
		}
		args = append(args, version)
	}

	query += " ORDER BY start_time DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		// If table doesn't exist, just return empty list
		if strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "no such table") {
			return []HistoryEntry{}, nil
		}
		return nil, fmt.Errorf("query history: %w", err)
	}
	defer rows.Close()

	var entries []HistoryEntry
	for rows.Next() {
		var e HistoryEntry
		var durationNs int64
		var errStr sql.NullString

		if err := rows.Scan(&e.ID, &e.Version, &e.Action, &e.Name, &e.StartTime, &durationNs, &e.Success, &errStr); err != nil {
			return nil, err
		}
		e.Duration = time.Duration(durationNs)
		if errStr.Valid {
			e.Error = errStr.String
		}
		entries = append(entries, e)
	}

	return entries, nil
}

// Helper to parse database URL into driver and DSN
func parseURL(url string) (string, string, error) {
	parts := strings.SplitN(url, "://", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid database URL format")
	}
	scheme := parts[0]
	dsn := parts[1]

	switch scheme {
	case "postgres", "postgresql":
		return "postgres", url, nil
	case "mysql":
		return "mysql", dsn, nil
	case "sqlite", "sqlite3":
		return "sqlite3", dsn, nil
	default:
		return "", "", fmt.Errorf("unsupported database scheme: %s", scheme)
	}
}
