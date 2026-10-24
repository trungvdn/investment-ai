package main

import (
	"context"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/service"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/store"
	"log/slog"
	"os"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	if os.Getenv("DATABASE_URL") == "" && os.Getenv("PGHOST") == "" {
		slog.Error("PostgreSQL configuration is required")
		os.Exit(1)
	}
	repo, err := store.Open(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		slog.Error("database unavailable")
		os.Exit(1)
	}
	defer repo.Close()
	if err = repo.Migrate(ctx); err != nil {
		slog.Error("migration failed")
		os.Exit(1)
	}
	svc := service.Service{Repo: repo, Providers: service.ProvidersFromEnv(), Logger: slog.Default()}
	result, err := svc.Ingest(ctx)
	if err != nil {
		slog.Error("ingestion failed")
		os.Exit(1)
	}
	slog.Info("ingestion finished", "processed", result.Saved, "failedSources", result.Failed)
	if result.Failed > 0 {
		os.Exit(2)
	}
}
