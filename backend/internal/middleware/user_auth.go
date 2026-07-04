package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"

	"github.com/ritankarsaha/agentbench/backend/internal/response"
)

type contextKey string

const UserIDContextKey contextKey = "user_id"

const jwksTTL = 10 * time.Minute

type jwksCache struct {
	mu        sync.RWMutex
	set       jwk.Set
	fetchedAt time.Time
	url       string
}

func newJWKSCache(url string) *jwksCache {
	return &jwksCache{url: url}
}

func (c *jwksCache) get(ctx context.Context) (jwk.Set, error) {
	c.mu.RLock()
	set, fetchedAt := c.set, c.fetchedAt
	c.mu.RUnlock()

	if set != nil && time.Since(fetchedAt) < jwksTTL {
		return set, nil
	}
	return c.refresh(ctx)
}

func (c *jwksCache) refresh(ctx context.Context) (jwk.Set, error) {
	set, err := jwk.Fetch(ctx, c.url)
	if err != nil {
		return nil, fmt.Errorf("jwks: fetch failed: %w", err)
	}
	c.mu.Lock()
	c.set = set
	c.fetchedAt = time.Now()
	c.mu.Unlock()
	return set, nil
}

// UserAuth validates a Supabase-issued human JWT via the project's JWKS
// endpoint (asymmetric signing keys — no shared secret involved).
func UserAuth(supabaseURL string) func(http.Handler) http.Handler {
	cache := newJWKSCache(supabaseURL + "/auth/v1/.well-known/jwks.json")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := bearerToken(r)
			if token == "" {
				response.WriteError(w, http.StatusUnauthorized, "missing_token", "missing bearer token")
				return
			}

			userID, err := verifyUserToken(r.Context(), cache, token)
			if err != nil {
				response.WriteError(w, http.StatusUnauthorized, "invalid_token", "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func verifyUserToken(ctx context.Context, cache *jwksCache, token string) (string, error) {
	set, err := cache.get(ctx)
	if err != nil {
		return "", err
	}

	parsed, err := jwt.Parse([]byte(token), jwt.WithKeySet(set))
	if err != nil {
		// Could be key rotation (unknown kid) — refresh once and retry.
		refreshed, refreshErr := cache.refresh(ctx)
		if refreshErr != nil {
			return "", err
		}
		parsed, err = jwt.Parse([]byte(token), jwt.WithKeySet(refreshed))
		if err != nil {
			return "", err
		}
	}

	sub, ok := parsed.Subject()
	if !ok || sub == "" {
		return "", fmt.Errorf("jwt: missing subject claim")
	}
	return sub, nil
}

func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(h, prefix) {
		return ""
	}
	return strings.TrimPrefix(h, prefix)
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(UserIDContextKey).(string)
	return v, ok
}
