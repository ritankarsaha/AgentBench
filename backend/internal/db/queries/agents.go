package queries

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("conflict")

const defaultListLimit = 20
const maxListLimit = 100

type Agent struct {
	ID                 uuid.UUID `db:"id" json:"id"`
	OwnerUserID        uuid.UUID `db:"owner_user_id" json:"-"`
	AgentThreadsHandle string    `db:"agentthreads_handle" json:"agentthreads_handle"`
	DisplayName        string    `db:"display_name" json:"display_name"`
	Description        *string   `db:"description" json:"description,omitempty"`
	Model              *string   `db:"model" json:"model,omitempty"`
	Framework          *string   `db:"framework" json:"framework,omitempty"`
	Tier               string    `db:"tier" json:"tier"`
	IsVerified         bool      `db:"is_verified" json:"is_verified"`
	TotalRuns          int       `db:"total_runs" json:"total_runs"`
	BestScore          float64   `db:"best_score" json:"best_score"`
	CreatedAt          time.Time `db:"created_at" json:"created_at"`
}

const agentColumns = `id, owner_user_id, agentthreads_handle, display_name, description, model, framework, tier, is_verified, total_runs, best_score, created_at`

type NewAgent struct {
	OwnerUserID            uuid.UUID
	AgentThreadsHandle     string
	AgentThreadsAPIKeyHash *string
	DisplayName            string
	Description            *string
	Model                  *string
	Framework              *string
	APIKeyHash             string
}

func CreateAgent(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID, in NewAgent) (*Agent, error) {
	rows, err := pool.Query(ctx, `
		insert into benchmark_agents (id, owner_user_id, agentthreads_handle, agentthreads_api_key_hash, display_name, description, model, framework, api_key_hash)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		returning `+agentColumns,
		id, in.OwnerUserID, in.AgentThreadsHandle, in.AgentThreadsAPIKeyHash, in.DisplayName, in.Description, in.Model, in.Framework, in.APIKeyHash,
	)
	if err != nil {
		return nil, fmt.Errorf("queries: create agent: %w", err)
	}
	agent, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Agent])
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrConflict
		}
		return nil, fmt.Errorf("queries: create agent: %w", err)
	}
	return &agent, nil
}

func GetAgentByID(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (*Agent, error) {
	rows, err := pool.Query(ctx, `select `+agentColumns+` from benchmark_agents where id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("queries: get agent by id: %w", err)
	}
	agent, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Agent])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("queries: get agent by id: %w", err)
	}
	return &agent, nil
}

func ListAgentsByOwner(ctx context.Context, pool *pgxpool.Pool, ownerUserID uuid.UUID) ([]Agent, error) {
	rows, err := pool.Query(ctx, `select `+agentColumns+` from benchmark_agents where owner_user_id = $1 order by created_at desc`, ownerUserID)
	if err != nil {
		return nil, fmt.Errorf("queries: list agents by owner: %w", err)
	}
	agents, err := pgx.CollectRows(rows, pgx.RowToStructByName[Agent])
	if err != nil {
		return nil, fmt.Errorf("queries: list agents by owner: %w", err)
	}
	return agents, nil
}

func GetAgentByHandle(ctx context.Context, pool *pgxpool.Pool, handle string) (*Agent, error) {
	rows, err := pool.Query(ctx, `select `+agentColumns+` from benchmark_agents where agentthreads_handle = $1`, handle)
	if err != nil {
		return nil, fmt.Errorf("queries: get agent by handle: %w", err)
	}
	agent, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Agent])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("queries: get agent by handle: %w", err)
	}
	return &agent, nil
}

// GetAgentAPIKeyHash is used only by the agent-auth middleware; the hash
// must never be attached to any handler-facing Agent value.
func GetAgentAPIKeyHash(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (string, error) {
	var hash string
	err := pool.QueryRow(ctx, `select api_key_hash from benchmark_agents where id = $1`, id).Scan(&hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("queries: get agent api key hash: %w", err)
	}
	return hash, nil
}

type AgentPatch struct {
	DisplayName *string
	Description *string
	Model       *string
	Framework   *string
}

func UpdateAgent(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID, patch AgentPatch) (*Agent, error) {
	var sets []string
	var args []any

	if patch.DisplayName != nil {
		args = append(args, *patch.DisplayName)
		sets = append(sets, fmt.Sprintf("display_name = $%d", len(args)))
	}
	if patch.Description != nil {
		args = append(args, *patch.Description)
		sets = append(sets, fmt.Sprintf("description = $%d", len(args)))
	}
	if patch.Model != nil {
		args = append(args, *patch.Model)
		sets = append(sets, fmt.Sprintf("model = $%d", len(args)))
	}
	if patch.Framework != nil {
		args = append(args, *patch.Framework)
		sets = append(sets, fmt.Sprintf("framework = $%d", len(args)))
	}
	if len(sets) == 0 {
		return GetAgentByID(ctx, pool, id)
	}

	args = append(args, id)
	query := fmt.Sprintf(`update benchmark_agents set %s where id = $%d returning `+agentColumns, strings.Join(sets, ", "), len(args))

	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("queries: update agent: %w", err)
	}
	agent, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Agent])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("queries: update agent: %w", err)
	}
	return &agent, nil
}

func DeleteAgent(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) error {
	tag, err := pool.Exec(ctx, `delete from benchmark_agents where id = $1`, id)
	if err != nil {
		return fmt.Errorf("queries: delete agent: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

type AgentFilter struct {
	Handle    *string
	Model     *string
	Framework *string
	Verified  *bool
	Cursor    *uuid.UUID
	Limit     int
}

func ClampListLimit(limit int) int {
	if limit <= 0 || limit > maxListLimit {
		return defaultListLimit
	}
	return limit
}

func ListAgents(ctx context.Context, pool *pgxpool.Pool, f AgentFilter) ([]Agent, error) {
	limit := ClampListLimit(f.Limit)

	where := []string{"true"}
	var args []any

	if f.Handle != nil {
		args = append(args, *f.Handle)
		where = append(where, fmt.Sprintf("agentthreads_handle = $%d", len(args)))
	}
	if f.Model != nil {
		args = append(args, *f.Model)
		where = append(where, fmt.Sprintf("model = $%d", len(args)))
	}
	if f.Framework != nil {
		args = append(args, *f.Framework)
		where = append(where, fmt.Sprintf("framework = $%d", len(args)))
	}
	if f.Verified != nil {
		args = append(args, *f.Verified)
		where = append(where, fmt.Sprintf("is_verified = $%d", len(args)))
	}
	if f.Cursor != nil {
		args = append(args, *f.Cursor)
		where = append(where, fmt.Sprintf(`(created_at, id) < (select created_at, id from benchmark_agents where id = $%d)`, len(args)))
	}

	args = append(args, limit)
	query := fmt.Sprintf(
		`select %s from benchmark_agents where %s order by created_at desc, id desc limit $%d`,
		agentColumns, strings.Join(where, " and "), len(args),
	)

	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("queries: list agents: %w", err)
	}
	agents, err := pgx.CollectRows(rows, pgx.RowToStructByName[Agent])
	if err != nil {
		return nil, fmt.Errorf("queries: list agents: %w", err)
	}
	return agents, nil
}

type RunSummary struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	Suite           string     `db:"suite" json:"suite"`
	EffectiveScore  *float64   `db:"effective_score" json:"effective_score"`
	IsTraceVerified bool       `db:"is_trace_verified" json:"is_trace_verified"`
	CompletedAt     *time.Time `db:"completed_at" json:"completed_at"`
}

func GetAgentRecentRuns(ctx context.Context, pool *pgxpool.Pool, agentID uuid.UUID, limit int) ([]RunSummary, error) {
	rows, err := pool.Query(ctx, `
		select r.id, s.slug as suite, r.effective_score, r.is_trace_verified, r.completed_at
		from benchmark_runs r
		join benchmark_suites s on s.id = r.suite_id
		where r.agent_id = $1 and r.status = 'complete'
		order by r.completed_at desc
		limit $2`,
		agentID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("queries: get agent recent runs: %w", err)
	}
	runs, err := pgx.CollectRows(rows, pgx.RowToStructByName[RunSummary])
	if err != nil {
		return nil, fmt.Errorf("queries: get agent recent runs: %w", err)
	}
	return runs, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
