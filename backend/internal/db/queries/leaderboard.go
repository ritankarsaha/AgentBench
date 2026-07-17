package queries

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LeaderboardRow struct {
	AgentThreadsHandle string     `db:"agentthreads_handle" json:"agentthreads_handle"`
	DisplayName        string     `db:"display_name" json:"display_name"`
	Model              *string    `db:"model" json:"model"`
	Framework          *string    `db:"framework" json:"framework"`
	Suite              string     `db:"suite" json:"suite"`
	BestScore          float64    `db:"best_score" json:"best_score"`
	RunCount           int        `db:"run_count" json:"run_count"`
	LastRunAt          *time.Time `db:"last_run_at" json:"last_run_at"`
	HasVerifiedScore   bool       `db:"has_verified_score" json:"has_verified_score"`
	HasAnyTrace        bool       `db:"has_any_trace" json:"has_any_trace"`
}

const leaderboardColumns = `agentthreads_handle, display_name, model, framework, suite, best_score, run_count, last_run_at, has_verified_score, has_any_trace`

func GetLeaderboard(ctx context.Context, pool *pgxpool.Pool, suite, sort string, limit int) ([]LeaderboardRow, error) {
	orderBy := "best_score desc"
	switch sort {
	case "runs":
		orderBy = "run_count desc, best_score desc"
	case "verified":
		orderBy = "has_verified_score desc, best_score desc"
	}

	query := fmt.Sprintf(`select %s from leaderboard where suite = $1 order by %s limit $2`, leaderboardColumns, orderBy)

	rows, err := pool.Query(ctx, query, suite, limit)
	if err != nil {
		return nil, fmt.Errorf("queries: get leaderboard: %w", err)
	}
	entries, err := pgx.CollectRows(rows, pgx.RowToStructByName[LeaderboardRow])
	if err != nil {
		return nil, fmt.Errorf("queries: get leaderboard: %w", err)
	}
	return entries, nil
}
