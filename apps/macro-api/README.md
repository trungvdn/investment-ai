# Live macro data — FRED + NSO

Go API and ingestion worker backed by PostgreSQL. No mock or in-memory production mode. First release enables **8 indicators**; the full 15-item catalog keeps the remaining NHNN/PMI/PPI items explicitly unimplemented.

## Sources and exact definitions

| Indicator code | Primary input | Transformation / meaning |
| --- | --- | --- |
| US_EFFR | FRED `DFF` | Federal funds effective rate, daily %, not Fed target bounds |
| US_10Y | FRED `DGS10` | 10-year constant-maturity Treasury yield, daily % |
| US_BROAD_USD | FRED `DTWEXBGS` | Nominal Broad USD Index (Jan 2006=100), **not ICE DXY** |
| VN_CPI_YOY | NSO economic releases; SDMX `CPIVNM.xml` / `PCPI_IX` for history | Explicit monthly YoY in release; otherwise same-series index ratio over exactly 12 months |
| VN_CORE_CPI_YOY | NSO price releases | Explicit monthly core inflation YoY; never cumulative-average inflation |
| VN_IIP_YOY | NSO `IIPVNM.xml` / `AIP_ISIC4_IX` | Total-industry monthly index ratio over exactly 12 months |
| VN_RETAIL_YOY | NSO economic releases | Monthly nominal retail and consumer-service revenue YoY, not inflation-adjusted growth |
| VN_EXPORT_YOY | NSO economic releases; `METVNM.xml` / `TXG_FOB_USD` for history | Monthly total merchandise exports FOB USD YoY; no interpolation if the previous-year month is missing |

Official sources:

- FRED API: https://fred.stlouisfed.org/docs/api/fred/series_observations.html
- NSO SDMX directory links: https://nsdp.nso.gov.vn/
- Files: https://nsdp.nso.gov.vn/GSO-chung/SDMXFiles/GSO/CPIVNM.xml , `IIPVNM.xml`, `METVNM.xml` at the same directory.
- NSO WordPress API: https://www.nso.gov.vn/wp-json/wp/v2/posts ; searches quoted official economic-report and price-release titles. The exact report URL is retained with each extracted observation.

NSO parsing is deterministic and conservative. Selectors require `REF_AREA=VN`, `COUNTERPART_AREA=_Z`, monthly frequency and the exact aggregate series. Unit multipliers and index-base metadata are checked. Release parsing requires an explicit report period, a monthly sentence and an unambiguous YoY rate; cumulative/quarterly passages are excluded. Conflicting or malformed data is rejected, with raw documents retained. NSO report observations have priority 110 versus SDMX-derived history at 50; both remain queryable as revisions/source variants.

The report feeds backfill the most recent 100 search results per feed. SDMX history is constrained by the actual source coverage. This does **not** promise a gap-free 3-year history for every indicator. The CPI SDMX file was behind newer official releases at implementation time; source precedence avoids making it the sole current source.

## Run with PostgreSQL

With Docker Compose, from repository root:

```sh
# Set a local PostgreSQL password in your shell or a gitignored root .env.
export MACRO_DB_PASSWORD='<your-local-password>'
# Set via your secret manager or shell; never put a key in frontend code.
export FRED_API_KEY='<your-fred-key>'
docker compose -f docker-compose.macro.yml up --build -d
```

PostgreSQL and Macro API bind to localhost, ports 5432 and 3000. The named volume persists across container restarts. If the FRED key is omitted, NSO can still ingest; FRED cards report `NOT_CONFIGURED` rather than synthetic values.

Without Docker, provide PostgreSQL and run from this directory (Go 1.26):

```sh
export DATABASE_URL='postgres://user:password@localhost:5432/macro?sslmode=disable'
export FRED_API_KEY='<your-fred-key>'
go run ./cmd/server
```

Use TLS and a managed secret store for production database connections. Alternatively use standard `PGHOST`, `PGPORT`, `PGUSER`, `PGPASSWORD`, `PGDATABASE`, `PGSSLMODE`. The example `.env.example` is documentation; Go does not automatically load it.

Startup runs the transactional, advisory-lock-protected migration and starts an asynchronous ingestion pass. Subsequent passes run every `INGEST_INTERVAL` (default `6h`, minimum `5m`). This is polling, not a promise of fresh releases every six hours. The frontend reads stored observations and never fetches providers itself. HTTP retries are bounded to three attempts for transport errors/429/5xx; failed sources do not prevent other connectors from running.

An explicit one-off run or backfill (same credentials):

```sh
go run ./cmd/ingest
```

Exit 0: all connectors produced observations; exit 2: some sources failed or lacked configuration; exit 1: run/storage failure. The worker processes each connector independently and preserves the last successful observations. A partial write caused by a database failure can be safely replayed; an entire multi-source run is not one transaction.

Optional backend-only `NSO_ECONOMIC_FEED_URL` / `NSO_PRICES_FEED_URL` overrides must return the same WordPress JSON shape and are intended for controlled source migrations. They are not public API parameters.

## React integration

```sh
npm ci --prefix apps/macro-dashboard
npm run dev
```

Vite routes `/api/macro` to `MACRO_API_URL` (default `http://localhost:3000`), with stock routes still proxied separately. In production, build the React app and run stock-analysis; its `/api/macro/` reverse proxy targets the same `MACRO_API_URL`. No CORS exception or frontend API key is needed. A separately hosted UI must proxy these paths equivalently.

## API (read-only)

- `GET /api/macro/dashboard`: full catalog, latest and canonical history, source-attempt state, `mode=live`, `score=null`, `scoreStatus=NOT_CONFIGURED`.
- `GET /api/macro/indicators`: definitions, unit, comparison basis, frequency and availability policy.
- `GET /api/macro/observations/{code}?from=YYYY-MM-DD&to=YYYY-MM-DD`: inclusive period-end range.
- Add `revisions=true` to return retained revisions and alternative source series, rather than one canonical observation per period.
- `GET /healthz`: process health. A failed database read makes dashboard requests return 503.

No GET request starts ingestion, and no unauthenticated write/ingestion endpoint is exposed.

## Storage and time semantics

`migrations/001_macro.sql` is embedded in the binary and registered in `macro_schema_migrations`. Raw bytes are deduplicated by source, source URL and SHA-256; original and most recent retrieval times are kept. Observations reference the raw document, exact series, unit, transformation, period and priority.

Changed values or source metadata create a new revision under a transaction and per-series/period advisory lock. Repeated identical observations only update the retrieval time. The old revision remains intact. Raw provenance is never discarded when parsing fails after a successful HTTP download.

- `period` / `periodEnd`: the economic period, not fetch time.
- `publishedAt`: only populated when the source supplies a publication time (NSO release timestamp, interpreted as Vietnam time). FRED/SDMX observation publication times remain null.
- `firstSeenAt`: when this system first observed that particular revision, **not** the original historical release date.
- `collectedAt`: most recent successful sighting of the unchanged revision.

Historical API results are the best currently stored observations. This release preserves vintages but does not implement a point-in-time backtest query; backfilled first-seen times cannot reconstruct unknown historical publication dates.

Freshness thresholds are transparent application policies, not official publication SLAs: monthly NSO 60 calendar days from period end, daily FRED rates 7 days, Broad USD 14 days. UI distinguishes missing configuration, ingestion errors, stale periods, latest available data and unimplemented connectors. Missing monthly observations are charted as gaps. Macro Score/regime labels remain disabled until a separately reviewed scoring model exists.

## Verification

```sh
go test ./...
go vet ./...
NSO_LIVE_SMOKE=1 go test ./internal/providers -run TestNSOLiveSmoke -v
MACRO_TEST_DATABASE_URL='postgres://...' go test -race ./internal/store
```

CI runs deterministic parser/service/API tests plus PostgreSQL 17 integration tests for idempotency, revision preservation, concurrent writes, source precedence and survival across repository restarts. Live network smoke tests are opt-in and do not make CI depend on upstream availability. FRED requests are tested with fixtures until a real key is supplied.
