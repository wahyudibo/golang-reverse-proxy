package dbmysql

import (
	"context"
	"database/sql"
)

type SessionRepository struct {
	db *sql.DB
}

func (r *SessionRepository) FindUserIDBySession(ctx context.Context, sessionID string) (int64, error) {
	var userID int64

	query := `SELECT user_id FROM am_session WHERE id = ?`

	if err := r.db.QueryRowContext(ctx, query, userID).Scan(userID); err != nil {
		return userID, err
	}

	return userID, nil
}
