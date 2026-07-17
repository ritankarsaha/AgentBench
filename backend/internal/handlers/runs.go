package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ritankarsaha/agentbench/backend/internal/db/queries"
	"github.com/ritankarsaha/agentbench/backend/internal/middleware"
	"github.com/ritankarsaha/agentbench/backend/internal/response"
	"github.com/ritankarsaha/agentbench/backend/internal/scoring"
)

type createRunRequest struct {
	Suite string `json:"suite"`
}

func CreateRun(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		agent, ok := middleware.AgentFromContext(r.Context())
		if !ok {
			response.WriteError(w, http.StatusUnauthorized, "missing_agent", "agent authentication required")
			return
		}

		var req createRunRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Suite == "" {
			response.WriteError(w, http.StatusBadRequest, "invalid_body", "suite is required")
			return
		}

		suite, err := queries.GetSuiteBySlug(r.Context(), pool, req.Suite)
		if err != nil {
			if errors.Is(err, queries.ErrNotFound) {
				response.WriteError(w, http.StatusNotFound, "not_found", "suite not found")
				return
			}
			response.WriteError(w, http.StatusInternalServerError, "lookup_failed", "failed to look up suite")
			return
		}

		tasks, err := queries.GetTasksForSuite(r.Context(), pool, suite.TaskIDs)
		if err != nil {
			response.WriteError(w, http.StatusInternalServerError, "lookup_failed", "failed to load suite tasks")
			return
		}
		if len(tasks) == 0 {
			response.WriteError(w, http.StatusConflict, "empty_suite", "suite has no active tasks")
			return
		}

		run, err := queries.CreateRun(r.Context(), pool, uuid.New(), agent.ID, suite.ID, len(tasks))
		if err != nil {
			response.WriteError(w, http.StatusInternalServerError, "create_failed", "failed to create run")
			return
		}

		response.WriteOK(w, http.StatusCreated, map[string]any{"run_id": run.ID, "tasks": tasks})
	}
}

func GetRun(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		runID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid run id")
			return
		}

		run, err := queries.GetRun(r.Context(), pool, runID)
		if err != nil {
			if errors.Is(err, queries.ErrNotFound) {
				response.WriteError(w, http.StatusNotFound, "not_found", "run not found")
				return
			}
			response.WriteError(w, http.StatusInternalServerError, "lookup_failed", "failed to look up run")
			return
		}

		if run.Status != "complete" {
			token := middleware.BearerToken(r)
			if token == "" {
				response.WriteError(w, http.StatusUnauthorized, "missing_token", "agent authentication required for an in-progress run")
				return
			}
			agent, err := middleware.VerifyAgentToken(r.Context(), pool, token)
			if err != nil {
				response.WriteError(w, http.StatusUnauthorized, "invalid_token", "invalid api key")
				return
			}
			if agent.ID != run.AgentID {
				response.WriteError(w, http.StatusForbidden, "forbidden", "not your run")
				return
			}
		}

		results, err := queries.GetRunTaskResults(r.Context(), pool, runID)
		if err != nil {
			response.WriteError(w, http.StatusInternalServerError, "lookup_failed", "failed to load task results")
			return
		}

		response.WriteOK(w, http.StatusOK, runDetail{Run: *run, Results: results})
	}
}

// runDetail flattens the run's fields alongside its task results, so
// GET /runs/:id's response shape matches POST /runs/:id/complete's
// (both have run fields at the top level of `data`) — the SDK polls both
// through the same code path.
type runDetail struct {
	queries.Run
	Results []queries.TaskResult `json:"results"`
}

type submitResultRequest struct {
	TaskID  uuid.UUID       `json:"task_id"`
	Output  json.RawMessage `json:"output"`
	TraceID *string         `json:"trace_id"`
}

func SubmitResult(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		agent, ok := middleware.AgentFromContext(r.Context())
		if !ok {
			response.WriteError(w, http.StatusUnauthorized, "missing_agent", "agent authentication required")
			return
		}

		runID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid run id")
			return
		}

		var req submitResultRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.Output) == 0 || req.TaskID == uuid.Nil {
			response.WriteError(w, http.StatusBadRequest, "invalid_body", "task_id and output are required")
			return
		}

		run, err := queries.GetRun(r.Context(), pool, runID)
		if err != nil {
			if errors.Is(err, queries.ErrNotFound) {
				response.WriteError(w, http.StatusNotFound, "not_found", "run not found")
				return
			}
			response.WriteError(w, http.StatusInternalServerError, "lookup_failed", "failed to look up run")
			return
		}
		if run.AgentID != agent.ID {
			response.WriteError(w, http.StatusForbidden, "forbidden", "not your run")
			return
		}
		if run.Status != "running" {
			response.WriteError(w, http.StatusBadRequest, "invalid_state", "run is not accepting results")
			return
		}

		suite, err := queries.GetSuiteByID(r.Context(), pool, run.SuiteID)
		if err != nil {
			response.WriteError(w, http.StatusInternalServerError, "lookup_failed", "failed to look up suite")
			return
		}
		if !containsID(suite.TaskIDs, req.TaskID) {
			response.WriteError(w, http.StatusBadRequest, "invalid_task", "task is not part of this run")
			return
		}

		task, err := queries.GetTaskScoringInfo(r.Context(), pool, req.TaskID)
		if err != nil {
			response.WriteError(w, http.StatusInternalServerError, "lookup_failed", "failed to look up task")
			return
		}

		score, err := scoring.Score(task.Type, task.ExpectedOutput, req.Output)
		if err != nil {
			response.WriteError(w, http.StatusBadRequest, "scoring_failed", err.Error())
			return
		}

		result, err := queries.SubmitTaskResult(r.Context(), pool, runID, req.TaskID, req.Output, score, req.TraceID)
		if err != nil {
			response.WriteError(w, http.StatusInternalServerError, "submit_failed", "failed to submit result")
			return
		}

		response.WriteOK(w, http.StatusOK, result)
	}
}

func CompleteRun(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		agent, ok := middleware.AgentFromContext(r.Context())
		if !ok {
			response.WriteError(w, http.StatusUnauthorized, "missing_agent", "agent authentication required")
			return
		}

		runID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid run id")
			return
		}

		run, err := queries.GetRun(r.Context(), pool, runID)
		if err != nil {
			if errors.Is(err, queries.ErrNotFound) {
				response.WriteError(w, http.StatusNotFound, "not_found", "run not found")
				return
			}
			response.WriteError(w, http.StatusInternalServerError, "lookup_failed", "failed to look up run")
			return
		}
		if run.AgentID != agent.ID {
			response.WriteError(w, http.StatusForbidden, "forbidden", "not your run")
			return
		}
		if run.Status != "running" {
			response.WriteError(w, http.StatusBadRequest, "invalid_state", "run is already complete or not running")
			return
		}

		agg, err := queries.GetRunAggregate(r.Context(), pool, runID)
		if err != nil {
			response.WriteError(w, http.StatusInternalServerError, "aggregate_failed", "failed to compute run score")
			return
		}

		completedAt := time.Now()
		var effectiveScore *float64
		if agg.RawScore != nil {
			ef := *agg.RawScore * scoring.DecayFactor(completedAt)
			effectiveScore = &ef
		}

		updated, err := queries.CompleteRun(r.Context(), pool, runID, agg.RawScore, effectiveScore, agg.TasksVerified, agg.IsTraceVerified, completedAt)
		if err != nil {
			response.WriteError(w, http.StatusInternalServerError, "complete_failed", "failed to complete run")
			return
		}

		response.WriteOK(w, http.StatusOK, updated)
	}
}

func containsID(ids []uuid.UUID, target uuid.UUID) bool {
	for _, id := range ids {
		if id == target {
			return true
		}
	}
	return false
}
