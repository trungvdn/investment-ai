# Investment AI Research UI

Dependency-free HTML/CSS/ES modules. The main page is Wyckoff Research; the original macro dashboard is preserved at `macro.html` with its existing mock-data labels.

## Run the complete flow

From the repository root:

```sh
cd apps/stock-analysis
go run ./cmd/server
```

Open `http://localhost:8081`. The Go service serves the UI and `/api/stocks/:symbol/report` from the same origin. No CORS configuration is required. `DASHBOARD_DIR` overrides the static directory (default `../macro-dashboard`, relative to the process working directory). When deploying the frontend separately, proxy `/api/stocks/` to this service; do not serve the HTML fallback for API routes.

Real KB Securities data is the default. An unavailable source produces an error, never a silent switch to synthetic data. To inspect UI offline, explicitly run `MARKET_DATA_PROVIDER=demo go run ./cmd/server`; every report is labeled as synthetic. No API secrets are embedded in the frontend.

## Behavior

- Enter a symbol or choose FPT/VCB/MBB/VCI; submit with Enter.
- Select daily/weekly and 6/12/24 months. Selecting weekly defaults to 24 months because the analyzer requires 60 candles; shorter weekly windows can return insufficient history.
- The report and chart use the same filtered, sorted OHLCV series. Weekly data is aggregated before analysis, not relabeled daily data. The last weekly candle may be incomplete.
- Inspect chart event markers or the full event list. Keyboard users can activate markers and navigate tabs with arrow/Home/End keys.
- Overview, events, conditional scenarios and data/methodology tabs preserve the backend's evidence, risks and limitations. No fabricated company names or confidence probabilities.
- Loading, timeout, invalid symbol, provider error, insufficient-history and unknown-structure states are explicit. Search parameters allow sharing the same analysis settings; results are recalculated on load.

## Verify

```sh
node --test apps/macro-dashboard/test/*.test.js
cd apps/stock-analysis
go test ./...
go vet ./...
```
