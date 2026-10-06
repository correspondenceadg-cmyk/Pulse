package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/correspondenceadg-cmyk/pulse/internal/auth"
	"github.com/correspondenceadg-cmyk/pulse/internal/config"
	"github.com/correspondenceadg-cmyk/pulse/internal/db"
	"github.com/correspondenceadg-cmyk/pulse/internal/events"
	"github.com/correspondenceadg-cmyk/pulse/internal/server"
	"github.com/correspondenceadg-cmyk/pulse/internal/votes"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))
	slog.SetDefault(logger)

	startupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := db.NewPool(startupCtx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := db.Migrate(startupCtx, pool); err != nil {
		return err
	}

	authRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(authRepo, cfg)
	authHandlers := auth.NewHandlers(authSvc, cfg.Env == "production")

	eventRepo := events.NewRepository(pool)
	eventSvc := events.NewService(eventRepo)
	eventHandlers := events.NewHandlers(eventSvc)

	voteRepo := votes.NewRepository(pool)
	voteSvc := votes.NewService(voteRepo)
	voteHandlers := votes.NewHandlers(voteSvc)

	router := server.NewRouter(cfg, pool, authHandlers, authSvc, eventHandlers, voteHandlers)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", srv.Addr, "env", cfg.Env)
		errCh <- srv.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case sig := <-sigCh:
		slog.Info("shutdown", "signal", sig.String())
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()
	return srv.Shutdown(shutdownCtx)
}
