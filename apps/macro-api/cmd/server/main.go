package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/trungvdn/investment-ai/apps/macro-api/internal/api/handler"
	app "github.com/trungvdn/investment-ai/apps/macro-api/internal/application/macrodata"
	infra "github.com/trungvdn/investment-ai/apps/macro-api/internal/infrastructure/macrodata"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	mode := os.Getenv("MACRO_DATA_MODE")
	if mode == "" {
		mode = "mock"
	}

	repo := infra.NewInMemoryRepository()
	provider := infra.NewMockProvider()
	normalizer := infra.NewNormalizer()

	ingestionSvc := app.NewIngestionService(provider, normalizer, repo)
	querySvc := app.NewQueryService(repo)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, querySvc, ingestionSvc)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		log.Printf("server listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen:%v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown failed:%+v", err)
	}
	log.Println("server exited")
}
