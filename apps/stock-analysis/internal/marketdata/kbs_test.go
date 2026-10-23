package marketdata

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchDailyUsesKBSECCompatibleHeaders(t *testing.T) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("User-Agent"); got != kbsecUserAgent {
			t.Fatalf("User-Agent = %q, want %q", got, kbsecUserAgent)
		}
		if got := r.URL.Query().Get("sdate"); got != "17-08-2026" {
			t.Fatalf("sdate = %q, want 17-08-2026", got)
		}
		if got := r.URL.Query().Get("edate"); got != "22-10-2026" {
			t.Fatalf("edate = %q, want 22-10-2026", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"symbol":"VCI","data_day":[{"t":"2026-10-22 07:00","o":19500,"h":19800,"l":19350,"c":19500,"v":6999500}]}`))
	}))
	defer server.Close()

	source := &KBSSource{Client: server.Client(), BaseURL: server.URL}
	from := time.Date(2026, time.August, 17, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.October, 22, 0, 0, 0, 0, time.UTC)
	candles, err := source.FetchDaily(context.Background(), "VCI", from, to)
	if err != nil {
		t.Fatalf("FetchDaily() error = %v", err)
	}
	if len(candles) != 1 {
		t.Fatalf("len(candles) = %d, want 1", len(candles))
	}
}
