# Stock Analysis — Wyckoff MVP

Go service for explainable Wyckoff phase/event analysis from daily OHLCV.

## Run and test

```bash
cd apps/stock-analysis
go test ./...
go run ./cmd/server
```

`GET /api/stocks/VCI/wyckoff` fetches the last two years of daily candles from the KB Securities public history endpoint. Optional query parameters: `from=YYYY-MM-DD` and `to=YYYY-MM-DD`. `GET /healthz` is the health check. Set `PORT` to change the port (default 8081).

For local deterministic testing only, set `MARKET_DATA_PROVIDER=demo`. Real market data is the default; provider errors return HTTP 502 instead of silently falling back to mock data.

## Wyckoff rules implemented (initial heuristic)

- Context: preceding directional move and recent trading range.
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
