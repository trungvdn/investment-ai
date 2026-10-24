package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/trungvdn/investment-ai/apps/stock-analysis/internal/marketdata"
	"github.com/trungvdn/investment-ai/apps/stock-analysis/internal/report"
	"github.com/trungvdn/investment-ai/apps/stock-analysis/internal/wyckoff"
)

func newHandler(provider marketdata.Provider, demo bool, uiDir string) http.Handler {
	analyzer := wyckoff.NewAnalyzer(wyckoff.DefaultConfig())
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
		if !regexp.MustCompile(`^[A-Z][A-Z0-9]{1,9}$`).MatchString(symbol) {
			http.Error(w, "invalid stock symbol", http.StatusBadRequest)
			return
		}

		timeframe := r.URL.Query().Get("timeframe")
		if timeframe == "" {
			timeframe = "1D"
		}
		if timeframe != "1D" && timeframe != "1W" {
			http.Error(w, "timeframe must be 1D or 1W", http.StatusBadRequest)
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
		if demo {
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

		candles = prepareCandles(candles, from, to, timeframe)
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
			payload := report.Build(result)
			_ = json.NewEncoder(w).Encode(struct {
				report.Report
				Candles   []wyckoff.Candle `json:"candles"`
				Timeframe string           `json:"timeframe"`
			}{payload, candles, timeframe})
			return
		}
		_ = json.NewEncoder(w).Encode(result)
	})

	if uiDir != "" {
		files := http.FileServer(http.Dir(uiDir))
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// BrowserRouter routes must resolve to the built React entry point.
			// Asset and API misses retain 404 responses rather than HTML fallbacks.
			if r.URL.Path == "/macro" || r.URL.Path == "/macro.html" {
				cloned := r.Clone(r.Context())
				cloned.URL.Path = "/"
				files.ServeHTTP(w, cloned)
				return
			}
			files.ServeHTTP(w, r)
		})
	}
	return mux
}

func main() {
	uiDir := os.Getenv("DASHBOARD_DIR")
	if uiDir == "" {
		uiDir = "../macro-dashboard/dist"
	}
	mux := newHandler(marketdata.NewKBSSource(), strings.EqualFold(os.Getenv("MARKET_DATA_PROVIDER"), "demo"), uiDir)
	addr := os.Getenv("PORT")
	if addr == "" {
		addr = "8081"
	}
	log.Printf("stock-analysis API listening on :%s", addr)
	log.Fatal(http.ListenAndServe(":"+addr, mux))
}

func demoCandles() []wyckoff.Candle {
	out := make([]wyckoff.Candle, 0, 550)
	now := time.Now().UTC()
	base := time.Date(now.Year()-2, now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	for day, i := base, 0; !day.After(now); day = day.AddDate(0, 0, 1) {
		if day.Weekday() == time.Saturday || day.Weekday() == time.Sunday {
			continue
		}
		close := 100 + float64((i%9)-4)*0.45
		out = append(out, wyckoff.Candle{Date: day, Open: close - .15, High: close + .65, Low: close - .65, Close: close, Volume: 100000 + float64(i%4)*5000})
		i++
	}
	return out
}
