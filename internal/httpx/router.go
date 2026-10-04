package httpx

import (
	"net/http"

	"github.com/correspondenceadg-cmyk/pulse/internal/config"
	"github.com/go-chi/chi/v5"
)

func NewRouter(cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	r.Use(RequestID)
	r.Use(Recoverer)
	r.Use(Logger)
	r.Use(SecurityHeaders)

	r.Get("/actuator/health", func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{"status": "UP"})
	})

	r.Route("/api", func(r chi.Router) {
		// auth, events, votes, polls, audit wired here as we build them
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		// Will serve the built React SPA for non-API paths once frontend is bundled
		Error(w, r, http.StatusNotFound, "not found")
	})

	return r
}