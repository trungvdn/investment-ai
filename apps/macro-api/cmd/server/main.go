package main

import (
	"context"
	"errors"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/httpapi"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/service"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/store"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if os.Getenv("DATABASE_URL") == "" && os.Getenv("PGHOST") == "" {
		logger.Error("DATABASE_URL or PostgreSQL PGHOST settings are required; no in-memory fallback")
		os.Exit(1)
	}
	interval := 6 * time.Hour
	if v := os.Getenv("INGEST_INTERVAL"); v != "" {
		var err error
		interval, err = time.ParseDuration(v)
		if err != nil || interval < 5*time.Minute {
			logger.Error("INGEST_INTERVAL must be at least 5m")
			os.Exit(1)
		}
	}
	startup, cancel := context.WithTimeout(ctx, 20*time.Second)
	repo, err := store.Open(startup, os.Getenv("DATABASE_URL"))
	if err != nil {
		cancel()
		logger.Error("database connection failed")
		os.Exit(1)
	}
	if err = repo.Migrate(startup); err != nil {
		cancel()
		repo.Close()
		logger.Error("database migration failed")
		os.Exit(1)
	}
	cancel()
	defer repo.Close()
	svc := &service.Service{Repo: repo, Providers: service.ProvidersFromEnv(), Logger: logger}
	workerDone := make(chan struct{})
	go func() {
		defer close(workerDone)
		for {
			runCtx, cancel := context.WithTimeout(ctx, 8*time.Minute)
			result, err := svc.Ingest(runCtx)
			cancel()
			if err != nil && !errors.Is(err, context.Canceled) {
				logger.Error("macro ingestion storage/run failure")
			} else {
				logger.Info("macro ingestion finished", "processed", result.Saved, "failedSources", result.Failed)
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(interval):
			}
		}
	}()
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	server := &http.Server{Addr: ":" + port, Handler: httpapi.New(svc), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	go func() {
		logger.Info("macro API listening", "port", port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("HTTP server failed")
			stop()
		}
	}()
	<-ctx.Done()
	shutdown, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdown)
	<-workerDone
}
