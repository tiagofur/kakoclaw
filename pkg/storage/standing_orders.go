package storage

import (
	"fmt"
	"time"
)

// migrateStandingOrders creates the standing_orders table.
// Called from migrateUserDB().
func (s *Storage) migrateStandingOrders() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS standing_orders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			content TEXT NOT NULL,
			label TEXT DEFAULT '',
			active BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
	}

	// In shared (non-user) DB mode, we also need the user_id column.
	if !s.isUserDB {
		queries = []string{
			`CREATE TABLE IF NOT EXISTS standing_orders (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL DEFAULT 0,
				content TEXT NOT NULL,
				label TEXT DEFAULT '',
				active BOOLEAN DEFAULT 1,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
			);`,
			`CREATE INDEX IF NOT EXISTS idx_standing_orders_user_id ON standing_orders(user_id);`,
		}
	}

	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("standing orders migration: %w", err)
		}
	}
	return nil
}

// StandingOrder represents a persistent instruction for the agent.
type StandingOrder struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Label     string    `json:"label"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

const (
	maxStandingOrders       = 20
	maxStandingOrderContent = 200
	maxStandingOrdersTotal  = 4000
)

// CreateStandingOrder creates a new standing order with default user key (nil).
func (s *Storage) CreateStandingOrder(content, label string) (int64, error) {
	return s.CreateStandingOrderForUser(nil, content, label)
}

// CreateStandingOrderForUser creates a new standing order for a specific user.
func (s *Storage) CreateStandingOrderForUser(userKey interface{}, content, label string) (int64, error) {
	if len(content) > maxStandingOrderContent {
		return 0, fmt.Errorf("content exceeds %d character limit (got %d)", maxStandingOrderContent, len(content))
	}

	// Check active count and total chars limits
	existing, err := s.ListStandingOrdersForUser(userKey, true)
	if err != nil {
		return 0, fmt.Errorf("checking existing orders: %w", err)
	}

	if len(existing) >= maxStandingOrders {
		return 0, fmt.Errorf("maximum of %d active standing orders reached", maxStandingOrders)
	}

	totalChars := 0
	for _, o := range existing {
		totalChars += len(o.Content)
	}
	if totalChars+len(content) > maxStandingOrdersTotal {
		return 0, fmt.Errorf("total content of active standing orders would exceed %d character limit", maxStandingOrdersTotal)
	}

	var query string
	var args []interface{}

	if s.isUserDB {
		query = `INSERT INTO standing_orders (content, label) VALUES (?, ?)`
		args = []interface{}{content, label}
	} else {
		uid := normalizeUserKey(userKey)
		query = `INSERT INTO standing_orders (user_id, content, label) VALUES (?, ?, ?)`
		args = []interface{}{uid, content, label}
	}

	result, err := s.db.Exec(query, args...)
	if err != nil {
		return 0, fmt.Errorf("creating standing order: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("getting insert id: %w", err)
	}
	return id, nil
}

// ListStandingOrders lists standing orders with default user key (nil).
func (s *Storage) ListStandingOrders(activeOnly bool) ([]StandingOrder, error) {
	return s.ListStandingOrdersForUser(nil, activeOnly)
}

// ListStandingOrdersForUser lists standing orders for a specific user.
func (s *Storage) ListStandingOrdersForUser(userKey interface{}, activeOnly bool) ([]StandingOrder, error) {
	var query string
	var args []interface{}

	if s.isUserDB {
		if activeOnly {
			query = `SELECT id, content, label, active, created_at, updated_at FROM standing_orders WHERE active = 1 ORDER BY id ASC`
		} else {
			query = `SELECT id, content, label, active, created_at, updated_at FROM standing_orders ORDER BY id ASC`
		}
	} else {
		uid := normalizeUserKey(userKey)
		if activeOnly {
			query = `SELECT id, content, label, active, created_at, updated_at FROM standing_orders WHERE user_id = ? AND active = 1 ORDER BY id ASC`
		} else {
			query = `SELECT id, content, label, active, created_at, updated_at FROM standing_orders WHERE user_id = ? ORDER BY id ASC`
		}
		args = []interface{}{uid}
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing standing orders: %w", err)
	}
	defer rows.Close()

	var orders []StandingOrder
	for rows.Next() {
		var o StandingOrder
		if err := rows.Scan(&o.ID, &o.Content, &o.Label, &o.Active, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning standing order: %w", err)
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating standing orders: %w", err)
	}
	return orders, nil
}

// UpdateStandingOrder updates content and label for a standing order with default user key.
func (s *Storage) UpdateStandingOrder(id int64, content, label string) error {
	return s.UpdateStandingOrderForUser(nil, id, content, label)
}

// UpdateStandingOrderForUser updates content and label for a standing order.
func (s *Storage) UpdateStandingOrderForUser(userKey interface{}, id int64, content, label string) error {
	if len(content) > maxStandingOrderContent {
		return fmt.Errorf("content exceeds %d character limit (got %d)", maxStandingOrderContent, len(content))
	}

	var query string
	var args []interface{}

	if s.isUserDB {
		query = `UPDATE standing_orders SET content = ?, label = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
		args = []interface{}{content, label, id}
	} else {
		uid := normalizeUserKey(userKey)
		query = `UPDATE standing_orders SET content = ?, label = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ?`
		args = []interface{}{content, label, id, uid}
	}

	_, err := s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("updating standing order: %w", err)
	}
	return nil
}

// DeleteStandingOrder deletes a standing order with default user key.
func (s *Storage) DeleteStandingOrder(id int64) error {
	return s.DeleteStandingOrderForUser(nil, id)
}

// DeleteStandingOrderForUser deletes a standing order for a specific user.
func (s *Storage) DeleteStandingOrderForUser(userKey interface{}, id int64) error {
	var query string
	var args []interface{}

	if s.isUserDB {
		query = `DELETE FROM standing_orders WHERE id = ?`
		args = []interface{}{id}
	} else {
		uid := normalizeUserKey(userKey)
		query = `DELETE FROM standing_orders WHERE id = ? AND user_id = ?`
		args = []interface{}{id, uid}
	}

	_, err := s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("deleting standing order: %w", err)
	}
	return nil
}

// ToggleStandingOrder toggles the active state with default user key.
func (s *Storage) ToggleStandingOrder(id int64) error {
	return s.ToggleStandingOrderForUser(nil, id)
}

// ToggleStandingOrderForUser toggles the active state of a standing order.
func (s *Storage) ToggleStandingOrderForUser(userKey interface{}, id int64) error {
	var query string
	var args []interface{}

	if s.isUserDB {
		query = `UPDATE standing_orders SET active = NOT active, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
		args = []interface{}{id}
	} else {
		uid := normalizeUserKey(userKey)
		query = `UPDATE standing_orders SET active = NOT active, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ?`
		args = []interface{}{id, uid}
	}

	_, err := s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("toggling standing order: %w", err)
	}
	return nil
}
