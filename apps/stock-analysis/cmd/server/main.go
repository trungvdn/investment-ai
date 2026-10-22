package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/trungvdn/investment-ai/apps/stock-analysis/internal/wyckoff"
)

func main() {
	analyzer:=wyckoff.NewAnalyzer(wyckoff.DefaultConfig())
	mux:=http.NewServeMux()
	mux.HandleFunc("/healthz",func(w http.ResponseWriter,r *http.Request){ w.Header().Set("Content-Type","application/json"); _=json.NewEncoder(w).Encode(map[string]string{"status":"ok"}) })
	mux.HandleFunc("/api/stocks/",func(w http.ResponseWriter,r *http.Request){
		if r.Method!=http.MethodGet { http.Error(w,"method not allowed",http.StatusMethodNotAllowed); return }
		parts:=strings.Split(strings.Trim(r.URL.Path,"/"),"/")
		if len(parts)!=4 || parts[0]!="api" || parts[1]!="stocks" || parts[3]!="wyckoff" { http.NotFound(w,r); return }
		symbol:=strings.ToUpper(parts[2])
		if symbol=="" { http.Error(w,"symbol is required",http.StatusBadRequest); return }
		// Explicitly marked demo data keeps the endpoint runnable before a real
		// Vietnamese market-data provider is connected.
		candles:=demoCandles()
		result,err:=analyzer.Analyze(symbol,candles)
		if err!=nil { http.Error(w,err.Error(),http.StatusUnprocessableEntity); return }
		result.DataMode="demo"
		w.Header().Set("Content-Type","application/json")
		_ = json.NewEncoder(w).Encode(result)
	})
	addr:=":8081"
	log.Printf("stock-analysis API listening on %s (demo OHLCV only)",addr)
	log.Fatal(http.ListenAndServe(addr,mux))
}

func demoCandles() []wyckoff.Candle {
	out:=make([]wyckoff.Candle,60)
	base:=time.Date(2026,8,1,0,0,0,0,time.UTC)
	for i:=range out {
		close:=100.0+float64((i%9)-4)*0.45
		out[i]=wyckoff.Candle{Date:base.AddDate(0,0,i),Open:close-0.15,High:close+0.65,Low:close-0.65,Close:close,Volume:100000+float64(i%4)*5000}
	}
	return out
}
