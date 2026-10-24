package httpapi

import (
	"encoding/json"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/service"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/testutil"
	"net/http/httptest"
	"testing"
)

func TestReadOnlyDashboardAndDateValidation(t *testing.T) {
	h := New(&service.Service{Repo: &testutil.Memory{}})
	r := httptest.NewRecorder()
	h.ServeHTTP(r, httptest.NewRequest("GET", "/api/macro/dashboard", nil))
	if r.Code != 200 || r.Header().Get("Content-Type") != "application/json; charset=utf-8" {
		t.Fatal("JSON dashboard unavailable")
	}
	var d service.Dashboard
	if err := json.Unmarshal(r.Body.Bytes(), &d); err != nil || d.Score != nil || len(d.Indicators) != 18 {
		t.Fatal("empty live dashboard must be explicit")
	}
	for _, path := range []string{"/api/macro/observations/VN_CPI_YOY?from=no", "/api/macro/observations/VN_CPI_YOY?from=2026-12-01&to=2026-01-01", "/api/macro/observations/VN_CPI_YOY?revisions=garbage"} {
		r = httptest.NewRecorder()
		h.ServeHTTP(r, httptest.NewRequest("GET", path, nil))
		if r.Code != 400 {
			t.Fatalf("%s expected 400, got %d", path, r.Code)
		}
	}
	for _, path := range []string{"/api/macro/observations/unknown", "/api/macro/observations", "/api/macro/ingest"} {
		r = httptest.NewRecorder()
		h.ServeHTTP(r, httptest.NewRequest("GET", path, nil))
		if r.Code != 404 {
			t.Fatalf("unexpected side-effect route %s", path)
		}
	}
	r = httptest.NewRecorder()
	h.ServeHTTP(r, httptest.NewRequest("POST", "/api/macro/dashboard", nil))
	if r.Code != 405 {
		t.Fatal("dashboard must be read-only")
	}
}

func TestOperationsEndpointIsReadOnlyAndReportsCoverage(t *testing.T) {
	repo := &testutil.Memory{}
	h := New(&service.Service{Repo: repo})
	r := httptest.NewRecorder()
	h.ServeHTTP(r, httptest.NewRequest("GET", "/api/macro/operations", nil))
	var result service.Operations
	if r.Code != 200 || json.Unmarshal(r.Body.Bytes(), &result) != nil || len(result.Alerts) != 18 || len(repo.Raw) != 0 {
		t.Fatal("operations did not report missing coverage without ingestion")
	}
	r = httptest.NewRecorder()
	h.ServeHTTP(r, httptest.NewRequest("POST", "/api/macro/operations", nil))
	if r.Code != 405 {
		t.Fatal("operations endpoint must not mutate state")
	}
}
