# One production table for historical and latest data

Backfill and scheduled jobs both write to `macro_observations`. There is no new
history data table and no shadow/activate step in this worker. The existing
`macro_raw_documents` retains original response bundles. Job leases, due times,
retry counters and health use `macro_source_state` and `macro_attempts`.
Existing legacy pilot/CPI tables are not deleted or rewritten.

Go Save and Python use the same PostgreSQL function
`macro_save_observation_v1` (migration 007). It locks
`indicator_code + period + series_id`, reuses unchanged values, and appends a
revision only when value/unit/period-end/semantic method changes. Crawl URL,
publication/evidence changes and ARTICLE_ID/PARSER annotations alone do not
create another value revision. True A -> B -> A corrections retain all revisions.
Raw snapshots remain independently deduplicated by source/URL/checksum.

The worker writes each bounded job atomically. Valid rows from a partially
successful crawl are retained with PARTIAL health and warnings. Conflicting or
invalid batches do not replace good observations. A source failure retains old
data, records failed health, and backs off. Backfill and daemon jobs share the
same source lease; --force bypasses due time but cannot steal a running lease.
Source identifiers matching Go providers share the existing scheduler key.

## Deployment

These jobs now **write production data**. Build the API and worker together so
migration 007 installs the shared writer before collection begins. The DB role
needs SELECT/INSERT/UPDATE on raw documents, observations, source state and
attempts, sequence USAGE, registry SELECT and EXECUTE on the shared function.
Use the application role or a dedicated ingestion role through HISTORY_DB_USER /
HISTORY_DB_PASSWORD. A restricted shadow-only PILOT_DB_USER is not sufficient.
The SQL functions use invoker privileges, not SECURITY DEFINER.

```sh
docker compose -f docker-compose.macro.yml -f docker-compose.history.yml up -d --build macro-api macro-history
docker compose -f docker-compose.macro.yml -f docker-compose.history.yml run --rm macro-history backfill --dataset mbs_pmi_history --from 2020-01 --to 2026-09 --start-page 1 --pages 2
docker compose -f docker-compose.macro.yml -f docker-compose.history.yml run --rm macro-history run --once --dataset te_vietnam_pmi --force
docker compose -f docker-compose.macro.yml -f docker-compose.history.yml run --rm macro-history coverage --indicator VN_PMI --from 2020-01 --to 2026-09
docker compose -f docker-compose.macro.yml -f docker-compose.history.yml run --rm macro-history report --dataset mbs_pmi_history
```

The daemon checks due work each minute using the configured six-hour intervals.
Backfill changes the requested page/date scope; it uses the exact same writer as
latest collection. Results show inserted physical rows, revised (a subset of
inserted), unchanged, rawId and warnings. No activate command is required.
WRITTEN means the bounded job wrote valid observations, not that history is complete.
Coverage reads the same production table that the API/dashboard uses.

## PMI

MBS report discovery, bounded PDF download and explicit month/value extraction
supply available history; Trading Economics supplies a latest-month observation
in the same table under its own source series. Source priority chooses one value
per period while revision history retains source vintages. The MBS public archive
and PDF accessibility remain partial/unverified; 12 consecutive real months have
not been established. Coverage must expose gaps instead of inferring values from
charts or forecasts. Exact PDF/HTML evidence remains in raw documents.

No production deployment or DB cleanup has been performed from development.

## Rates and credit

NHNN credit uses national month-end total YTD values, never regional, mid-month or
YoY values. Archive pages use validated real pagination. An API access denial may
fall back to a public NHNN total table, explicitly marked PARTIAL. O/N imports
only dated VND overnight rows; starred rows use the stated prior reference date.
A complete daily NHNN archive is unverified. Public historical table URLs may be
configured with SBV_INTERBANK_HISTORY_URLS; no guessed date-query is used.

FRED_API_KEY is required for EFFR and US 10Y. History requests start in 1900 and
retain available responses, not just the Go adapter's default five-year window.
Missing "." values stay missing; API credentials are not stored in provenance.
Daily coverage lists unobserved weekdays without pretending all are trading days.

```sh
docker compose -f docker-compose.macro.yml -f docker-compose.history.yml run --rm macro-history backfill --dataset sbv_credit_history --from 2020-01 --to 2026-09 --start-page 1 --pages 5
docker compose -f docker-compose.macro.yml -f docker-compose.history.yml run --rm macro-history backfill --dataset sbv_interbank_history
docker compose -f docker-compose.macro.yml -f docker-compose.history.yml run --rm macro-history backfill --dataset fred_effr_history
docker compose -f docker-compose.macro.yml -f docker-compose.history.yml run --rm macro-history backfill --dataset fred_10y_history
```

## IIP and PPI

IIP parses full SDMX input before deriving monthly YoY using same-series prior-year
levels. PPI uses the quarterly national industrial B–E index with same-quarter
prior-year=100 and subtracts 100 once. No monthly PPI values are manufactured.
Both feed macro_observations directly, with source raw bytes preserved.

```sh
docker compose -f docker-compose.macro.yml -f docker-compose.history.yml run --rm macro-history backfill --dataset nso_iip
docker compose -f docker-compose.macro.yml -f docker-compose.history.yml run --rm macro-history backfill --dataset nso_ppi_history
docker compose -f docker-compose.macro.yml -f docker-compose.history.yml run --rm macro-history coverage --indicator VN_PPI_YOY --from 2020-01 --to 2026-09
```

Merge/deploy the stacked PRs in order #12, #13, #14. All groups use this single
production writer. Legacy pilot diagnostics remain optional and are not required
by this worker.

## Lãi suất huy động VND, kỳ hạn 6–12 tháng

Dataset `sbv_deposit_rates` supplies one `VN_DEPOSIT_RATE` monthly indicator:
NHNN's published range of average VND deposit rates for state-owned and joint
stock commercial banks, new and existing deposits, term 6–12 months. It is not
a regulatory ceiling, standing policy rate, individual bank quote, or midpoint.

The Python worker checks https://sbv.gov.vn/vi/trang-chu daily, discovers links
named “Diễn biến lãi suất của tổ chức tín dụng đối với khách hàng tháng M/YYYY”,
then reads the HTML or downloads the report's PDF attachment and uses the
existing Poppler runtime. Reference month comes from the report heading, not
the publication/download date. The parser restricts VND, bank population,
new/existing deposits and exact 6–12 month tenor. USD, lending rates, shorter
terms, conflicting bounds and future months fail validation.

For example, NHNN's September 2026 report, published 22 October 2026, reports
6.6–8.1%/year for this definition. Discovery supports four reports per bounded
job/page; this is not a complete historical archive. For older reports or when
the index is inaccessible, configure `SBV_DEPOSIT_REPORT_URLS` as a space-separated
list of explicit HTTPS NHNN article/PDF URLs (maximum 20). These still require
successful retrieval and the same validation. Access/parse failures remain
visible through existing retries and source alerts; there is no static-value
fallback. Index failure with explicit reports returns PARTIAL and retains valid
observations without claiming latest discovery succeeded.

Migration 010 adds nullable `upper_value` to existing `macro_observations`, and
registers one dataset/binding. `value` stores the published lower bound,
`upper_value` the upper bound. Scalar indicators omit `upperValue` in API JSON.
A shared range-writer wrapper retains both lower-only and upper-only revisions
and the existing lock/idempotency/raw provenance rules. No new data table.

Indicator Monitor displays “lower–upper %”, and its chart plots both bounds.
No standalone interest-rate section or three standing-policy indicators remain.
Previous-period scalar changes are not calculated for ranges. The prior draft
policy dataset is disabled if present; its stored audit data is preserved.

Deploy from the repository root, after merging:

```sh
docker compose -f docker-compose.macro.yml -f docker-compose.history.yml up -d --build macro-api macro-history
# Wait for Macro API /healthz (migration 010), then trigger collection:
docker compose -f docker-compose.macro.yml -f docker-compose.history.yml run --rm macro-history run --dataset sbv_deposit_rates --once --force
```

Rebuild the frontend for production, or restart `npm run dev`. The range field
requires the API, Python collector and frontend changes to be deployed together.
Reruns do not create duplicate observations. A busy lease must finish before
another manual run can claim it. PARTIAL exits nonzero to keep discovery/source
issues visible, even when valid older reports were saved.
