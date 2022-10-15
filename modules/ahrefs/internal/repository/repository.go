package repository

type Repository interface {
	UsageLimit() UsageLimitRepository
	Session() SessionRepository
}
