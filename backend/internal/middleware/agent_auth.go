package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
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

var ErrInvalidAgentToken = errors.New("invalid api key")

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

			agent, err := VerifyAgentToken(r.Context(), pool, token)
			if err != nil {
				response.WriteError(w, http.StatusUnauthorized, "invalid_token", "invalid api key")
				return
			}

			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), AgentContextKey, agent)))
		})
	}
}

// VerifyAgentToken validates a plaintext "ab_..." key and returns the
// associated agent, using the same 60s cache as the AgentAuth middleware.

func VerifyAgentToken(ctx context.Context, pool *pgxpool.Pool, token string) (*queries.Agent, error) {
	cacheKey := hashToken(token)
	if cached, ok := agentAuthCache.Load(cacheKey); ok {
		entry := cached.(cachedAgentAuth)
		if time.Now().Before(entry.expiresAt) {
			return entry.agent, nil
		}
		agentAuthCache.Delete(cacheKey)
	}

	agentID, secret, ok := apikey.Parse(token)
	if !ok {
		return nil, ErrInvalidAgentToken
	}

	hash, err := queries.GetAgentAPIKeyHash(ctx, pool, agentID)
	if err != nil || !apikey.Verify(hash, secret) {
		return nil, ErrInvalidAgentToken
	}

	agent, err := queries.GetAgentByID(ctx, pool, agentID)
	if err != nil {
		return nil, ErrInvalidAgentToken
	}

	agentAuthCache.Store(cacheKey, cachedAgentAuth{agent: agent, expiresAt: time.Now().Add(agentAuthCacheTTL)})
	return agent, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func AgentFromContext(ctx context.Context) (*queries.Agent, bool) {
	v, ok := ctx.Value(AgentContextKey).(*queries.Agent)
	return v, ok
}

func BearerToken(r *http.Request) string {
	return bearerToken(r)
}
