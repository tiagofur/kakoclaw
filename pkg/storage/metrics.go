package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

// SaveMetricCounters upserts all counter key-value pairs into metrics_counters.
func (s *Storage) SaveMetricCounters(counters map[string]int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.Prepare(`INSERT INTO metrics_counters(key, value) VALUES(?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value`)
	if err != nil {
		return fmt.Errorf("prepare upsert: %w", err)
	}
	defer stmt.Close()

	for k, v := range counters {
		if _, err := stmt.Exec(k, v); err != nil {
			return fmt.Errorf("upsert counter %s: %w", k, err)
		}
	}

	return tx.Commit()
}

// LoadMetricCounters returns all previously stored counter values.
func (s *Storage) LoadMetricCounters() (map[string]int64, error) {
	rows, err := s.db.Query(`SELECT key, value FROM metrics_counters`)
	if err != nil {
		return nil, fmt.Errorf("query counters: %w", err)
	}
	defer rows.Close()

	result := make(map[string]int64)
	for rows.Next() {
		var k string
		var v int64
		if err := rows.Scan(&k, &v); err != nil {
			return nil, fmt.Errorf("scan counter: %w", err)
		}
		result[k] = v
	}
	return result, rows.Err()
}

// AppendMetricEvent persists a single event as JSON and prunes the table to maxRecentMetricEvents.
const maxRecentMetricEvents = 100

// AppendMetricEvent stores an event and keeps only the latest maxRecentMetricEvents rows.
func (s *Storage) AppendMetricEvent(event interface{}) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`INSERT INTO metrics_events(payload) VALUES(?)`, string(data)); err != nil {
		return fmt.Errorf("insert event: %w", err)
	}

	// Prune older events beyond the cap
	if _, err := tx.Exec(`DELETE FROM metrics_events WHERE id NOT IN (
		SELECT id FROM metrics_events ORDER BY id DESC LIMIT ?
	)`, maxRecentMetricEvents); err != nil {
		return fmt.Errorf("prune events: %w", err)
	}

	return tx.Commit()
}

// LoadRecentEvents returns up to maxRecentMetricEvents event payloads (oldest first).
func (s *Storage) LoadRecentEvents() ([]json.RawMessage, error) {
	rows, err := s.db.Query(`SELECT payload FROM metrics_events ORDER BY id ASC LIMIT ?`, maxRecentMetricEvents)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query events: %w", err)
	}
	defer rows.Close()

	var events []json.RawMessage
	for rows.Next() {
		var payload string
		if err := rows.Scan(&payload); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		events = append(events, json.RawMessage(payload))
	}
	return events, rows.Err()
}

// SaveMetricBreakdowns persists the per-model and per-tool stats as JSON blobs.
func (s *Storage) SaveMetricBreakdowns(modelMetrics json.RawMessage, toolMetrics json.RawMessage) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.Prepare(`INSERT INTO metrics_breakdowns(key, value) VALUES(?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec("llm_by_model", string(modelMetrics)); err != nil {
		return err
	}
	if _, err := stmt.Exec("tool_by_name", string(toolMetrics)); err != nil {
		return err
	}

	return tx.Commit()
}

// LoadMetricBreakdowns retrieves the per-model and per-tool stats as JSON blobs.
func (s *Storage) LoadMetricBreakdowns() (json.RawMessage, json.RawMessage, error) {
	var modelMetrics, toolMetrics json.RawMessage
	rows, err := s.db.Query(`SELECT key, value FROM metrics_breakdowns WHERE key IN ('llm_by_model', 'tool_by_name')`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, nil, err
		}
		if k == "llm_by_model" {
			modelMetrics = json.RawMessage(v)
		} else if k == "tool_by_name" {
			toolMetrics = json.RawMessage(v)
		}
	}
	return modelMetrics, toolMetrics, nil
}
