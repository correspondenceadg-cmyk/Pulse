package server

import (
	"context"
	"net/http"
	"time"

	"github.com/correspondenceadg-cmyk/pulse/internal/auth"
	"github.com/correspondenceadg-cmyk/pulse/internal/config"
	"github.com/correspondenceadg-cmyk/pulse/internal/events"
	"github.com/correspondenceadg-cmyk/pulse/internal/httpx"
	"github.com/correspondenceadg-cmyk/pulse/internal/votes"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(
	cfg *config.Config,
	pool *pgxpool.Pool,
	authHandlers *auth.Handlers,
	authSvc *auth.Service,
	eventHandlers *events.Handlers,
	voteHandlers *votes.Handlers,
) http.Handler {
	r := chi.NewRouter()

	r.Use(httpx.RequestID)
	r.Use(httpx.Recoverer)
	r.Use(httpx.Logger)
	r.Use(httpx.SecurityHeaders)

	r.Get("/actuator/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			httpx.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{
				"status": "DOWN",
				"db":     "unreachable",
			})
			return
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "UP"})
	})

	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/register", authHandlers.Register)
		r.Post("/login", authHandlers.Login)
		r.Post("/refresh", authHandlers.Refresh)
		r.Post("/logout", authHandlers.Logout)
	})

	adapter := authAdapter{authSvc}

	r.Route("/api/events", func(r chi.Router) {
		r.Get("/", eventHandlers.List)
		r.Get("/{id}", eventHandlers.Get)

		r.With(func(next http.Handler) http.Handler {
			return httpx.OptionalAuth(adapter, next)
		}).Get("/{id}/stats", voteHandlers.Stats)

		r.Group(func(r chi.Router) {
			r.Use(func(next http.Handler) http.Handler {
				return httpx.RequireAuth(adapter, next)
			})
			r.Post("/", eventHandlers.Create)
			r.Put("/{id}", eventHandlers.Update)
			r.Delete("/{id}", eventHandlers.Delete)
			r.Post("/{id}/vote", voteHandlers.Vote)
		})
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		httpx.Error(w, r, http.StatusNotFound, "not found")
	})

	_ = cfg
	return r
}

type authAdapter struct{ svc *auth.Service }

func (a authAdapter) ParseAccess(raw string) (*httpx.AuthClaims, error) {
	claims, err := a.svc.ParseAccess(raw)
	if err != nil {
		return nil, err
	}
	return &httpx.AuthClaims{UserID: claims.UserID, Role: claims.Role}, nil
}
