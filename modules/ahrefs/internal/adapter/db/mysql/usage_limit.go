package dbmysql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/wahyudibo/golang-reverse-proxy/modules/ahrefs/internal/repository"
)

type UsageLimitRepository struct {
	db *sql.DB
}

func (r *UsageLimitRepository) Create(ctx context.Context, userID int64) error {
	query := `INSERT INTO ahref_usage_limit VALUES (?, 0, 0, NOW())`
	if _, err := r.db.ExecContext(ctx, query, userID); err != nil {
		return err
	}

	return nil
}

func (r *UsageLimitRepository) Retrieve(ctx context.Context, userID int64) (*repository.UsageLimit, error) {
	var m repository.UsageLimit

	query := `
	SELECT
		user_id,
		report_usage,
		export_usage,
		TIMESTAMPDIFF(MINUTE,last_accessed_at, NOW()) AS limit_reset_at
	FROM ahref_usage_limit
	WHERE user_id = ?`

	if err := r.db.QueryRowContext(ctx, query, userID).Scan(&m.UserID, &m.ReportUsage, &m.ExportUsage, &m.LimitResetAt); err != nil {
		return nil, err
	}

	return &m, nil
}

func (r *UsageLimitRepository) Update(ctx context.Context, userID int64, reportUsage, exportUsage int, updateLastAccessedAt bool) error {
	query := `UPDATE ahref_usage_limit SET report_usage = ?, export_usage = ? `

	if updateLastAccessedAt {
		query += ", last_accessed_at = NOW() "
	}

	query += "WHERE user_id = ?"

	res, err := r.db.ExecContext(ctx, query, reportUsage, exportUsage, userID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("row is not found")
	}

	return nil
}
