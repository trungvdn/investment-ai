# Live macro data — FRED + NSO + NHNN + S&P Global + ICE

Go API and ingestion worker backed by PostgreSQL. No mock or in-memory production mode. The catalog now contains **18 indicators**, with **18 definitions enabled**. DXY requires an authorized export feed; enabled does not mean data is available. M2 YoY may intentionally have no observation until a comparable prior-year month exists.

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

Startup runs the transactional, advisory-lock-protected migration and starts an asynchronous scheduler. It checks due sources every minute; per-source cadence and retry state are stored in PostgreSQL (see Operations below). `INGEST_INTERVAL` is the fallback cadence (default `6h`, range `5m` to `168h`). This is polling, not a promise of fresh releases every six hours. The frontend reads stored observations and never fetches providers itself. HTTP retries are bounded to three attempts for transport errors/429/5xx; failed sources do not prevent other connectors from running.

An explicit scheduler tick (same credentials; respects persisted due dates and active leases):

```sh
go run ./cmd/ingest
```

Exit 0: connectors produced observations or explicitly reported an expected methodology/publication gap; exit 2: some sources failed or lacked configuration; exit 1: run/storage failure. The worker processes each connector independently and preserves the last successful observations. A partial write caused by a database failure can be safely replayed; an entire multi-source run is not one transaction.

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


## NHNN + quarterly PPI extension

Additional official inputs:

| Code | Input and meaning |
| --- | --- |
| VN_USD_VND | NHNN central USD/VND rate from https://sbv.gov.vn/vi/tỷ-giá ; exact central-rate table, not the commercial reference buy/sell or tax table. Effective date is the economic date, and the issue date is recorded separately (date precision only). |
| VN_INTERBANK_ON | NHNN weighted-average VND overnight interbank rate from https://sbv.gov.vn/vi/lãi-suất1 ; exact O/N row and the effective date of its scoped interbank table. Policy rates and other tenors are excluded. |
| VN_M2_LEVEL | Published total means of payment, billion VND, methodology from October 2025 onward. |
| VN_M2_YTD | Published change from previous year-end, not YoY. Empty/non-comparable YTD observations are left absent, never zero. |
| VN_M2_GROWTH | YoY calculated only from exact corresponding months under the post-October-2025 methodology. Latest period without a comparable denominator is `NOT_COMPARABLE`. |
| VN_CREDIT_GROWTH | Published **total** economy-wide credit growth versus previous year-end, not a sector growth rate and not YoY. |
| VN_PPI_YOY | NSDP `PPIVNM.xml`, exact industrial aggregate `VNM_PPPI_ISIC4_BTE_BY_PP_IX`, quarterly. This series is already same quarter of previous year = 100, so percentage change is **index minus 100**, not another four-quarter ratio. |

M2 and credit use the public NHNN page's own structured endpoint `https://sbv.gov.vn/o/article/v1.0/articles`, scope 20117, structure IDs 10038520 (M2) and 10034332 (credit), status=0. Public reference pages:

- https://sbv.gov.vn/vi/tổng-phương-tiện-thanh-toán-và-tiền-gửi-của-khách-hàng-tại-tctd
- https://sbv.gov.vn/vi/du-no-tin-dung-doi-voi-nen-kt-dttktt

Pagination is bounded, every downloaded page is retained, and each observation refers to its actual originating page. Parsing requires a Vietnamese report title and matching explicit **month-end** field. Legacy non-month-end/inconsistent records and empty templates are skipped with warnings, rather than treating upload dates as economic periods. Revisions are chosen by modification timestamp and then article version; conflicting same-vintage totals fail validation. Raw article IDs remain in transformation metadata.

NHNN explicitly reports a methodology change beginning October 2025. Consequently, this connector does not stitch earlier M2 levels into the new series or calculate YoY across that boundary. The dashboard adds separate level and YTD entries, explaining the expansion from 15 to 17 catalog items. Known historical periods without meaningful YTD are left as gaps. `NOT_COMPARABLE` is an expected data-policy state and does not by itself make the ingestion CLI fail. A real upstream error still does.

The exact method boundary is reviewed against the NHNN source note, not inferred from a change in numeric values. Unrecognized methodology-change notes are rejected for manual review. Later revisions or methodology changes require rechecking these assumptions before altering the boundary.

Quarterly PPI retains `YYYY-Qn` and the real quarter-end date in storage/API. The UI leaves missing quarters as gaps and does not fabricate monthly values. Percentage-series changes are in percentage points; YTD comparisons across the year boundary are suppressed to avoid presenting the reset as an economic decline.

No SQL schema change is needed: the existing observation period string and period-end date support the new codes and quarters. Existing vintages remain intact. Refresh uses the per-source operational cadences below. Historical M2/credit availability is constrained by verified source records; FX and O/N currently ingest the visible official snapshot and accumulate daily history over time. No historical daily backfill is fabricated.

Backend-only optional overrides `SBV_FX_URL`, `SBV_INTERBANK_URL` and `SBV_ARTICLES_URL` support controlled source migrations and tests. They do not bypass authentication or protection mechanisms. Source HTTP failures remain visible; body-read failures now receive the same bounded retries as transport failures.

Official cross-check for industrial PPI semantics: https://www.nso.gov.vn/tin-tuc-thong-ke/2026/10/thong-cao-bao-chi-ve-tinh-hinh-gia-thang-chin-quy-iii-va-9-thang-nam-2026/ reports industrial PPI Q3/2026 at +4.75% YoY, matching the SDMX same-quarter index 104.75115176098 after subtracting 100 (rounding difference only).

Opt-in network verification:

```sh
SBV_PPI_LIVE_SMOKE=1 go test ./internal/providers -run TestSBVPPILiveSmoke -v
```

Deterministic tests cover central FX versus buy/sell rates, effective versus issue dates, O/N versus policy/other-tenor rates, aggregate versus sector credit, M2 definition breaks/blank YTD/revisions, multi-page raw provenance, quarterly PPI transformations and truncated HTTP responses. PostgreSQL CI also round-trips quarterly observations across a connection restart.

## Phase 5 — PMI, DXY and NHNN bills

**PMI:** original S&P Global Vietnam Manufacturing PMI headline, seasonally adjusted, index points, monthly; 50 means unchanged business conditions, not 50% growth. By default, discovers up to six explicitly titled Vietnam manufacturing release links from https://www.pmi.spglobal.com/Public/Release/PressReleases?language=en . `PMI_RELEASE_URLS` overrides discovery for a controlled source migration or backfill (comma separated; maximum 36, exactly one vintage per month). This is a public-release reader, not access to the licensed historical database.

Example release URL verified in the S&P public index for September 2026:

```sh
export PMI_RELEASE_URLS='https://www.pmi.spglobal.com/Public/Home/PressRelease/dba8ca68ce13403591b68cb3e07175d9'
```

When using an explicit override, update release URLs monthly. Automatic discovery rejects unrecognized links/layouts and emits a source-quality alert; it never guesses release identifiers. Failed download/parsing retains previous observations. No numbers are embedded in production code. The parser reads page one of PDF releases with `pdftotext` (install `poppler-utils`; included in the Docker image), or text from HTML releases. It requires the Vietnam manufacturing headline, explicit report month/year and publication embargo time; it rejects other countries, subindices, duplicate months, ambiguous values and future releases. Each release has its own exact raw-document link. Publication time is the stated embargo, distinct from month-end and collection time. A revision must replace the URL for that month, rather than providing two vintages in one ingestion batch.

**NHNN bills:** https://sbv.gov.vn/vi/nghiệp-vụ-thị-trường-mở — the official OMO result table, scoped to `ls01-wrap`, exact result date and explicit **Bán tín phiếu NHNN** sections. `VN_SBV_TBILL` is `sum(awarded volume × awarded annual rate) / sum(awarded volume)`, across bill tenors in that result. Units are %/year and billion VND. Tenors and awarded total are recorded in the method metadata. The tenor basket may change, so the UI suppresses a misleading day-to-day yield change. This series is neither Treasury bills, repo rates, outstanding stock nor net liquidity injection. History accumulates from observed result snapshots; no fabricated backfill.

On a valid result table without bill awards or a published bill rate, the connector returns `NO_PUBLISHED_RATE`, stores the raw result, does not fabricate a zero yield and does not count the source as a failed ingestion. Previous awards retain their own dates and are not current quotes. An award with positive volume but missing/invalid rate is a parsing failure; HTTP 403/429, missing result blocks or changed table schemas are source failures, not evidence of no auction. `SBV_BILLS_URL` optionally overrides the result-page location.

Live access is environment-dependent: SBV has returned HTTP 403 on the deployment host in an earlier run, and S&P returned HTTP 403 during phase-5 validation. No access-control bypass or substitute market series is used. Verify from the deployment host after configuring URLs:

```sh
PHASE5_LIVE_SMOKE=1 go test ./internal/providers -run TestPhase5LiveSmoke -v -count=1
```

Deterministic tests include actual PDF extraction, headline vs prior/subindex selection, December rollover, publication embargo, URL boundaries, volume-weighted bill rates, repo/Treasury exclusion, zero awards, malformed tables and persistence of expected missing-rate status. CI installs the same PDF reader used in production.

## Operations — collection, retries and quality

The server polls due sources once a minute. Each provider has a persisted `next_due`, lease/token, last successful check, consecutive failure count and safe error message in migration 002. A PostgreSQL atomic claim prevents simultaneous runs across API replicas or the one-off worker. Per-source fetch/ingestion has a 2-minute deadline; claims and completion queries have 10-second deadlines. A crashed owner can be reclaimed after its 10-minute lease; old owner tokens cannot complete newer runs. Shutdown cancels active requests. A worker restarting preserves due times and cooldowns. These are interval schedules measured from completion, not wall-clock cron jobs; slow sources may delay later checks in the sequential worker.

Default cadences:

| Providers | Cadence |
| --- | --- |
| Central FX, ICE DXY | 1 hour |
| Interbank, NHNN bill results, S&P PMI index | 2 hours |
| NHNN M2 / credit statistics | 12 hours |
| FRED / NSO and other providers | `INGEST_INTERVAL`, default 6 hours |

All polling runs throughout the week; missing weekend/holiday releases do not generate synthetic observations. Override individual IDs with `INGEST_SCHEDULE` JSON, for example `{"sbv:fx":"1h","spglobal:pmi:vn":"2h"}`. Durations outside 5m..168h and unknown IDs fail startup. Timestamps are stored in UTC and shown in Vietnam time on the operations panel. Changes to cadence apply after the next due collection; they do not reset an existing persisted cooldown. The one-off `cmd/ingest` runs the same due-source scheduler and respects leases. After changing credentials/URLs or resolving a network issue, an operator can request an immediate check using `go run ./cmd/ingest -source ice:dxy -due` (substitute the exact configured provider ID). This adjusts the due time but refuses to override an active lease. There is no unauthenticated HTTP reset/ingest action.

HTTP attempts are limited to three with exponential delay plus jitter for transport/body failures, 429 and 5xx. Other 4xx (including 403) are not retried inline. `Retry-After` seconds/date is honored; delays over 60 seconds are persisted as a future retry rather than blocking the worker. Across scheduler ticks transient failures use 5m, 15m, 45m… capped at 6h; access denial/data-quality failures cool down 6h; missing configuration cools down 24h. A longer upstream Retry-After takes precedence. Error classes distinguish access denial, rate limiting, DNS, TLS, timeout, read, transport and data quality. Request URLs, authorization headers and API keys are omitted from errors. The last good stored observations remain available.

`GET /api/macro/operations` is read-only and exposes source schedules and current alerts. Warnings are also emitted as structured JSON logs with provider, class, consecutive failures and next due time; recovery resets the counter and logs a recovery event. Alerts include source outages, missing/stale indicator data and schedules overdue by 15 minutes with no active lease. A successful check is not proof of a fresh observation: stale economic periods remain stale. Normal M2 comparability gaps and absence of published bill yields are not outages. No email/Slack delivery has been configured. `/healthz` remains process liveness; operations/dashboard return 503 on storage failures.

Quality gates validate economic periods, future publication dates, unit/definition agreement, finite values, positive currency/index/monetary levels, PMI bounds and growth rates not below -100%. Duplicate same-indicator/period/series observations reject the entire provider batch before normalized writes. Exact raw responses are retained when parsing/validation fails; valid previous observations and vintages are not replaced by failed batches. Deterministic tests cover these cases, retries/cancellation/secret-safe failures and recovery. PostgreSQL CI additionally verifies concurrent claims, expiry/fencing, migration idempotence and schedule persistence after reopening connections.

### Actual DXY versus Broad USD

`US_DXY` is a separate daily ICE spot index (March 1973=100); `US_BROAD_USD` remains FRED DTWEXBGS. DX futures are not accepted as spot DXY. Configure an authorized export endpoint via `DXY_FEED_URL` (HTTPS) and optional `DXY_FEED_TOKEN` (Bearer). The endpoint contract is UTF-8 CSV with exact columns:

```csv
date,value,symbol
2026-10-23,101.2,DXY
```

The row above is a schema example, not seeded market data. The feed must export the actual daily spot-index observation for the stated date. Duplicate dates, other symbols, nonfinite/nonpositive values and future dates fail quality checks. The backend retains the raw export and publishes ICE's index-definition source URL, never the credential-bearing feed URL. An unconfigured feed is `NOT_CONFIGURED`; there is no fabricated substitute. Live DXY ingestion cannot be verified until an authorized endpoint is provided. Official definition: https://www.ice.com/fixed-income-data-services/index-solutions/currency-indices .

After deployment: verify `/api/macro/operations`, inspect source errors, supply the DXY feed if needed, and check a successful next-due run from the deployment host. A 403 remains a source-access failure; scheduling/retry handling does not claim to remove SBV/S&P access restrictions.

## Resolving PMI HTTP 403

A 403 is an access denial before parsing; it is not caused by the headline parser or DXY configuration. Both the public S&P release index and a specific release URL returned 403 during investigation. Those responses do not establish whether the blocking layer is S&P or a network intermediary. Merely setting `PMI_RELEASE_URLS` does not fix access when the document URL is denied too. The connector now identifies `release index` versus `release document` in errors and preserves `ACCESS_DENIED` classification and cooldowns. It does not retry 403 inline or send DXY credentials to S&P.

Three independent PMI transport modes, in precedence order:

1. **Authorized export:** `PMI_FEED_URL` plus optional `PMI_FEED_TOKEN`. This mode does not request the public index or PDF URLs. The token is sent only to the configured export endpoint, and credential-bearing endpoint URLs are not published in provenance or errors. `PMI_FEED_TOKEN` without `PMI_FEED_URL` is a configuration error. DXY variables have no effect on this mode.
2. **Official local PDFs:** `PMI_RELEASE_FILES` is a JSON array of absolute local paths and original official release URLs. The worker uses the existing PDF/headline/date checks, retains the exact PDF as raw provenance, and makes no S&P HTTP request. Files must be regular, readable PDFs of at most 16 MiB, one vintage per report month (maximum 36). The files must be obtained from a permitted source; no sample observations are seeded.
3. **Public HTTP:** existing explicit release URLs or automatic index discovery. This continues to report actionable errors if public access is denied. There is no silent third-party scrape, index substitution, or guessed release ID.

Authorized export contract (illustrative schema, not a production data seed):

```json
{
  "seriesId": "SPGLOBAL_VN_MANUFACTURING_PMI",
  "country": "VN",
  "frequency": "MONTHLY",
  "observations": [{
    "period": "2026-09",
    "value": 51.9,
    "publishedAt": "2026-10-01T00:30:00Z",
    "sourceUrl": "https://www.pmi.spglobal.com/Public/Home/PressRelease/dba8ca68ce13403591b68cb3e07175d9"
  }]
}
```

The export must contain the actual seasonally adjusted Vietnam manufacturing headline, original report month, original publication/embargo timestamp and official release provenance. Other countries/series/frequencies, duplicate months, missing values, values outside 0–100, future observations, contradictory publication dates, trailing JSON and unknown fields are rejected. The feed itself must use HTTPS; authentication and the conversion to this export schema belong to the configured data provider/integration. This is not a generic parser for arbitrary vendors' API responses. Last good observations remain unchanged when a feed fails or returns invalid data.

Local-PDF deployment example:

```sh
export PMI_PDF_DIR=/absolute/host/path/pmi
export PMI_RELEASE_FILES='[{"path":"/data/pmi/release.pdf","sourceUrl":"https://www.pmi.spglobal.com/Public/Home/PressRelease/dba8ca68ce13403591b68cb3e07175d9"}]'
docker compose -f docker-compose.macro.yml -f docker-compose.pmi-local.example.yml up -d --build macro-api
docker compose -f docker-compose.macro.yml -f docker-compose.pmi-local.example.yml exec macro-api ingest -source spglobal:pmi:vn -due
```

Use the original URL corresponding to the PDF you actually downloaded, not an unrelated sample URL. The container's `macro` user must be able to read the directory/files; the example mounts the directory read-only. When adding a new monthly release or revision, update the file/URL mapping and recreate the container. For an export, set `PMI_FEED_URL` and `PMI_FEED_TOKEN` on the ordinary compose service instead. Then requeue with `ingest -source spglobal:pmi:vn -due`: restarting alone preserves the existing six-hour access-denial cooldown. Check `/api/macro/operations` and the indicator's source/period after the run. Requeue refuses to steal an active lease.

Regression tests reproduce 403 at both public stages, verify classification/no retry/no credential leakage, and ingest the same headline through an authenticated export or a locally supplied PDF without contacting the denied public host. Live export ingestion still requires a configured endpoint; live local ingestion requires the actual release file. This change provides those working ingestion paths; it does not claim to unblock the public website.
