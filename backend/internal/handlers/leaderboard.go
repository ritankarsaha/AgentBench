package handlers

import (
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ritankarsaha/agentbench/backend/internal/db/queries"
	"github.com/ritankarsaha/agentbench/backend/internal/response"
)

const defaultLeaderboardLimit = 50
const maxLeaderboardLimit = 200

func Leaderboard(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		suite := q.Get("suite")
		if suite == "" {
			suite = "standard"
		}

		limit := defaultLeaderboardLimit
		if v := q.Get("limit"); v != "" {
			n, err := strconv.Atoi(v)
			if err != nil || n <= 0 || n > maxLeaderboardLimit {
				response.WriteError(w, http.StatusBadRequest, "invalid_limit", "limit must be a positive integer up to 200")
				return
			}
			limit = n
		}

		rows, err := queries.GetLeaderboard(r.Context(), pool, suite, q.Get("sort"), limit)
		if err != nil {
			response.WriteError(w, http.StatusInternalServerError, "lookup_failed", "failed to load leaderboard")
			return
		}

		response.WriteOK(w, http.StatusOK, rows)
	}
}
