package queries

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// EnsureContributorReputation idempotently bootstraps a reputation row for
// a human user on first login
func EnsureContributorReputation(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) error {
	_, err := pool.Exec(ctx, `
		insert into contributor_reputation (user_id)
		values ($1)
		on conflict (user_id) do nothing`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("queries: ensure contributor reputation: %w", err)
	}
	return nil
}
