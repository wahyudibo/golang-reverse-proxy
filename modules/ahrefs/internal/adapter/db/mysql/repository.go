package dbmysql

import (
	"database/sql"

	"github.com/wahyudibo/golang-reverse-proxy/modules/ahrefs/internal/repository"
)

type Repository struct {
	usageLimit *UsageLimitRepository
	session    *SessionRepository
}

func New(db *sql.DB) *Repository {
	usageLimitRepository := &UsageLimitRepository{
		db: db,
	}
	sessionRepository := &SessionRepository{
		db: db,
	}

	return &Repository{
		usageLimit: usageLimitRepository,
		session:    sessionRepository,
	}
}

func (r *Repository) UsageLimit() repository.UsageLimitRepository {
	return r.usageLimit
}

func (r *Repository) Session() repository.SessionRepository {
	return r.session
}
