# Historical indicator collection

## PMI / MBS

Migration 007 registers `mbs_pmi_history`. The new history worker writes only raw,
per-run audit and normalized source-vintage tables. It does not activate dashboard
values or alter Go provider ownership. Trading Economics remains the separate latest
PMI collector.

MBS index pages and download links are discovered from actual HTML, not guessed
monthly slugs. Each job processes at most eight reports on one listing page,
20 MiB of documents and the first 20 pages of each PDF. Exact PDF bytes are archived
as base64 plus SHA256 in the raw bundle. Text extraction uses installed Poppler;
there is no LLM call. Forecasts and unlabelled chart numbers are excluded.
A year omitted next to an explicitly stated month is resolved against the report
issue year with December/January rollover and recorded in evidence.

The inspected listing only exposed a partial archive. This adapter is **not**
proof of a complete 2020–present series. Some PDFs could not be fetched with the
research browser. Runtime access, all older layouts and 12 consecutive real months
still require a deployment coverage check. Unsupported files become warnings;
conflicting values reject a job. Original producer S&P Global and secondary MBS
publisher are recorded separately.

```sh
docker compose -f docker-compose.macro.yml -f docker-compose.history.yml up -d --build macro-api macro-history
docker compose -f docker-compose.macro.yml -f docker-compose.history.yml run --rm macro-history backfill --dataset mbs_pmi_history --from 2020-01 --to 2026-09 --start-page 1 --pages 2
docker compose -f docker-compose.macro.yml -f docker-compose.history.yml run --rm macro-history coverage --indicator VN_PMI --from 2020-01 --to 2026-09
docker compose -f docker-compose.macro.yml -f docker-compose.history.yml run --rm macro-history report --dataset mbs_pmi_history
```

Dates are requested coverage, not an assertion of source availability. `PARTIAL`
and `SOURCE_ERROR` exit nonzero for manual runs. Successful old observations can
be `STALE` for current-update purposes and still valid historical records.
`NOT_DUE_OR_BUSY` never steals an active source lease. Reruns reuse source vintages;
run snapshots remain independent audit events. DB permissions are the same extended
shadow privileges documented in SOURCES.md. No production write permission is needed.

The daemon checks due work every minute using the registry's six-hour interval
and page one; manual archive pages use the same DB leases. Keep current/update
collectors alongside this service. No historical source is silently promoted.

## Lãi suất và tăng trưởng tín dụng

Migration 008 adds `sbv_credit_history`, `sbv_interbank_history`,
`fred_effr_history` and `fred_10y_history`. They remain shadow-only.

Credit requests one actual NHNN article page at a time, validates pagination and
national month-end total fields, and keeps the latest explicit source version.
Growth is % YTD, never YoY; regional/sector-only and mid-month figures are rejected.
On API ACCESS_DENIED only, a separate public NHNN HTML table is tried. That fallback
is explicitly PARTIAL and does not claim the entire archive is accessible.

O/N reads only explicit dated NHNN VND overnight tables. Starred rows use the
stated prior reference date. Actual dated links or public URLs configured in
`SBV_INTERBANK_HISTORY_URLS` can add historical tables; no guessed date-query
parameters are used. Full daily history is **not yet verified**. Coverage reports
unobserved weekdays and does not pretend holidays are missing trading sessions.
No weekly average is substituted for a daily O/N observation.

FRED needs `FRED_API_KEY` in the history service. Requests start in 1900 and keep
all returned available history (not the Go adapter's default five-year window).
Missing "." observations remain missing; truncated responses fail. API keys never
enter raw provenance. The --from/--to range filters normalized imports by month.
Disable unused optional datasets in the registry if credentials are not configured.

```sh
docker compose -f docker-compose.macro.yml -f docker-compose.history.yml run --rm macro-history backfill --dataset sbv_credit_history --from 2020-01 --to 2026-09 --start-page 1 --pages 5
docker compose -f docker-compose.macro.yml -f docker-compose.history.yml run --rm macro-history backfill --dataset sbv_interbank_history --from 2020-01 --to 2026-09
docker compose -f docker-compose.macro.yml -f docker-compose.history.yml run --rm macro-history backfill --dataset fred_effr_history
docker compose -f docker-compose.macro.yml -f docker-compose.history.yml run --rm macro-history backfill --dataset fred_10y_history
docker compose -f docker-compose.macro.yml -f docker-compose.history.yml run --rm macro-history coverage --indicator VN_INTERBANK_ON --from 2020-01 --to 2026-09
```

Inspect lastPage/totalCount to choose further credit pages. All original reports
are retained. A successful bounded job is not a declaration of complete history.
No source is activated for dashboard publication in this phase.

## IIP and quarterly industrial PPI

Migration 009 registers `nso_ppi_history`; `nso_iip` reuses its existing registry
and shared lease, including when the pilot daemon is already running. IIP parses
the complete XML before deriving monthly YoY from the same series/prior-year month.
Missing denominators are not fabricated. PPI selects only national industry B–E,
quarterly frequency, zero unit multiplier and the same-quarter-prior-year=100
basis; the published index is reduced by 100 exactly once.

Both SDMX imports retain the entire available transformed history. --from/--to
provides a requested coverage window, never removes required IIP denominator data.
Coverage understands quarterly periods; it does not manufacture monthly PPI.
The replay CLI supports `nso_ppi_history` so CI executes the actual Go parser on
the same captured XML and checks Python values/periods. This verifies parsing,
not live source availability or completeness.

```sh
docker compose -f docker-compose.macro.yml -f docker-compose.history.yml run --rm macro-history backfill --dataset nso_iip
docker compose -f docker-compose.macro.yml -f docker-compose.history.yml run --rm macro-history backfill --dataset nso_ppi_history
docker compose -f docker-compose.macro.yml -f docker-compose.history.yml run --rm macro-history coverage --indicator VN_IIP_YOY --from 2020-01 --to 2026-09
docker compose -f docker-compose.macro.yml -f docker-compose.history.yml run --rm macro-history coverage --indicator VN_PPI_YOY --from 2020-01 --to 2026-09
```

Merge/deploy these stacked changes in order: PMI worker, rates/credit, IIP/PPI.
All groups use semantic source-vintage deduplication, independent raw/run audit,
persistent scheduling/leases and partial-source diagnostics. The development
execution environment is unavailable; GitHub Actions runs Python/PostgreSQL/Go
tests and Docker smoke checks. No production backfill or source activation was
performed. PMI 12-month real coverage and a complete NHNN daily O/N archive remain
source-verification work, explicitly visible through missing-period reports.
