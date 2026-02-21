package storage

import "database/sql"

// GetUserIDForChannelSender returns the user ID mapped to a channel sender ID.
// Returns 0 if no mapping is found.
func (s *Storage) GetUserIDForChannelSender(channel, senderID string) (int64, error) {
	var userID int64
	err := s.db.QueryRow(
		`SELECT user_id FROM channel_users WHERE channel = ? AND sender_id = ?`,
		channel,
		senderID,
	).Scan(&userID)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return userID, nil
}

// SetUserIDForChannelSender upserts a mapping between a channel sender ID and user ID.
func (s *Storage) SetUserIDForChannelSender(channel, senderID string, userID int64) error {
	_, err := s.db.Exec(
		`INSERT INTO channel_users (channel, sender_id, user_id)
		 VALUES (?, ?, ?)
		 ON CONFLICT(channel, sender_id) DO UPDATE SET user_id = excluded.user_id, updated_at = CURRENT_TIMESTAMP`,
		channel,
		senderID,
		userID,
	)
	return err
}
