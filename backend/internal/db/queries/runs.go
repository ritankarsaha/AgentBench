package queries

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Suite struct {
	ID      uuid.UUID   `db:"id"`
	Slug    string      `db:"slug"`
	TaskIDs []uuid.UUID `db:"task_ids"`
}

const suiteColumns = `id, slug, task_ids`

func GetSuiteBySlug(ctx context.Context, pool *pgxpool.Pool, slug string) (*Suite, error) {
	rows, err := pool.Query(ctx, `select `+suiteColumns+` from benchmark_suites where slug = $1`, slug)
	if err != nil {
		return nil, fmt.Errorf("queries: get suite by slug: %w", err)
	}
	suite, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Suite])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("queries: get suite by slug: %w", err)
	}
	return &suite, nil
}

func GetSuiteByID(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (*Suite, error) {
	rows, err := pool.Query(ctx, `select `+suiteColumns+` from benchmark_suites where id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("queries: get suite by id: %w", err)
	}
	suite, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Suite])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("queries: get suite by id: %w", err)
	}
	return &suite, nil
}

type TaskForRun struct {
	ID    uuid.UUID       `db:"id" json:"id"`
	Type  string          `db:"type" json:"type"`
	Title string          `db:"title" json:"title"`
	Input json.RawMessage `db:"input" json:"input"`
}

func GetTasksForSuite(ctx context.Context, pool *pgxpool.Pool, taskIDs []uuid.UUID) ([]TaskForRun, error) {
	rows, err := pool.Query(ctx, `
		select id, type, title, input
		from benchmark_tasks
		where id = any($1) and status = 'active'
		order by array_position($1::uuid[], id)`,
		taskIDs,
	)
	if err != nil {
		return nil, fmt.Errorf("queries: get tasks for suite: %w", err)
	}
	tasks, err := pgx.CollectRows(rows, pgx.RowToStructByName[TaskForRun])
	if err != nil {
		return nil, fmt.Errorf("queries: get tasks for suite: %w", err)
	}
	return tasks, nil
}

type TaskScoringInfo struct {
	ID             uuid.UUID       `db:"id"`
	Type           string          `db:"type"`
	ExpectedOutput json.RawMessage `db:"expected_output"`
	Weight         float64         `db:"weight"`
}

func GetTaskScoringInfo(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (*TaskScoringInfo, error) {
	rows, err := pool.Query(ctx, `select id, type, expected_output, weight from benchmark_tasks where id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("queries: get task scoring info: %w", err)
	}
	info, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TaskScoringInfo])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("queries: get task scoring info: %w", err)
	}
	return &info, nil
}

type Run struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	AgentID         uuid.UUID  `db:"agent_id" json:"-"`
	SuiteID         uuid.UUID  `db:"suite_id" json:"suite_id"`
	Status          string     `db:"status" json:"status"`
	RawScore        *float64   `db:"raw_score" json:"raw_score"`
	EffectiveScore  *float64   `db:"effective_score" json:"effective_score"`
	TasksTotal      int        `db:"tasks_total" json:"tasks_total"`
	TasksComplete   int        `db:"tasks_complete" json:"tasks_complete"`
	TasksVerified   int        `db:"tasks_verified" json:"tasks_verified"`
	IsTraceVerified bool       `db:"is_trace_verified" json:"is_trace_verified"`
	StartedAt       time.Time  `db:"started_at" json:"started_at"`
	CompletedAt     *time.Time `db:"completed_at" json:"completed_at"`
}

const runColumns = `id, agent_id, suite_id, status, raw_score, effective_score, tasks_total, tasks_complete, tasks_verified, is_trace_verified, started_at, completed_at`

func CreateRun(ctx context.Context, pool *pgxpool.Pool, id, agentID, suiteID uuid.UUID, tasksTotal int) (*Run, error) {
	rows, err := pool.Query(ctx, `
		insert into benchmark_runs (id, agent_id, suite_id, status, tasks_total)
		values ($1, $2, $3, 'running', $4)
		returning `+runColumns,
		id, agentID, suiteID, tasksTotal,
	)
	if err != nil {
		return nil, fmt.Errorf("queries: create run: %w", err)
	}
	run, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Run])
	if err != nil {
		return nil, fmt.Errorf("queries: create run: %w", err)
	}
	return &run, nil
}

func GetRun(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (*Run, error) {
	rows, err := pool.Query(ctx, `select `+runColumns+` from benchmark_runs where id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("queries: get run: %w", err)
	}
	run, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Run])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("queries: get run: %w", err)
	}
	return &run, nil
}

type TaskResult struct {
	ID            uuid.UUID       `db:"id" json:"id"`
	RunID         uuid.UUID       `db:"run_id" json:"run_id"`
	TaskID        uuid.UUID       `db:"task_id" json:"task_id"`
	Status        string          `db:"status" json:"status"`
	AgentOutput   json.RawMessage `db:"agent_output" json:"agent_output"`
	Score         *float64        `db:"score" json:"score"`
	TraceID       *string         `db:"trace_id" json:"trace_id"`
	TraceVerified bool            `db:"trace_verified" json:"trace_verified"`
	SubmittedAt   *time.Time      `db:"submitted_at" json:"submitted_at"`
	ScoredAt      *time.Time      `db:"scored_at" json:"scored_at"`
}

const taskResultColumns = `id, run_id, task_id, status, agent_output, score, trace_id, trace_verified, submitted_at, scored_at`

func GetRunTaskResults(ctx context.Context, pool *pgxpool.Pool, runID uuid.UUID) ([]TaskResult, error) {
	rows, err := pool.Query(ctx, `select `+taskResultColumns+` from task_results where run_id = $1 order by submitted_at`, runID)
	if err != nil {
		return nil, fmt.Errorf("queries: get run task results: %w", err)
	}
	results, err := pgx.CollectRows(rows, pgx.RowToStructByName[TaskResult])
	if err != nil {
		return nil, fmt.Errorf("queries: get run task results: %w", err)
	}
	return results, nil
}

func SubmitTaskResult(ctx context.Context, pool *pgxpool.Pool, runID, taskID uuid.UUID, output json.RawMessage, score float64, traceID *string) (*TaskResult, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("queries: submit task result: begin: %w", err)
	}
	defer tx.Rollback(ctx)

	var inserted bool
	err = tx.QueryRow(ctx, `
		insert into task_results (run_id, task_id, status, agent_output, score, trace_id, submitted_at, scored_at)
		values ($1, $2, 'scored', $3, $4, $5, now(), now())
		on conflict (run_id, task_id) do update
			set agent_output = excluded.agent_output,
			    score = excluded.score,
			    trace_id = excluded.trace_id,
			    status = 'scored',
			    submitted_at = now(),
			    scored_at = now()
		returning (xmax = 0)`,
		runID, taskID, output, score, traceID,
	).Scan(&inserted)
	if err != nil {
		return nil, fmt.Errorf("queries: submit task result: upsert: %w", err)
	}

	if inserted {
		if _, err := tx.Exec(ctx, `update benchmark_runs set tasks_complete = tasks_complete + 1 where id = $1`, runID); err != nil {
			return nil, fmt.Errorf("queries: submit task result: increment: %w", err)
		}
	}

	rows, err := tx.Query(ctx, `select `+taskResultColumns+` from task_results where run_id = $1 and task_id = $2`, runID, taskID)
	if err != nil {
		return nil, fmt.Errorf("queries: submit task result: reload: %w", err)
	}
	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TaskResult])
	if err != nil {
		return nil, fmt.Errorf("queries: submit task result: reload: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("queries: submit task result: commit: %w", err)
	}
	return &result, nil
}

type RunAggregate struct {
	RawScore        *float64
	TasksVerified   int
	IsTraceVerified bool
}

func GetRunAggregate(ctx context.Context, pool *pgxpool.Pool, runID uuid.UUID) (*RunAggregate, error) {
	var agg RunAggregate
	err := pool.QueryRow(ctx, `
		select
			sum(tr.score * bt.weight) / nullif(sum(bt.weight), 0),
			count(*) filter (where tr.trace_verified),
			coalesce(bool_and(tr.trace_verified), false)
		from task_results tr
		join benchmark_tasks bt on bt.id = tr.task_id
		where tr.run_id = $1`,
		runID,
	).Scan(&agg.RawScore, &agg.TasksVerified, &agg.IsTraceVerified)
	if err != nil {
		return nil, fmt.Errorf("queries: get run aggregate: %w", err)
	}
	return &agg, nil
}

// CompleteRun persists precomputed aggregate values
func CompleteRun(ctx context.Context, pool *pgxpool.Pool, runID uuid.UUID, rawScore, effectiveScore *float64, tasksVerified int, isTraceVerified bool, completedAt time.Time) (*Run, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("queries: complete run: begin: %w", err)
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
		update benchmark_runs
		set status = 'complete', raw_score = $2, effective_score = $3, tasks_verified = $4, is_trace_verified = $5, completed_at = $6
		where id = $1
		returning `+runColumns,
		runID, rawScore, effectiveScore, tasksVerified, isTraceVerified, completedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("queries: complete run: update: %w", err)
	}
	run, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Run])
	if err != nil {
		return nil, fmt.Errorf("queries: complete run: update: %w", err)
	}

	if effectiveScore != nil {
		if _, err := tx.Exec(ctx, `
			update benchmark_agents
			set total_runs = total_runs + 1, best_score = greatest(best_score, $1)
			where id = $2`,
			*effectiveScore, run.AgentID,
		); err != nil {
			return nil, fmt.Errorf("queries: complete run: update agent: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("queries: complete run: commit: %w", err)
	}
	return &run, nil
}
