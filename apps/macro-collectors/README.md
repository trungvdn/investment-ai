# Scheduled macro collectors

Python 3.12 service that discovers and downloads public-source artifacts. It does not replace the Go API/storage layer. The first plugin collects official Vietnam manufacturing PMI PDFs; Go remains the authoritative headline/value/date parser and saves observations plus exact raw provenance.

## Automatic PMI collection

- A fresh installation attempts a bootstrap collection immediately for the previous calendar month.
- After success, the next scheduled run is **07:30 on day 1 of the following month, Asia/Ho_Chi_Minh**. The daemon checks due work each minute.
- Day 1 may be a weekend/holiday or precede the release. If the expected month is not present, retry every 6 hours through day 10; do not reuse an older report as the new month. After day 10, emit `RELEASE_OVERDUE` and check daily.
- Transport errors use persisted 15m exponential backoff capped at 6h. HTTP 401/403 is not retried inline and cools down 6h. Retry-After can postpone a run further (bounded to one year for malformed/extreme hints). The HTTP client checks a 180-second budget before attempts, with up to three attempts per request, a 30-second socket timeout and a 16 MiB response bound. PDF text extraction is bounded to page one and 15 seconds per artifact.
- After success, no automatic revision checks run until next month; use the manual force command to check a revised release sooner.

This official-PDF workflow **does not solve a publisher/network HTTP 403 by changing languages**. It uses the same official public S&P index and release URLs with a transparent user agent and standard certificate verification. There is no proxy rotation, challenge solver, credential reuse or third-party data substitution. If the deployment network cannot access those URLs, collection reports `ACCESS_DENIED`; use the authorized feed/local-PDF modes or resolve access with the source/network operator. Discovery and scheduling are covered by deterministic tests; live S&P access remains denied in the current validation environment.

## Deploy alongside the API

```sh
docker compose -f docker-compose.macro.yml -f docker-compose.collectors.yml up -d --build
docker compose -f docker-compose.macro.yml -f docker-compose.collectors.yml logs -f macro-collectors
```

The override starts the Python service and mounts the same named data volume read/write for the collector and read-only for Go. It clears PMI feed credentials and sets `PMI_RELEASE_MANIFEST=/collector-data/vn_pmi/manifest.json` on the Go service. This selects the collector bridge; it does not issue a second public S&P request from Go. Allow several minutes for the first Go ingestion check; a missing initial manifest is retried as a transient startup condition. The collector and API must see the same absolute artifact paths. A short init container owns the dedicated volume directory as UID 1000, matching both non-root runtime images; this avoids a root-owned fresh-volume startup failure.

One-off run (respect schedule):

```sh
docker compose -f docker-compose.macro.yml -f docker-compose.collectors.yml run --rm macro-collectors --once
```

Explicit retry after fixing connectivity, then prompt Go to import:

```sh
docker compose -f docker-compose.macro.yml -f docker-compose.collectors.yml run --rm macro-collectors --once --force
docker compose -f docker-compose.macro.yml -f docker-compose.collectors.yml exec macro-api ingest -source spglobal:pmi:vn -due
```

`--force` ignores the due time, **never** the per-collector process lock. Exit code 1 means a collection error; pending publication, not-due and busy are non-error outcomes. The last successful PDF mapping remains in the manifest on failures, but its error status prevents the Go bridge from falsely marking the source healthy. Files and manifests are atomically replaced and fsynced; PDFs use immutable SHA-256 filenames. State and next-due dates survive container restarts. POSIX file locks release after a crash. Run workers on the same local volume, or shard by collector name; file locks are not a distributed cross-host scheduler.

## Integration and monitoring

The manifest has version, collector ID, target month, checked/next-due timestamps, status (`READY`, `PENDING`, `ERROR`), safe error class/message, failure count and `releases` (absolute PDF path + original S&P URL). Go imports a READY manifest and validates the full PDF, including matching its actual headline month to the manifest target. PENDING becomes `AWAITING_RELEASE`, not a zero value or source outage. ERROR propagates to the existing source-health panel/logs. A manifest not updated within 24 hours after its scheduled run is reported as an overdue worker. The exact manifest and imported PDF remain linked to observations; old stored values survive collection failures.

Logs are JSON records and omit response bodies, credentials and private local paths. Check both `macro-collectors` logs and `/api/macro/operations`: fetching a PDF successfully is not proof that Go accepted its values. The scheduled release month and actual observation period must agree. No Slack/email destination is configured.

## Add another indicator

Implement a source plugin with:

1. `target(now)` returning the economic period to collect.
2. `schedule` implementing `next_success(now)`, `next_pending(now)` and `pending_overdue(now)`. `MonthlySchedule` and `IntervalSchedule` cover monthly and daily/hourly plugins.
3. `collect(directory, target, now)` returning validated artifact/export metadata. Raise `NotPublished` for a normal release delay or `CollectionError` for a safe, classified source error.

Register its factory in `__main__.py`, give it a unique collector ID and add source-specific tests. Each collector has independent locks/state/next-due times, so they can be assigned to separate containers. Add the corresponding Go import adapter/catalog definition when introducing a new artifact schema; the PMI bridge intentionally does not accept other indices. The shared HTTP client, atomic publication, scheduler, state, logging and container remain reusable.

Tests (no network or third-party Python packages):

```sh
cd apps/macro-collectors
python -m unittest discover -s tests -v
```

Tests cover month/year rollover, Sunday release delays, restart persistence, independent interval plugins, immutable artifacts, source/URL validation, no retry on 403, Retry-After deferral, preservation of previous releases and force-run locking. Go tests separately exercise actual PDF extraction and the manifest bridge's raw provenance and health propagation.

## Shadow ingestion pilot

The separate `macro_collectors.pilot` command runs CPI and IIP SDMX datasets directly against PostgreSQL in **shadow mode only**. It never writes `macro_observations`, `macro_attempts` or `macro_raw_documents`, and never changes ingestion ownership. See `PILOT.md` for configuration, comparisons and review gates.

## Secondary PMI HTML source

The direct-DB shadow pilot also supports the user-selected Trading Economics Vietnam PMI page, independently of the official PDF workflow. See [PILOT.md](PILOT.md#trading-economics-vietnam-pmi-html) for migration, scheduling, manual commands and validation limits.
