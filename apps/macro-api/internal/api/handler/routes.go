package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	app "github.com/trungvdn/investment-ai/apps/macro-api/internal/application/macrodata"
)

type APIHandler struct {
	query  app.MacroDataQueryService
	ingest *app.MacroDataIngestionService
}

func RegisterRoutes(mux *http.ServeMux, query *app.MacroDataQueryService, ingest *app.MacroDataIngestionService) {
	h := &APIHandler{query: *query, ingest: ingest}
	mux.HandleFunc("/api/macro/indicators", h.handleIndicators)
	mux.HandleFunc("/api/macro/observations", h.handleObservations)
}

func (h *APIHandler) handleIndicators(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) == 5 && parts[4] != "" { // /api/macro/indicators/{code}
		code := parts[4]
		for _, ind := range h.query.ListIndicators(context.Background()) {
			if ind.Code == code {
				json.NewEncoder(w).Encode(map[string]interface{}{"indicator": ind})
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": map[string]string{"code": "INDICATOR_NOT_FOUND", "message": "indicator not found"}})
		return
	}
	// list
	json.NewEncoder(w).Encode(map[string]interface{}{"indicators": h.query.ListIndicators(context.Background())})
}

func parseDatePtr(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (h *APIHandler) handleObservations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	q := r.URL.Query()
	indicators := q.Get("indicators")
	from := q.Get("from")
	to := q.Get("to")
	if indicators != "" {
		codes := strings.Split(indicators, ",")
		fromT, _ := parseDatePtr(from)
		toT, _ := parseDatePtr(to)
		res, _ := h.query.QueryByIndicators(context.Background(), codes, fromT, toT)
		json.NewEncoder(w).Encode(map[string]interface{}{"observations": res})
		return
	}
	// expect /api/macro/observations/{code}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) == 5 && parts[4] != "" {
		code := parts[4]
		fromT, _ := parseDatePtr(from)
		toT, _ := parseDatePtr(to)
		obs, _ := h.query.GetHistorical(context.Background(), code, fromT, toT)
		latest, _ := h.query.GetLatest(context.Background(), code)
		json.NewEncoder(w).Encode(map[string]interface{}{"observations": obs, "latest": latest})
		return
	}
	// trigger ingestion for now
	go h.ingest.Run(context.Background())
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "ingestion started"})
}
