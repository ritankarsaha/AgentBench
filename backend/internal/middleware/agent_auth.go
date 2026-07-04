package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ritankarsaha/agentbench/backend/internal/apikey"
	"github.com/ritankarsaha/agentbench/backend/internal/db/queries"
	"github.com/ritankarsaha/agentbench/backend/internal/response"
)

const AgentContextKey contextKey = "agent"

const agentAuthCacheTTL = 60 * time.Second

type cachedAgentAuth struct {
	agent     *queries.Agent
	expiresAt time.Time
}

var agentAuthCache sync.Map

// AgentAuth validates an "ab_<agentID>_<secret>" API key. Verified results
// are cached for 60s (keyed by a hash of the token, not the plaintext) since
// bcrypt is deliberately slow and the SDK polls run status frequently.
func AgentAuth(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := bearerToken(r)
			if token == "" {
				response.WriteError(w, http.StatusUnauthorized, "missing_token", "missing bearer token")
				return
			}

			cacheKey := hashToken(token)
			if cached, ok := agentAuthCache.Load(cacheKey); ok {
				entry := cached.(cachedAgentAuth)
				if time.Now().Before(entry.expiresAt) {
					next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), AgentContextKey, entry.agent)))
					return
				}
				agentAuthCache.Delete(cacheKey)
			}

			agentID, secret, ok := apikey.Parse(token)
			if !ok {
				response.WriteError(w, http.StatusUnauthorized, "invalid_token", "invalid api key")
				return
			}

			hash, err := queries.GetAgentAPIKeyHash(r.Context(), pool, agentID)
			if err != nil || !apikey.Verify(hash, secret) {
				response.WriteError(w, http.StatusUnauthorized, "invalid_token", "invalid api key")
				return
			}

			agent, err := queries.GetAgentByID(r.Context(), pool, agentID)
			if err != nil {
				response.WriteError(w, http.StatusUnauthorized, "invalid_token", "invalid api key")
				return
			}

			agentAuthCache.Store(cacheKey, cachedAgentAuth{agent: agent, expiresAt: time.Now().Add(agentAuthCacheTTL)})
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), AgentContextKey, agent)))
		})
	}
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func AgentFromContext(ctx context.Context) (*queries.Agent, bool) {
	v, ok := ctx.Value(AgentContextKey).(*queries.Agent)
	return v, ok
}
