package httpapi

import (
	"encoding/json"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/macro"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/service"
	"net/http"
	"strings"
	"time"
)

func write(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
func New(s *service.Service) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { write(w, 200, map[string]string{"status": "ok"}) })
	mux.HandleFunc("GET /api/macro/operations", func(w http.ResponseWriter, r *http.Request) {
		result, err := s.Operations(r.Context(), time.Now().UTC())
		if err != nil {
			write(w, 503, map[string]string{"error": "STORAGE_UNAVAILABLE"})
			return
		}
		write(w, 200, result)
	})
	mux.HandleFunc("GET /api/macro/indicators", func(w http.ResponseWriter, r *http.Request) {
		write(w, 200, map[string]any{"indicators": macro.Catalog()})
	})
	mux.HandleFunc("GET /api/macro/dashboard", func(w http.ResponseWriter, r *http.Request) {
		result, err := s.Dashboard(r.Context(), time.Now().UTC())
		if err != nil {
			write(w, 503, map[string]string{"error": "STORAGE_UNAVAILABLE", "message": "Macro storage is unavailable"})
			return
		}
		write(w, 200, result)
	})
	mux.HandleFunc("GET /api/macro/observations/{code}", func(w http.ResponseWriter, r *http.Request) {
		code := r.PathValue("code")
		ind, ok := macro.Find(code)
		if !ok {
			write(w, 404, map[string]string{"error": "INDICATOR_NOT_FOUND"})
			return
		}
		var from, to *time.Time
		for key, dest := range map[string]**time.Time{"from": &from, "to": &to} {
			value := r.URL.Query().Get(key)
			if value != "" {
				date, err := time.Parse("2006-01-02", value)
				if err != nil {
					write(w, 400, map[string]string{"error": "INVALID_DATE"})
					return
				}
				*dest = &date
			}
		}
		if from != nil && to != nil && from.After(*to) {
			write(w, 400, map[string]string{"error": "INVALID_DATE_RANGE"})
			return
		}
		revisions := r.URL.Query().Get("revisions")
		if revisions != "" && revisions != "true" && revisions != "false" {
			write(w, 400, map[string]string{"error": "INVALID_REVISIONS_PARAMETER"})
			return
		}
		observations, err := s.Repo.History(r.Context(), code, from, to, revisions == "true")
		if err != nil {
			write(w, 503, map[string]string{"error": "STORAGE_UNAVAILABLE"})
			return
		}
		write(w, 200, map[string]any{"indicator": ind, "observations": observations})
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/macro/") && r.URL.Path != "/healthz" {
			http.NotFound(w, r)
			return
		}
		mux.ServeHTTP(w, r)
	})
}
