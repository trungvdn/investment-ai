# Stock Analysis — Wyckoff MVP

Go service for explainable Wyckoff phase/event analysis from daily OHLCV.

## Run and test

```bash
cd apps/stock-analysis
go test ./...
go run ./cmd/server
```

`GET /api/stocks/VCI/wyckoff` returns the raw, structured Wyckoff analyzer result. `GET /api/stocks/VCI/report` returns the standardized, presentation-oriented report. Both fetch the last two years of daily candles from the KB Securities public history endpoint. Optional query parameters: `from=YYYY-MM-DD` and `to=YYYY-MM-DD`. `GET /healthz` is the health check. Set `PORT` to change the port (default 8081).

## Standardized analysis report (schema v1.0)

The report endpoint returns a stable JSON contract with `schemaVersion`, `symbol`, `asOf`, `headline`, `summary`, `structure`, `phase`, `confidence`, `tradingRange`, `events`, `evidence`, `risks`, `watchlist`, `dataQuality`, and `disclaimer`.

- `headline` and `summary` provide a deterministic human-readable interpretation of the current analyzer classification.
- `tradingRange.status` distinguishes `IDENTIFIED` from `PROVISIONAL`; a provisional range is not a confirmed structure.
- `events` retains the event dates, confidence scores, and evidence text from the analyzer.
- `evidence` preserves the analyzer's evidence statements; `risks` highlights uncertainty, low confidence, missing range, or demo data.
- `watchlist` contains follow-up conditions to observe, not buy/sell instructions.
- `dataQuality` identifies the data mode and history count; demo data is explicitly marked.
- `disclaimer` states that the output is for research and is not investment advice.

The raw `/wyckoff` endpoint remains available so existing consumers are not forced to migrate immediately.

For local deterministic testing only, set `MARKET_DATA_PROVIDER=demo`. Real market data is the default; provider errors return HTTP 502 instead of silently falling back to mock data.

## Wyckoff rules implemented (initial heuristic)

- Context: preceding directional move and recent trading range.
- Trading Range: evaluate a bounded recent window excluding the latest five candles, with width, close overlap, repeated reactions near both edges, and preceding trend context. A breakout can then be evaluated against the range it is leaving.
- Accumulation event sequence: PS → SC → AR → ST is tracked chronologically as supporting evidence; it is not treated as mandatory for every valid schematic.
- Spring/Test: Spring requires penetration and recovery of a detected range boundary; Test must follow within 15 sessions, hold above the Spring low, and show reduced supply.
- SOS/LPS: SOS requires a close above the candidate range with strong close location and expanding relative volume; LPS is only considered as a pullback after SOS within the active range window.
- Event reporting is scoped to the most recent 70 candles (roughly a trading quarter) so events from prior cycles do not dominate the current analysis.
- LPS/LPSY use local historical breakout levels and volume baselines, and are restricted to the recent event window.
- Phase assignment is gated by the interpreted structure; an `UNCLEAR` structure remains `UNKNOWN` rather than being assigned Phase D from an isolated LPS/LPSY event.
- Events: candidate Selling Climax (SC), Buying Climax (BC), Automatic Rally/Reaction (AR), Spring, UTAD, SOS, SOW and low-volume Spring Test.
- Accumulation: normally follows markdown; Phase A looks for stopping action/climax and AR/ST, B is range/cause building, C is Spring + Test, D is SOS/LPS and E is markup.
- Distribution: normally follows markup; Phase A looks for buying climax and reaction, B is range/cause building, C is UT/UTAD, D is SOW/LPSY and E is markdown.
- Price/volume checks include relative volume, candle spread, close location, support/resistance penetration, and directional context.
- If evidence conflicts or is insufficient, phase can remain UNKNOWN. Wait for confirmation rather than anticipate.

## Data source and limitations

The built-in KB Securities history endpoint is an unofficial third-party interface and may change or become unavailable. Prices from this endpoint are scaled down by 1,000 as documented in the community integration example. For production, use a licensed feed or official TCBS OpenAPI (requires credentials). Confirm data adjustment policy and corporate-action handling before historical analysis.

This is an MVP heuristic, not a calibrated model or investment recommendation. Phase detection is subjective and needs manual labels and historical backtesting. Confidence is a rule-strength score, not a probability.
