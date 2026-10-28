# Manual gap fills

`2026-10-27.sql` inserts user-supplied CPI monthly YoY (not monthly MoM or
average CPI) and credit growth YTD (not YoY) into `macro_observations`, using
`macro_save_observation_v1` and a raw JSON audit record in `macro_raw_documents`.
No source retrieval or new table is required. Values are stored as percentage
points: 4.65 means 4.65%, not 0.0465.

From the repository root, after deploying this change:

```sh
git pull --ff-only origin main
docker compose -f docker-compose.macro.yml up -d --build macro-api
docker compose -f docker-compose.macro.yml exec -T postgres \
  psql -U macro -d macro -v ON_ERROR_STOP=1 \
  < apps/macro-api/data/manual/2026-10-27.sql
```

The database must already have migrations through 009 (applied by Macro API on
startup). Wait for `/healthz` to return 200 before running the script. Restart
Vite to use the updated source URL validation, or rebuild the production React
bundle. No collector/backfill execution is needed.

The transaction skips existing visible periods, reports inserted/skipped rows,
and is safe to rerun. It does not overwrite existing values. The manual source
is explicitly labelled `Manual (user supplied; unverified)` and links to the SQL
in this repository, not an unverified official release. Publication time is
unknown; collection time records the actual insertion time.

Manual entries have priority -1000000 so official sources take precedence.
They remain eligible while a CPI canonical batch is active, allowing missing
months to appear without modifying that batch's provenance or selection report.
When a canonical batch includes an official value for the same month, that value
wins. Existing observations and source revisions remain intact.

Inspect all source rows (the API selects its visible source per period):

```sh
docker compose -f docker-compose.macro.yml exec -T postgres \
  psql -U macro -d macro -c "SELECT indicator_code,period,value,unit,source,series_id FROM macro_observations WHERE (indicator_code='VN_CPI_YOY' AND period IN ('2025-12','2026-01','2026-02','2026-03')) OR (indicator_code='VN_CREDIT_GROWTH' AND period='2026-09') ORDER BY indicator_code,period,priority DESC;"
```

This is an explicit operator script, not a startup migration. It requires access
to the target PostgreSQL instance and does not run merely by merging the PR.
