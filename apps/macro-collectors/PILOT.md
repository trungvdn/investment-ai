# Current source-selection phase

Default collection remains shadow-only. Migration 006 adds source vintages, historical backfill and an explicit, gated CPI dashboard selection path. See [SOURCES.md](SOURCES.md) for commands, privileges and rollback. The earlier SDMX parity gate below does not automatically activate this new policy.

# NSO Python ingestion pilot

Original SDMX parity scope: `nso_cpi` (VN_CPI_YOY / PCPI_IX) and `nso_iip` (VN_IIP_YOY / AIP_ISIC4_IX). Both use structured official NSO SDMX and the same-series twelve-month YoY transformation already implemented in Go. PMI/SBV are not migration candidates in this pilot because their public access is currently unreliable. No LLM is needed for these XML sources.

## Registry and ownership

Migration 003 introduces sources, datasets, indicator definitions and many-to-many indicator bindings. A source has many datasets; a dataset can output many indicators; an indicator can have bindings to multiple datasets. Bindings identify a versioned extractor/transform, source series, unit, priority and validation rules. Dataset cadence is independent of indicator freshness. CPI and IIP are seeded as enabled, six-hour datasets **owned by Go**. Existing configuration is not overwritten by rerunning migrations.

The registry is initially a two-indicator pilot, not a replacement for the entire Go catalog. Python accepts the tested SDMX, national CPI article and Trading Economics PMI adapters. Editing a row does not enable arbitrary code, URLs or a new parser. Add a tested adapter before expanding the pilot.

## Deploy without changing production values

```sh
docker compose -f docker-compose.macro.yml -f docker-compose.pilot.yml up -d --build
docker compose -f docker-compose.macro.yml -f docker-compose.pilot.yml logs -f macro-pilot
```

Go applies migration 003. The pilot retries if the API has not yet completed migrations. This is a separate service from the optional PMI PDF collector; it has no dependency cycle and can run independently. Both existing Go NSO connectors stay active. Configure `PILOT_DB_USER` / `PILOT_DB_PASSWORD` for a restricted role, or the compose default reuses the application's DB role for an initial controlled deployment.

Manual whole pilot, one indicator or a single dataset (same locks, no duplicated active work):

```sh
docker compose -f docker-compose.macro.yml -f docker-compose.pilot.yml run --rm macro-pilot run --once --force
docker compose -f docker-compose.macro.yml -f docker-compose.pilot.yml run --rm macro-pilot run --once --dataset nso_cpi --force
docker compose -f docker-compose.macro.yml -f docker-compose.pilot.yml run --rm macro-pilot run --once --indicator VN_CPI_YOY --force
docker compose -f docker-compose.macro.yml -f docker-compose.pilot.yml run --rm macro-pilot report
```

`--force` bypasses due time but cannot steal a running lease. Atomic PostgreSQL claims fence stale owners and allow recovery after a ten-minute lease. Jobs run sequentially within one process; additional processes coordinate through the same pilot state table. Source failures retain old production data and prior comparison runs. The pilot never silently switches a dataset away from Go.

## What gets compared

Python stores exact downloaded bytes/checksum in `macro_pilot_raw`, normalized candidates in `macro_pilot_observations`, and a JSON comparison report in `macro_pilot_runs`. The state/schedule is separate from production `macro_source_state`. Raw bytes deduplicate by dataset/checksum; repeated runs retain independent audit history, and repeated same-day runs do not count as extra matching days.

Comparisons use the **exact source series**, not the canonical dashboard value: CPI's report-based source may have higher priority or different rounding and must not be mistaken for a parser disagreement. Before comparing:

- The Go provider must have a successful completed attempt after its latest raw fetch, with no active collection lease.
- The Python payload checksum must equal the latest Go raw snapshot. A new source revision produces `VINTAGE_MISMATCH`, not a false parser failure; Go is allowed to catch up and Python retries in five minutes. An operator can requeue the original provider, for example `ingest -source nso:sdmx:PCPI_IX -due`, before rerunning the pilot; this still does not transfer ownership.
- Within a repeatable-read transaction, compare every available normalized period, value (absolute tolerance 1e-8), unit, period end, method, source URL and priority. Missing periods on either side and metadata mismatches are explicitly listed.
- Freshness is evaluated separately: report latest periods, the expected previous calendar month, missing release periods, a ten-day publication window and the existing sixty-day maximum age. Equal but old values do not produce a fresh-data MATCH.

Statuses include NO_BASELINE, BASELINE_BUSY, BASELINE_INCOMPLETE, VINTAGE_MISMATCH, SOURCE_ERROR, MISMATCH, AWAITING_RELEASE, RELEASE_LAG, STALE and MATCH. Nothing is eligible simply because the HTTP request succeeded. Normal publication lag is not fabricated as zero. Original Go observations and revisions remain untouched.

## Review gate, not automatic cutover

`report` emits `eligibleForReview` only if:

- The latest run is MATCH with at least twelve comparable observations and was checked within the last 24 hours.
- At least three distinct **Vietnam calendar days** within fourteen days have MATCH runs for the same parser version and binding configuration hash, each with at least twelve matched observations.
- The latest run has no value/metadata/missing-period discrepancy or freshness lag.

Eligibility is only an input to review. The original shadow/parity path itself exposes no ownership-promotion command. The separate CPI canonical-selection path now has a gated production publisher described in SOURCES.md. Do not change `ingestion_owner` manually. Dataset ingestion ownership remains Go. CPI canonical selection uses explicit observation links, shared production revision locks, an activation gate, audited deactivation and separate health; it does not disable already-running Go providers. Broader ingestion ownership transfer is still a later phase. This prevents an untested switch from silently replacing the existing ingestion path.

## Restricted DB role

Provision credentials separately. A pilot role needs SELECT on `macro_sources`, `macro_datasets`, `macro_indicator_definitions`, `macro_indicator_bindings`, `macro_observations`, `macro_raw_documents`, `macro_attempts`, `macro_source_state`; SELECT/INSERT on the four `macro_pilot_*` tables, UPDATE on `macro_pilot_state`, and UPDATE(checksum) on `macro_pilot_raw`; and sequence USAGE for the pilot raw/run ID sequences. It does not need production-table INSERT/UPDATE/DELETE or schema creation. Use that role for the strongest enforcement of shadow-only writes. No credentials are printed by the CLI.

## Validation limits

Fixtures prove parser parity, discrepancy detection, freshness classification, missing data and state/ownership guards. CI PostgreSQL tests exercise the actual shared schema, shadow-only persistence, checksums, repeat runs and claims. A deployed Go baseline and successful upstream fetches are required to produce a real comparison report. No production DB, credentials or baseline data are available in the development environment; eligibility is never inferred from fixtures or hardcoded demonstration values.

## Initial live replay check — 2026-10-26

The development check fetched the current official NSO XML files and replayed the **same bytes** through the existing Go parser (`cmd/sdmx-replay`) and the new Python parser. CPI matched 51 periods with no value/metadata discrepancies, but its newest SDMX period was October 2025; IIP matched 105 periods with no discrepancies, newest September 2026. CPI is therefore a source-freshness failure even though the parsers agree, and must not qualify for cutover. A higher-priority report-based CPI source may be newer; this pilot does not confuse it with the SDMX series.

This was a source/parser replay, **not a production database comparison**. Real DB comparisons and three-day review eligibility require deploying the shadow service against the existing Go baseline. The replay does not contribute matching days or seed production values.

## National CPI report collector

Migration 004 adds `gso_cpi_reports`, which scrapes the national **Chỉ số giá tiêu dùng** topic at `https://www.gso.gov.vn/cpi-vi/`. When the old host is unavailable it tries the same publisher's current `https://www.nso.gov.vn/cpi-vi/`. It discovers dated CPI/price reports from the page's actual links; it does not guess article slugs or retrieve provincial CPI sites. It processes the newest linked report month and checks duplicate publications for contradictory values.

Three separate bindings are persisted to shadow observations:

| Indicator | Meaning | Unit |
| --- | --- | --- |
| VN_CPI_MOM | Current month versus previous month | % MoM |
| VN_CPI_YOY | Current month versus the same month last year | % YoY |
| VN_CPI_AVG_YOY | Average January..report-month versus the same period last year | % average YoY |

The extractor does not confuse December-baseline growth, quarterly averages, core CPI, gold/USD indices or component-price changes with those figures. Vietnamese month names, decimal commas, increases/decreases and unchanged values are handled explicitly. Missing monthly MoM/YoY or conflicting values reject the article rather than substituting an average for monthly YoY. If the source omits the YTD average, valid monthly values are retained and the average is reported as missing, never filled with a quarter average or zero. Source HTML (including rejected article evidence), article URLs, publication dates/precision, matched text, paragraph index and average-window length remain auditable. No LLM is required for this stable national report format.

Run or inspect this dataset:

```sh
docker compose -f docker-compose.macro.yml -f docker-compose.pilot.yml run --rm macro-pilot run --once --dataset gso_cpi_reports --force
docker compose -f docker-compose.macro.yml -f docker-compose.pilot.yml run --rm macro-pilot report --dataset gso_cpi_reports
```

If you already use other Compose overrides, retain those `-f` arguments when adding the pilot override. Migration 004 is applied by the rebuilt Go API; restart/rebuild it before starting this collector. Existing SDMX datasets remain as historical/comparison sources. `--indicator VN_CPI_YOY` resolves all enabled mapped pilot datasets, so the new report source and legacy SDMX source can both be inspected without collapsing their provenance.

This new article dataset uses **cross-source review**, not the SDMX same-payload parity gate: it compares monthly YoY with Go's currently selected value for the same indicator/period, reports previously missing periods and marks MoM/average metrics as new metrics without an existing baseline. Numerical agreement between different publications is explicitly not parser parity. `CROSS_SOURCE_REVIEW`/`CROSS_SOURCE_MISMATCH` never contributes to automatic review eligibility; no production values or ingestion ownership are changed. The two additional metric definitions are in the pilot registry, not newly enabled production dashboard cards.

The public topic page and linked September 2026 reports were inspected during implementation. Direct network fetches from this execution environment were blocked by environment policy, so live scraping must still be verified on the deployment host. Deterministic tests cover the national summary structure, all comparison bases, precision/evidence, source aliases and DB shadow persistence.

## Trading Economics Vietnam PMI (HTML)

Dataset `te_vietnam_pmi` reads the user-selected public page
<https://tradingeconomics.com/vietnam/manufacturing-pmi>. This is HTML, not a PDF;
no PDF download, LLM, API key or S&P request is needed for this adapter.
Trading Economics is the secondary publisher; S&P Global remains the original
index producer. `VN_PMI` is an index level, not a percentage change.

Migration 005 registers the dataset/binding. Rebuild/start macro-api first to apply
migrations, then rebuild macro-pilot. The existing daemon checks this dataset every
six hours, including release day and subsequent days; manual runs use the same lease.

```sh
docker compose -f docker-compose.macro.yml -f docker-compose.pilot.yml up -d --build macro-api macro-pilot
docker compose -f docker-compose.macro.yml -f docker-compose.pilot.yml run --rm macro-pilot run --once --indicator VN_PMI --force
docker compose -f docker-compose.macro.yml -f docker-compose.pilot.yml run --rm macro-pilot report --dataset te_vietnam_pmi
```

The parser uses the actual current/previous summary heading with explicit month/year,
checks month adjacency and range 0–100, and ignores forecasts, scripts and historical
news. Missing/conflicting headings or layout changes fail with `DATA_QUALITY` and
retain raw HTML. A source 403 remains `ACCESS_DENIED`; there is no bypass.
Publication time stays null when not explicitly provided; fetch time is separate.
Reports show the observed reference month and freshness, not the collection month.
Only the latest published observation is extracted; this is not a history backfill.

Each run keeps audit observations, while identical HTML deduplicates in raw storage.
Compare values against the same production indicator/period with tolerance 0.01.
No baseline is `NEW_SOURCE_PERIOD`; a numerical match is `CROSS_SOURCE_VALUE_MATCH`,
never same-payload parser parity. All writes remain in shadow tables, no dashboard
update or cutover is performed. Run this dataset on the deployed network to verify
actual HTTP accessibility: the local runtime's network policy blocked direct requests,
although the public page was inspected through web retrieval on 2026-10-26.
