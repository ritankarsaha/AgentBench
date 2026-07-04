package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ritankarsaha/agentbench/backend/internal/response"
)

func Health(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			response.WriteError(w, http.StatusServiceUnavailable, "db_unreachable", err.Error())
			return
		}

		response.WriteOK(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}
