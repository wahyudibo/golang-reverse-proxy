package repository

import "context"

type UsageLimitRepository interface {
	Create(ctx context.Context, userID int64) error
	Retrieve(ctx context.Context, userID int64) (*UsageLimit, error)
	Update(ctx context.Context, userID int64, reportUsage, exportUsage int, lastAccessedAt bool) error
}

type UsageLimit struct {
	UserID       int64
	ReportUsage  int
	ExportUsage  int
	LimitResetAt int
}
