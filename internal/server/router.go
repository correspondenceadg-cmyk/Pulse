package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/correspondenceadg-cmyk/pulse/internal/auth"
	"github.com/correspondenceadg-cmyk/pulse/internal/config"
	"github.com/correspondenceadg-cmyk/pulse/internal/events"
	"github.com/correspondenceadg-cmyk/pulse/internal/httpx"
	"github.com/correspondenceadg-cmyk/pulse/internal/polls"
	"github.com/correspondenceadg-cmyk/pulse/internal/votes"
	"github.com/correspondenceadg-cmyk/pulse/internal/web"
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
	pollHandlers *polls.Handlers,
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

	r.Get("/debug/embed", web.DebugHandler().ServeHTTP)

	adapter := authAdapter{authSvc}
	optional := func(next http.Handler) http.Handler { return httpx.OptionalAuth(adapter, next) }
	required := func(next http.Handler) http.Handler { return httpx.RequireAuth(adapter, next) }

	r.Route("/api", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandlers.Register)
			r.Post("/login", authHandlers.Login)
			r.Post("/refresh", authHandlers.Refresh)
			r.Post("/logout", authHandlers.Logout)
		})

		r.Route("/events", func(r chi.Router) {
			r.Get("/", eventHandlers.List)
			r.Get("/{id}", eventHandlers.Get)
			r.With(optional).Get("/{id}/stats", voteHandlers.Stats)
			r.With(optional).Get("/{id}/polls", pollHandlers.List)

			r.Group(func(r chi.Router) {
				r.Use(required)
				r.Post("/", eventHandlers.Create)
				r.Put("/{id}", eventHandlers.Update)
				r.Delete("/{id}", eventHandlers.Delete)
				r.Post("/{id}/vote", voteHandlers.Vote)
				r.Post("/{id}/polls", pollHandlers.Create)
			})
		})

		r.Route("/polls", func(r chi.Router) {
			r.With(optional).Get("/{id}/results", pollHandlers.Results)

			r.Group(func(r chi.Router) {
				r.Use(required)
				r.Post("/{id}/open", pollHandlers.Open)
				r.Post("/{id}/close", pollHandlers.Close)
				r.Post("/{id}/vote", pollHandlers.Vote)
			})
		})

		r.NotFound(func(w http.ResponseWriter, r *http.Request) {
			httpx.Error(w, r, http.StatusNotFound, "not found")
		})
	})

	spa, err := web.Handler()
	if err != nil {
		slog.Error("spa unavailable", "err", err)
		spa = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			httpx.Error(w, r, http.StatusNotFound, "not found")
		})
	}

	r.Handle("/*", spa)
	r.Handle("/", spa)

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
