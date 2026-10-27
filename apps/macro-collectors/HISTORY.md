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
