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

## NHNN policy rates

`sbv_policy_rates` supplies three EVENT series: VN_SBV_REFINANCING,
VN_SBV_DISCOUNT and VN_SBV_OVERNIGHT, in % per year. The last is the NHNN
standing overnight lending rate, separate from VN_INTERBANK_ON market rates.
Migration 010 registers bindings; all values/raw documents/scheduling use the
existing production tables. No monthly/daily synthetic observations are added.

The Python history worker checks the current NHNN rates table every 24 hours.
It requires all three explicitly labelled rates and unambiguous effective dates.
HTTP/parse failures use the government republication of decision 1123/QD-NHNN
as a bounded historical fallback (4.5%, 3%, 5%, effective 2023-06-19). This fallback
is lower priority, returns PARTIAL, and does not certify the current policy rate.
It is not a complete decision archive. Future changes are read from the current
NHNN table; if that page remains blocked/changes structure, the source stays
unverified and alerts continue. No anti-bot bypass or guessed current values.

Sources:
- https://sbv.gov.vn/vi/lãi-suất1
- https://xaydungchinhsach.chinhphu.vn/nhnn-tiep-tuc-dieu-chinh-lai-suat-dieu-hanh-119230616150218505.htm

Deploy/rebuild API (migration 010), frontend, and Python worker, from repo root:

```sh
docker compose -f docker-compose.macro.yml -f docker-compose.history.yml up -d --build macro-api macro-history
docker compose -f docker-compose.macro.yml -f docker-compose.history.yml run --rm macro-history run --dataset sbv_policy_rates --once --force
```

For a single policy indicator use `--indicator VN_SBV_REFINANCING`; the dataset
fetches all three rates together. Reruns reuse the shared observation writer.
`coverage --dataset sbv_policy_rates` reports decision events without asserting
missing monthly periods. PARTIAL intentionally exits nonzero for operators;
verified historical rows are still saved. If a scheduled run owns the lease,
wait for it rather than trying to steal it.

A decision from years ago can still be in effect. EVENT freshness therefore
uses the last successful current-table check (7 days), not decision age. Failure
retains the last observation but marks SOURCE_ERROR; stale/missing checks are
shown as STALE. The UI labels effective dates and uses a step chart.
