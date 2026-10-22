# Stock Analysis — Wyckoff MVP

Initial Go modular component for classifying a stock's Wyckoff market structure and phase from OHLCV candles.

## Run

```bash
cd apps/stock-analysis
go test ./...
go run ./cmd/server
```

Then request `GET http://localhost:8081/api/stocks/VCI/wyckoff` or `GET /healthz`.

## Current scope and limitations

- Rule-based, explainable heuristics; not an investment recommendation.
- Endpoint currently uses deterministic demo candles. It does **not** fetch live/historical Vietnamese stock data.
- Phase confidence is a heuristic score, not a calibrated probability.
- Wyckoff interpretation is subjective; validate with manually labeled historical examples before using for decisions.
- Next steps: provider interface, real OHLCV ingestion, stronger trading-range/event rules, and historical backtesting.
