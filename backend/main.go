package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"github.com/ritankarsaha/agentbench/backend/internal/config"
	"github.com/ritankarsaha/agentbench/backend/internal/db"
	"github.com/ritankarsaha/agentbench/backend/internal/handlers"
	appmw "github.com/ritankarsaha/agentbench/backend/internal/middleware"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(appmw.CORS(cfg.AppEnv, cfg.FrontendURL))

	r.Get("/health", handlers.Health(pool))

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/agents", func(r chi.Router) {
			r.With(appmw.UserAuth(cfg.SupabaseURL)).Post("/register", handlers.RegisterAgent(pool))

			r.Group(func(r chi.Router) {
				r.Use(appmw.AgentAuth(pool))
				r.Get("/me", handlers.GetMe)
				r.Put("/me", handlers.UpdateMe(pool))
				r.Delete("/me", handlers.DeleteMe(pool))
			})

			r.Get("/", handlers.ListAgents(pool))
			r.Get("/{handle}", handlers.GetAgentByHandle(pool))
		})
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("agentbench backend listening on :%s (env=%s)", cfg.Port, cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
