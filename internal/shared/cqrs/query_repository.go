package cqrs

import (
	"context"
)

// UserQueryRepository defines read operations for users
type UserQueryRepository interface {
	// FindByID retrieves a user by ID
	FindByID(ctx context.Context, id string) (*UserReadModel, error)

	// FindByEmail retrieves a user by email
	FindByEmail(ctx context.Context, email string) (*UserReadModel, error)

	// FindAll retrieves users matching the query
	FindAll(ctx context.Context, query *UserQuery) ([]*UserReadModel, error)

	// GetStatistics retrieves aggregated statistics for a user
	GetStatistics(ctx context.Context, userID string) (*UserStatistics, error)

	// Count returns the total count of users matching the query
	Count(ctx context.Context, query *UserQuery) (int, error)
}

// TransactionQueryRepository defines read operations for transactions
type TransactionQueryRepository interface {
	// FindByID retrieves a transaction by ID
	FindByID(ctx context.Context, id string) (*TransactionReadModel, error)

	// FindByUser retrieves transactions for a specific user
	FindByUser(ctx context.Context, userID string, limit, offset int) ([]*TransactionReadModel, error)

	// FindAll retrieves transactions matching the query
	FindAll(ctx context.Context, query *TransactionQuery) ([]*TransactionReadModel, error)

	// Count returns the total count of transactions matching the query
	Count(ctx context.Context, query *TransactionQuery) (int, error)

	// FindRecent retrieves the most recent transactions
	FindRecent(ctx context.Context, limit int) ([]*TransactionReadModel, error)
}

// ProjectionUpdater handles updating read models from events
type ProjectionUpdater interface {
	// UpdateUserProjection updates the user read model
	UpdateUserProjection(ctx context.Context, event interface{}) error

	// UpdateTransactionProjection updates the transaction read model
	UpdateTransactionProjection(ctx context.Context, event interface{}) error

	// RebuildProjections rebuilds all projections from event store
	RebuildProjections(ctx context.Context) error
}
