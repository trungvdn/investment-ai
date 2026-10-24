package main

import (
	"context"
	"flag"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/providers"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/service"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/store"
	"log/slog"
	"os"
	"time"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	source := flag.String("source", "", "collect only this configured provider ID")
	due := flag.Bool("due", false, "make the selected source due now, without overriding active leases")
	flag.Parse()
	if *due && *source == "" {
		slog.Error("-due requires -source")
		os.Exit(1)
	}
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
	ids := []string{}
	for _, p := range svc.Providers {
		ids = append(ids, p.ID())
	}
	schedule, err := service.ParseSchedule(os.Getenv("INGEST_SCHEDULE"), ids)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	if *source != "" {
		selected := []providers.Provider{}
		for _, p := range svc.Providers {
			if p.ID() == *source {
				selected = append(selected, p)
			}
		}
		if len(selected) == 0 {
			slog.Error("unknown source ID")
			os.Exit(1)
		}
		svc.Providers = selected
		if *due {
			if err = repo.MakeSourceDue(ctx, *source); err != nil {
				slog.Error("source cannot be made due; it may already be running")
				os.Exit(1)
			}
		}
	}
	err = svc.Tick(ctx, repo, schedule, 6*time.Hour)
	if err != nil {
		slog.Error("ingestion failed")
		os.Exit(1)
	}
	states, err := repo.SourceStates(ctx)
	if err != nil {
		os.Exit(1)
	}
	failed := 0
	for _, st := range states {
		if st.Kind != "" && (*source == "" || st.Provider == *source) {
			failed++
		}
	}
	slog.Info("scheduled ingestion finished", "sourcesWithErrors", failed)
	if failed > 0 {
		os.Exit(2)
	}
}
