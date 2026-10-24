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
		if err != nil || interval < 5*time.Minute || interval > 7*24*time.Hour {
			logger.Error("INGEST_INTERVAL must be between 5m and 168h")
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
	ids := []string{}
	for _, p := range svc.Providers {
		ids = append(ids, p.ID())
	}
	schedule, err := service.ParseSchedule(os.Getenv("INGEST_SCHEDULE"), ids)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	workerDone := make(chan struct{})
	go func() {
		defer close(workerDone)
		for {
			if err := svc.Tick(ctx, repo, schedule, interval); err != nil && !errors.Is(err, context.Canceled) {
				logger.Error("macro scheduler failure")
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Minute):
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
