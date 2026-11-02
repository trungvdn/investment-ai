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

## Research UI and chart series

The service serves the built React app from `../macro-dashboard/dist` (override with `DASHBOARD_DIR`). First run `npm ci --prefix apps/macro-dashboard` and `npm run build` from the repository root, then start the Go service from this directory. Open `http://localhost:8081`; the React macro dashboard is at `/macro`, with a compatibility redirect from `/macro.html`. Development uses Vite with a stock API proxy; see the frontend README.

The `/report` response adds `candles` (the exact OHLCV observations analyzed) and `timeframe` (`1D` by default, or `1W`). Both analysis endpoints accept `timeframe=1D|1W`. Date filtering is inclusive, observations are ordered/deduplicated, and weekly OHLCV is aggregated by ISO week **before** analysis. Weekly bar dates identify the last observed session; the latest week may be incomplete. At least 60 observations are still required, so use two years for weekly analysis. Existing report fields and the raw analyzer response shape remain available.

Macro API integration: `/api/macro/` is reverse-proxied to `MACRO_API_URL` (default `http://127.0.0.1:3000`). Start the PostgreSQL-backed macro service separately; an unavailable macro service returns a JSON 502 and never falls back to the React entry point. No FRED credentials belong in this stock server or frontend.

## Reviewed annotation benchmark

`go test ./internal/wyckoff -run 'TestReviewed|TestSampleBaseline' -count=1 -v` reads the nine approved annotation JSONs and their colocated CSVs. Positive reviewed hypotheses, rejected proposed labels and uncertain labels are separate. Replay passes only OHLCV prefixes into the analyzer; annotations are scoring targets, never analyzer inputs. Matching uses the reviewed date interval. A match at any eligible replay cutoff and a match at the final cutoff are reported separately.

Set `WYCKOFF_BENCHMARK_OUTPUT=/tmp/result.json` to save the detailed result. CI compares the pinned pre-fix main analyzer (`525120c`, the completed Spring/Test proposal before lifecycle changes) with the candidate using identical annotations and prints a comparison in the job summary. Sparse labels do not establish global precision or investment accuracy; there is no accuracy pass gate. Structure-ambiguous rejected labels and unsupported labels such as `markdown_continuation` are reported separately. Assessment disagreements are logged rather than changing human labels.

See [dated evidence checkpoints](data/annotations/CONFIRMATION_CHECKPOINTS.md). Candidate labels remain candidates; `label_confirmed_as_of` is a conservative retrospective label cutoff, not the earliest market confirmation date.

## Structure lifecycle (additive API fields)

Both raw analysis and standardized reports expose `structures` and `active_structure_id`. Each retained structure has `structure_id`, optional `parent_structure_id`, `kind` (a hypothesis), `start_date`, `identified_at`, `range_end`, `last_observed`, frozen `support`/`resistance`, `status`, `breakout_date`, transition history and structure-scoped events. `identified_at` is the date the range was first detectable on the supplied prefix; historical range start is not its availability date. `range_end` stops advancing when price leaves the range. `last_observed` continues through the requested cutoff.

The detector replays observed candles at 12, 20, 40 and 70-bar scales. It requires price overlap, multiple separated boundary reactions and limited directional drift; similar overlapping candidates are deduplicated. A narrower range can have an enclosing price/time context. A separate 70-bar swing envelope can be retained as `kind=context` without enough edge tests to qualify as a trading range; it never generates SOS/LPS or drives markup. Historical parents must overlap the child formation in time. A parent relationship is reported as of the current cutoff and can be discovered later than the child. Automatically generated IDs are independent of annotation IDs and deterministic for identical history; truncating or revising input history may change them.

A volume-backed upside breakout retains original boundaries and a candidate SOS. Sustained follow-through can produce MARKUP; an actual reduced-volume pullback within 15 bars can produce candidate LPS at the original resistance. Two closes back inside record FAILED_BREAKOUT; subsequent loss of original support records MARKDOWN while preserving the original SOS. Two closes below a range without an upside breakout record BREAKDOWN, without assuming a distribution/SOW. A preceding advance plus sustained upside escape supports the reaccumulation-to-markup hypothesis. Confidence remains heuristic.

Lifecycle state is reconstructed from each requested OHLCV history and returned in the report; no new database or cross-request persistence is introduced. Completed structures remain available under `structures` for that history. The legacy top-level event list remains limited to the last 70 candles, while each structure retains its own older events. Legacy events without defensible attribution keep no `structure_id`; consumers must not attach them to an arbitrary range. This change exposes data in the API; the frontend does not yet render a lifecycle timeline.

Tests cover frozen boundaries, retained SOS after failure, LPS using original resistance, prefix cutoffs, NLG nested range detection, MSB markup after reaccumulation, and report field propagation. The nine reviewed annotations remain unchanged. Existing benchmark interval recall is unscoped; structure-aware matching is still a separate task and cannot be inferred from that metric.
