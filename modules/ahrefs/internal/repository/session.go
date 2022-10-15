package repository

import "context"

type SessionRepository interface {
	FindUserIDBySession(ctx context.Context, sessionID string) (int64, error)
}
