package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	app "github.com/trungvdn/investment-ai/apps/macro-api/internal/application/macrodata"
	infra "github.com/trungvdn/investment-ai/apps/macro-api/internal/infrastructure/macrodata"
)

func TestHandlersListIndicators(t *testing.T) {
	repo := infra.NewInMemoryRepository()
	provider := infra.NewMockProvider()
	normalizer := infra.NewNormalizer()
	ingest := app.NewIngestionService(provider, normalizer, repo)
	query := app.NewQueryService(repo)

	mux := http.NewServeMux()
	RegisterRoutes(mux, query, ingest)

	ts := httptest.NewServer(mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/macro/indicators")
	if err != nil {
		t.Fatalf("get err %v", err)
	}
	defer resp.Body.Close()
	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	if _, ok := body["indicators"]; !ok {
		t.Fatalf("expected indicators")
	}
}
