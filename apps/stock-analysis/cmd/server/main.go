package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/trungvdn/investment-ai/apps/stock-analysis/internal/marketdata"
	"github.com/trungvdn/investment-ai/apps/stock-analysis/internal/report"
	"github.com/trungvdn/investment-ai/apps/stock-analysis/internal/wyckoff"
)

func main() {
	analyzer := wyckoff.NewAnalyzer(wyckoff.DefaultConfig())
	provider := marketdata.NewKBSSource()
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/api/stocks/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) != 4 || parts[0] != "api" || parts[1] != "stocks" ||
			(parts[3] != "wyckoff" && parts[3] != "report") {
			http.NotFound(w, r)
			return
		}

		symbol := strings.ToUpper(parts[2])
		if symbol == "" {
			http.Error(w, "symbol is required", http.StatusBadRequest)
			return
		}

		to := time.Now()
		from := to.AddDate(-2, 0, 0)
		if v := r.URL.Query().Get("from"); v != "" {
			parsed, err := time.Parse("2006-01-02", v)
			if err != nil {
				http.Error(w, "invalid from date; expected YYYY-MM-DD", http.StatusBadRequest)
				return
			}
			from = parsed
		}
		if v := r.URL.Query().Get("to"); v != "" {
			parsed, err := time.Parse("2006-01-02", v)
			if err != nil {
				http.Error(w, "invalid to date; expected YYYY-MM-DD", http.StatusBadRequest)
				return
			}
			to = parsed
		}
		if from.After(to) {
			http.Error(w, "from must not be after to", http.StatusBadRequest)
			return
		}

		var candles []wyckoff.Candle
		var err error
		mode := "kbsec"
		if strings.EqualFold(os.Getenv("MARKET_DATA_PROVIDER"), "demo") {
			candles = demoCandles()
			mode = "demo"
		} else {
			candles, err = provider.FetchDaily(r.Context(), symbol, from, to)
		}
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "MARKET_DATA_UNAVAILABLE",
				"message": err.Error(),
			})
			return
		}

		result, err := analyzer.Analyze(symbol, candles)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "ANALYSIS_FAILED",
				"message": err.Error(),
			})
			return
		}

		result.DataMode = mode
		w.Header().Set("Content-Type", "application/json")
		if parts[3] == "report" {
			_ = json.NewEncoder(w).Encode(report.Build(result))
			return
		}
		_ = json.NewEncoder(w).Encode(result)
	})

	// The dashboard is a separate static frontend and calls this read-only API cross-origin.
	// No cookies or credentials are used by this public market-data endpoint.
	corsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		mux.ServeHTTP(w, r)
	})

	addr := os.Getenv("PORT")
	if addr == "" {
		addr = "8081"
	}
	log.Printf("stock-analysis API listening on :%s", addr)
	log.Fatal(http.ListenAndServe(":"+addr, corsHandler))
}

func demoCandles() []wyckoff.Candle {
	out := make([]wyckoff.Candle, 90)
	base := time.Now().AddDate(0, 0, -90)
	for i := range out {
		close := 100 + float64((i%9)-4)*0.45
		out[i] = wyckoff.Candle{
			Date:   base.AddDate(0, 0, i),
			Open:   close - 0.15,
			High:   close + 0.65,
			Low:    close - 0.65,
			Close:  close,
			Volume: 100000 + float64(i%4)*5000,
		}
	}
	return out
}
