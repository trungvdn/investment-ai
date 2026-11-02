# Stock Analysis — context-first Wyckoff research

The default analyzer starts from the latest **40 observations**, assesses prior direction and confirmed swings, constructs a campaign, then reads its stopping/test/breakout sequence. It no longer merges independent candle labels from every range into one current structure. The former default event scanner has been removed.

Method and limits: [analyzer-method.md](../../docs/wyckoff/analyzer-method.md).

## Run

```bash
cd apps/stock-analysis
go test ./...
go run ./cmd/server
```

The default port is 8081 (`PORT` overrides it). Real data is fetched from the KB Securities public daily history endpoint. `MARKET_DATA_PROVIDER=demo` explicitly selects deterministic synthetic data; provider errors never silently fall back to demo. The feed is unofficial, price units are scaled by 1,000, and corporate-action adjustment policy remains unverified.

## API

- `/api/stocks/VCI/wyckoff`: authoritative primary-campaign analysis, plus alternative structures.
- `/api/stocks/VCI/report`: schema-v1 presentation report and exact analyzed OHLCV.
- `/healthz`: health check.

Both analysis routes accept inclusive `from`, `to` and `timeframe=1D|1W`. Weekly candles are aggregated before analysis; the minimum is 40 resulting observations. The last weekly bar may be incomplete.

Existing structure/phase/confidence/range/events/evidence fields remain. New fields expose `analysisWindow`, `priceStructure`, and `context`, including the forming stage, confirmed HH/HL or LH/LL, window reasons, counter-evidence and alternatives. Event `observed_at` is separate from event date. One primary structure supplies chart markers and event evidence; other campaign hypotheses remain under `structures`. Failed outcomes retain history and are marked ×.

The assessment can expand to retain a stopping cluster and its preceding context. An event range can narrow to 12/20 bars only inside an established parent; the assessment itself always retains at least 40 observations. Absence of a required pattern does not authorize fabricated SC/BC or a forced Spring/Test.

## UI

Build the frontend with `npm ci --prefix apps/macro-dashboard` and `npm run build` from the repository root. The service serves `../macro-dashboard/dist`; `DASHBOARD_DIR` overrides it. `/macro` serves the macro dashboard. Vite development uses a stock API proxy.

The research page shows the forming stage, assessment-window size, swing pattern, price/volume and one campaign timeline. HH/HL–LH/LL has a separate chart toggle and only displays pivots confirmed by the cutoff. Macro routes reverse-proxy to `MACRO_API_URL` (default `http://127.0.0.1:3000`); start the PostgreSQL macro service separately.

## Tests and evaluation

```bash
go test ./internal/wyckoff -run TestContextFirst -count=1 -v
go test ./internal/wyckoff -run 'TestReviewedAnnotationBenchmark|TestStructureScopedBenchmark' -count=1 -v
```

Tests exercise mirrored accumulation/distribution campaigns, event availability, phase progression, confirmed swing structure, input validation, reviewed SZC context, and all nine annotation datasets. CI checks formatting, Go tests/vet, frontend tests/build and historical-prefix replay.

Set `WYCKOFF_BENCHMARK_OUTPUT` / `WYCKOFF_SCOPED_OUTPUT` for JSON inventories. CI archives the complete baseline module and uses the same evaluation code and annotations. Scope matching uses fixed geometric/time tolerances, never literal equality between generated and manual IDs. Missing geometry or attribution remains explicit.

The analyzer's current primary-event contract differs from the former all-candidate inventory. Historical benchmark reports describe their own baseline/scope; do not equate CI success or a new recall number with validated investment accuracy. Reviewed labels remain unchanged. Further out-of-sample validation and source-adjustment verification are required before reliability claims.
