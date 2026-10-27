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
