package storage

import "fmt"

// BackfillUserID assigns a user ID to records missing user_id values.
func (s *Storage) BackfillUserID(userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("invalid user ID")
	}

	queries := []string{
		`UPDATE chats SET user_id = ? WHERE user_id IS NULL OR user_id = 0`,
		`UPDATE tasks SET user_id = ? WHERE user_id IS NULL OR user_id = 0`,
		`UPDATE sessions SET user_id = ? WHERE user_id IS NULL OR user_id = 0`,
	}

	for _, q := range queries {
		if _, err := s.db.Exec(q, userID); err != nil {
			return fmt.Errorf("backfilling user_id: %w", err)
		}
	}

	return nil
}
