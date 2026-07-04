package apikey

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const prefix = "ab_"
const bcryptCost = 12
const bcryptMaxInputBytes = 72

// Generate creates a new agent API key of the form "ab_<agentID>_<secret>".
// Only the secret half is ever bcrypt-hashed or compared: the full key
// (prefix + UUID + secret) would exceed bcrypt's 72-byte input limit, and
// the agent ID is not secret — it exists purely for O(1) lookup.
func Generate(agentID uuid.UUID) (plaintext string, hash string, err error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", fmt.Errorf("apikey: failed to generate secret: %w", err)
	}
	secret := hex.EncodeToString(raw)
	plaintext = fmt.Sprintf("%s%s_%s", prefix, agentID.String(), secret)

	hashed, err := bcrypt.GenerateFromPassword([]byte(secret), bcryptCost)
	if err != nil {
		return "", "", fmt.Errorf("apikey: failed to hash secret: %w", err)
	}
	return plaintext, string(hashed), nil
}

// Parse splits a plaintext key into its agent ID and secret.
func Parse(token string) (agentID uuid.UUID, secret string, ok bool) {
	rest, found := strings.CutPrefix(token, prefix)
	if !found {
		return uuid.Nil, "", false
	}
	idPart, secretPart, found := strings.Cut(rest, "_")
	if !found || secretPart == "" {
		return uuid.Nil, "", false
	}
	id, err := uuid.Parse(idPart)
	if err != nil {
		return uuid.Nil, "", false
	}
	return id, secretPart, true
}

// Verify checks a plaintext secret (as returned by Parse) against its bcrypt hash.
func Verify(hash, secret string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(secret)) == nil
}

// HashArbitrary bcrypt-hashes an opaque secret of unknown provenance (e.g. an
// externally-issued API key). Inputs over bcrypt's 72-byte limit are
// pre-hashed with SHA-256 to avoid silent truncation or a hard error.
func HashArbitrary(secret string) (string, error) {
	input := []byte(secret)
	if len(input) > bcryptMaxInputBytes {
		sum := sha256.Sum256(input)
		input = sum[:]
	}
	hashed, err := bcrypt.GenerateFromPassword(input, bcryptCost)
	if err != nil {
		return "", fmt.Errorf("apikey: failed to hash secret: %w", err)
	}
	return string(hashed), nil
}
