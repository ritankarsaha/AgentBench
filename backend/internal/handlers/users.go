package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ritankarsaha/agentbench/backend/internal/db/queries"
	"github.com/ritankarsaha/agentbench/backend/internal/middleware"
	"github.com/ritankarsaha/agentbench/backend/internal/response"
)

func SyncUser(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			response.WriteError(w, http.StatusUnauthorized, "missing_user", "user authentication required")
			return
		}

		id, err := uuid.Parse(userID)
		if err != nil {
			response.WriteError(w, http.StatusUnauthorized, "invalid_user", "invalid user id")
			return
		}

		if err := queries.EnsureContributorReputation(r.Context(), pool, id); err != nil {
			response.WriteError(w, http.StatusInternalServerError, "sync_failed", "failed to sync user")
			return
		}

		response.WriteOK(w, http.StatusOK, map[string]bool{"synced": true})
	}
}
