package httpx

import (
	"context"
	"net/http"
	"time"

	"github.com/correspondenceadg-cmyk/pulse/internal/config"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(cfg *config.Config, pool *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()

	r.Use(RequestID)
	r.Use(Recoverer)
	r.Use(Logger)
	r.Use(SecurityHeaders)

	r.Get("/actuator/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			WriteJSON(w, http.StatusServiceUnavailable, map[string]string{
				"status": "DOWN",
				"db":     "unreachable",
			})
			return
		}
		WriteJSON(w, http.StatusOK, map[string]string{"status": "UP"})
	})

	r.Route("/api", func(r chi.Router) {
		// auth, events, votes, polls wired here next
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		Error(w, r, http.StatusNotFound, "not found")
	})

	return r
}