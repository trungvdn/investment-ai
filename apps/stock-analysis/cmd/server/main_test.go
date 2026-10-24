package main

import (
	"context"
	"encoding/json"
	"github.com/trungvdn/investment-ai/apps/stock-analysis/internal/wyckoff"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type fakeProvider struct{ candles []wyckoff.Candle }

func (p fakeProvider) FetchDaily(context.Context, string, time.Time, time.Time) ([]wyckoff.Candle, error) {
	return p.candles, nil
}

func TestReportChartUsesAnalyzedSeries(t *testing.T) {
	input := demoCandles()
	handler := newHandler(fakeProvider{input}, false, "")
	for _, frame := range []string{"1D", "1W"} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest("GET", "/api/stocks/FPT/report?timeframe="+frame, nil))
		if recorder.Code != 200 {
			t.Fatalf("%s: %d %s", frame, recorder.Code, recorder.Body.String())
		}
		var payload struct {
			Symbol    string           `json:"symbol"`
			Timeframe string           `json:"timeframe"`
			Candles   []wyckoff.Candle `json:"candles"`
			Quality   struct {
				Count int `json:"dataPoints"`
			} `json:"dataQuality"`
		}
		if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
			t.Fatal(err)
		}
		if payload.Symbol != "FPT" || payload.Timeframe != frame || len(payload.Candles) != payload.Quality.Count || len(payload.Candles) < 60 {
			t.Fatalf("inconsistent report: %+v", payload)
		}
	}
}
func TestInvalidRequests(t *testing.T) {
	handler := newHandler(fakeProvider{demoCandles()}, false, "")
	for _, path := range []string{"/api/stocks/FPT/report?timeframe=1H", "/api/stocks/FPT/report?from=bad", "/api/stocks/FPT/report?from=2026-10-01&to=2026-01-01", "/api/stocks/FPT%3C/report"} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest("GET", path, nil))
		if recorder.Code != 400 {
			t.Fatalf("%s: expected 400, got %d", path, recorder.Code)
		}
	}
}
func TestWeeklyOHLCVAndInclusiveDates(t *testing.T) {
	candle := func(date string, o, h, l, c, v float64) wyckoff.Candle {
		d, _ := time.Parse("2006-01-02", date)
		return wyckoff.Candle{Date: d, Open: o, High: h, Low: l, Close: c, Volume: v}
	}
	a := candle("2025-12-29", 10, 13, 9, 12, 100)
	b := candle("2026-01-02", 12, 15, 11, 14, 200)
	c := candle("2026-01-05", 14, 18, 13, 17, 300)
	bars := prepareCandles([]wyckoff.Candle{c, b, a, b}, a.Date, c.Date, "1W")
	if len(bars) != 2 || bars[0].Open != 10 || bars[0].High != 15 || bars[0].Low != 9 || bars[0].Close != 14 || bars[0].Volume != 300 || !bars[0].Date.Equal(b.Date) {
		t.Fatalf("incorrect weekly aggregation: %+v", bars)
	}
	bars = prepareCandles([]wyckoff.Candle{a, b, c}, b.Date, b.Date, "1D")
	if len(bars) != 1 || !bars[0].Date.Equal(b.Date) {
		t.Fatal("date boundaries must be inclusive")
	}
}

func TestReactRoutesAndMissingAssets(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<div id=\"root\">React entry</div>"), 0600); err != nil {
		t.Fatal(err)
	}
	handler := newHandler(fakeProvider{}, false, dir)
	for _, path := range []string{"/", "/macro", "/macro.html"} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest("GET", path, nil))
		if recorder.Code != 200 || !strings.Contains(recorder.Body.String(), "React entry") {
			t.Fatalf("%s must serve React entry: %d", path, recorder.Code)
		}
	}
	for _, path := range []string{"/assets/missing.js", "/api/stocks/FPT/missing"} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest("GET", path, nil))
		if recorder.Code != 404 || strings.Contains(recorder.Body.String(), "React entry") {
			t.Fatalf("%s must stay 404", path)
		}
	}
}
