package middleware

import (
	"net/http"

	"github.com/rs/cors"
)

func CORS(appEnv, frontendURL string) func(http.Handler) http.Handler {
	allowedOrigins := []string{frontendURL}
	if appEnv != "production" {
		allowedOrigins = []string{"*"}
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "X-AgentBench-Trace-Sig"},
		AllowCredentials: appEnv == "production",
		MaxAge:           300,
	})

	return c.Handler
}
