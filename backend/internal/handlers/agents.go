package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ritankarsaha/agentbench/backend/internal/apikey"
	"github.com/ritankarsaha/agentbench/backend/internal/db/queries"
	"github.com/ritankarsaha/agentbench/backend/internal/middleware"
	"github.com/ritankarsaha/agentbench/backend/internal/response"
)

type registerAgentRequest struct {
	AgentThreadsHandle string  `json:"agentthreads_handle"`
	AgentThreadsAPIKey string  `json:"agentthreads_api_key"`
	DisplayName        string  `json:"display_name"`
	Description        *string `json:"description"`
	Model              *string `json:"model"`
	Framework          *string `json:"framework"`
}

func RegisterAgent(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			response.WriteError(w, http.StatusUnauthorized, "missing_user", "user authentication required")
			return
		}
		ownerID, err := uuid.Parse(userID)
		if err != nil {
			response.WriteError(w, http.StatusUnauthorized, "invalid_user", "invalid user id")
			return
		}

		var req registerAgentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteError(w, http.StatusBadRequest, "invalid_body", "malformed JSON body")
			return
		}
		if req.AgentThreadsHandle == "" || req.AgentThreadsAPIKey == "" || req.DisplayName == "" {
			response.WriteError(w, http.StatusBadRequest, "missing_fields", "agentthreads_handle, agentthreads_api_key, and display_name are required")
			return
		}

		agentID := uuid.New()
		plaintextKey, apiKeyHash, err := apikey.Generate(agentID)
		if err != nil {
			response.WriteError(w, http.StatusInternalServerError, "key_generation_failed", "failed to generate api key")
			return
		}
		agentThreadsHash, err := apikey.HashArbitrary(req.AgentThreadsAPIKey)
		if err != nil {
			response.WriteError(w, http.StatusInternalServerError, "key_generation_failed", "failed to process agentthreads api key")
			return
		}

		agent, err := queries.CreateAgent(r.Context(), pool, agentID, queries.NewAgent{
			OwnerUserID:            ownerID,
			AgentThreadsHandle:     req.AgentThreadsHandle,
			AgentThreadsAPIKeyHash: agentThreadsHash,
			DisplayName:            req.DisplayName,
			Description:            req.Description,
			Model:                  req.Model,
			Framework:              req.Framework,
			APIKeyHash:             apiKeyHash,
		})
		if err != nil {
			if errors.Is(err, queries.ErrConflict) {
				response.WriteError(w, http.StatusConflict, "handle_taken", "agentthreads_handle is already registered")
				return
			}
			response.WriteError(w, http.StatusInternalServerError, "create_failed", "failed to create agent")
			return
		}

		response.WriteOK(w, http.StatusCreated, map[string]any{
			"agent":   agent,
			"api_key": plaintextKey,
		})
	}
}

func GetMe(w http.ResponseWriter, r *http.Request) {
	agent, ok := middleware.AgentFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "missing_agent", "agent authentication required")
		return
	}
	response.WriteOK(w, http.StatusOK, agent)
}

type updateAgentRequest struct {
	DisplayName *string `json:"display_name"`
	Description *string `json:"description"`
	Model       *string `json:"model"`
	Framework   *string `json:"framework"`
}

func UpdateMe(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		agent, ok := middleware.AgentFromContext(r.Context())
		if !ok {
			response.WriteError(w, http.StatusUnauthorized, "missing_agent", "agent authentication required")
			return
		}

		var req updateAgentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteError(w, http.StatusBadRequest, "invalid_body", "malformed JSON body")
			return
		}

		updated, err := queries.UpdateAgent(r.Context(), pool, agent.ID, queries.AgentPatch{
			DisplayName: req.DisplayName,
			Description: req.Description,
			Model:       req.Model,
			Framework:   req.Framework,
		})
		if err != nil {
			response.WriteError(w, http.StatusInternalServerError, "update_failed", "failed to update agent")
			return
		}

		response.WriteOK(w, http.StatusOK, updated)
	}
}

func DeleteMe(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		agent, ok := middleware.AgentFromContext(r.Context())
		if !ok {
			response.WriteError(w, http.StatusUnauthorized, "missing_agent", "agent authentication required")
			return
		}

		if err := queries.DeleteAgent(r.Context(), pool, agent.ID); err != nil {
			response.WriteError(w, http.StatusInternalServerError, "delete_failed", "failed to delete agent")
			return
		}

		response.WriteOK(w, http.StatusOK, map[string]bool{"deleted": true})
	}
}

func ListAgents(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		var filter queries.AgentFilter

		if v := q.Get("handle"); v != "" {
			filter.Handle = &v
		}
		if v := q.Get("model"); v != "" {
			filter.Model = &v
		}
		if v := q.Get("framework"); v != "" {
			filter.Framework = &v
		}
		if v := q.Get("verified"); v != "" {
			b, err := strconv.ParseBool(v)
			if err != nil {
				response.WriteError(w, http.StatusBadRequest, "invalid_verified", "verified must be true or false")
				return
			}
			filter.Verified = &b
		}
		if v := q.Get("cursor"); v != "" {
			id, err := uuid.Parse(v)
			if err != nil {
				response.WriteError(w, http.StatusBadRequest, "invalid_cursor", "cursor must be a valid agent id")
				return
			}
			filter.Cursor = &id
		}
		if v := q.Get("limit"); v != "" {
			limit, err := strconv.Atoi(v)
			if err != nil {
				response.WriteError(w, http.StatusBadRequest, "invalid_limit", "limit must be an integer")
				return
			}
			filter.Limit = limit
		}

		agents, err := queries.ListAgents(r.Context(), pool, filter)
		if err != nil {
			response.WriteError(w, http.StatusInternalServerError, "list_failed", "failed to list agents")
			return
		}

		var cursor string
		if len(agents) > 0 && len(agents) == queries.ClampListLimit(filter.Limit) {
			cursor = agents[len(agents)-1].ID.String()
		}

		response.WriteOKWithCursor(w, http.StatusOK, agents, cursor)
	}
}

func GetAgentByHandle(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handle := chi.URLParam(r, "handle")

		agent, err := queries.GetAgentByHandle(r.Context(), pool, handle)
		if err != nil {
			if errors.Is(err, queries.ErrNotFound) {
				response.WriteError(w, http.StatusNotFound, "not_found", "agent not found")
				return
			}
			response.WriteError(w, http.StatusInternalServerError, "lookup_failed", "failed to look up agent")
			return
		}

		runs, err := queries.GetAgentRecentRuns(r.Context(), pool, agent.ID, 10)
		if err != nil {
			response.WriteError(w, http.StatusInternalServerError, "lookup_failed", "failed to load recent runs")
			return
		}

		response.WriteOK(w, http.StatusOK, map[string]any{
			"agent":       agent,
			"recent_runs": runs,
		})
	}
}
