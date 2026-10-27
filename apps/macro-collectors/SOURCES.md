# Historical sources and current updates

The versioned inventory covers all 18 Go dashboard indicators plus the two distinct
CPI pilot metrics. `macro_collectors/source_catalog.json` records candidate history
and update sources, definition, access requirements and unresolved work. It does
not assert a history range from the existence of a URL. The database coverage
command measures actual loaded periods, source vintages and missing months.

## Current implementation

| Indicator | History | Current updates | State |
|---|---|---|---|
| CPI YoY | NSO SDMX PCPI_IX, same-month prior-year transformation | National GSO/NSO CPI archive and topic | Backfill, preview, gated activation and automatic refresh implemented |
| CPI MoM / average YoY | Successfully parsed national archive articles | Same CPI topic | Separate shadow metrics; not spliced into CPI YoY or added to Go dashboard |
| PMI | Unresolved historical export/series source | Trading Economics public page | Latest-month shadow adapter; no fabricated history |
| IIP YoY | NSO SDMX AIP_ISIC4_IX | Same SDMX | Existing shadow adapter; no canonical activation in this phase |
| Other indicators | Existing Go providers and candidate sources in inventory | Existing Go providers | Coverage audit first; no blanket Python migration |

The national CPI category was inspected on 2026-10-26: it lists 27 pages,
including reference months back to October 2004. This is **discovery evidence**,
not a claim that every historical article parses or that the archive is complete.
Observed database coverage is authoritative. The initial SDMX replay previously
reached October 2025; refetch and measure it rather than assuming that cutoff.

Sources inspected:
- https://www.nso.gov.vn/cpi-vi/
- https://www.nso.gov.vn/cpi-vi/?paged=2
- https://www.nso.gov.vn/cpi-vi/?paged=27
- https://nsdp.nso.gov.vn/GSO-chung/SDMXFiles/GSO/CPIVNM.xml
- https://tradingeconomics.com/vietnam/manufacturing-pmi

Trading Economics may present a JavaScript/bot-verification page. That is not an
empty PMI value and is never bypassed or replaced with a forecast; inspect the
recorded source error. A historical PMI source remains unresolved. FRED's existing
Go adapter defaults to a rolling five-year request, not all history. M2's October
2025 methodology break must not be bridged. ICE DXY is not FRED Broad USD.

## Build, inspect and collect

From the repository root, retain any compose overrides used by your deployment:

```sh
docker compose -f docker-compose.macro.yml -f docker-compose.pilot.yml up -d --build macro-api macro-pilot

docker compose -f docker-compose.macro.yml -f docker-compose.pilot.yml run --rm --entrypoint python macro-pilot -m macro_collectors.sources inventory

docker compose -f docker-compose.macro.yml -f docker-compose.pilot.yml run --rm --entrypoint python macro-pilot -m macro_collectors.sources coverage --indicator VN_CPI_YOY --from 2025-01 --to 2026-09
```

Migration 006 is applied by macro-api. The default policy maps
`nso_cpi_yoy` to history and `gso_cpi_yoy` to updates, with tolerance 0.02
percentage points and at least three consecutive overlapping months.
The policy remains inactive on installation and on repeated migrations.

Backfill defaults to **SDMX only**, independently of report access or enabled archive datasets.
It downloads and normalizes the complete available SDMX history using prior-year
levels for YoY; it does not filter the upstream download to the requested dates.
Without dates, load all available SDMX history:

```sh
docker compose -f docker-compose.macro.yml -f docker-compose.pilot.yml run --rm --entrypoint python macro-pilot -m macro_collectors.sources backfill --indicator VN_CPI_YOY --source sdmx
```

`HISTORY_LOADED` means all currently stored SDMX history was loaded and its internal
range has no missing months. It is not a claim that the source covers all years
or the latest published month. To check a desired range, supply dates; these
control the completeness report:

```sh
docker compose -f docker-compose.macro.yml -f docker-compose.pilot.yml run --rm --entrypoint python macro-pilot -m macro_collectors.sources backfill --indicator VN_CPI_YOY --source sdmx --from 2021-01 --to 2026-09
```

The example dates are a requested historical range, not a claim that SDMX covers it.
The result reports `availableFrom`, `availableTo`, `storedPeriods`,
`coveredRequestedPeriods` and `missingRequestedMonths`. `BACKFILL_INCOMPLETE`
returns a nonzero exit code when requested months are unavailable. An index month
without its same-month prior-year level cannot yield YoY and is not fabricated.
A stale SDMX source can still supply valid historical data; it does not certify
that the latest month is current.

To additionally scan national report archives for bridge/gap periods, explicitly
select `--source combined`:

```sh
docker compose -f docker-compose.macro.yml -f docker-compose.pilot.yml run --rm --entrypoint python macro-pilot -m macro_collectors.sources backfill --indicator VN_CPI_YOY --source combined --from 2025-01 --to 2026-09 --start-page 1 --pages 3
```

Combined mode fetches SDMX first and commits each archive page independently.

Use the coverage report to choose the needed range and further archive pages.
A job accepts 1–10 pages, within pages 1–100, and up to 12 article URLs per page.
It does not guess article slugs or recursively crawl arbitrary hosts.
Each page has its own HTTP budget. Rerun a page range safely after interruption.
`NOT_DUE_OR_BUSY` means an existing lease is active; it is never stolen.

Good articles on a partially failing page are preserved, with warnings and status
`PARTIAL`; the command exits nonzero and never calls that page complete.
Conflicting values for the same month/basis reject the page with raw evidence.
Unsupported older layouts, missing comparators and absent average values are
reported as missing, never inferred. Repeated source/period content does not add
source revisions; changed content creates an immutable new revision. Collection
runs remain separate audit events, and identical raw bytes deduplicate.

Refresh the current source after archive work, then inspect coverage and preview:

```sh
docker compose -f docker-compose.macro.yml -f docker-compose.pilot.yml run --rm macro-pilot run --once --indicator VN_CPI_YOY --force

docker compose -f docker-compose.macro.yml -f docker-compose.pilot.yml run --rm --entrypoint python macro-pilot -m macro_collectors.sources preview --indicator VN_CPI_YOY
```

## Selection and dashboard activation

Preview does not write production data. It reports every overlapping pair,
difference, source choice, missing month, latest/expected month and blocking issue.
For the implemented national CPI YoY policy:

- Both bindings must be enabled and refer to the same indicator/definition.
- Every overlapping month must agree within 0.02 percentage points.
- At least three overlap months must be consecutive.
- The combined range must contain no internal monthly gaps and reach at least
  the previous calendar month; future/nonfinite/wrong-unit values are rejected.
- Both sources must have been checked today or yesterday, with no latest
  `SOURCE_ERROR` or `PARTIAL` result.
- SDMX history is preferred for every available historical month, including
  overlaps. Reports supply newer periods or missing historical months. Values are never averaged or interpolated.

After inspecting an eligible preview, the explicit activation command atomically
publishes selected values with original raw provenance and pins the chosen
production observation ID for each period:

```sh
docker compose -f docker-compose.macro.yml -f docker-compose.pilot.yml run --rm --entrypoint python macro-pilot -m macro_collectors.sources activate --indicator VN_CPI_YOY
```

Activation reruns the gate in a repeatable-read transaction; it cannot activate a
stale preview. If blocked, it exits nonzero and preserves the previous active
batch. No other indicator can be activated in this phase. This command requires
the additional DB write privileges below; a read-only/shadow role cannot publish.

The existing Go API/dashboard reads the selected CPI batch. All source revisions
remain queryable through revision history. Existing Go collectors continue running;
this is a transfer of **canonical selection**, not a transfer of every provider's
ingestion ownership. Later Go source writes cannot override a pinned selection.
After activation, normal CPI pilot runs refresh the active batch automatically
only when the same gate still passes. A source failure or conflicting revision
retains the previous batch and records failed canonical health instead of replacing
good values. Activation and automatic refresh are idempotent and audited.

Return to Go's priority-based selection:

```sh
docker compose -f docker-compose.macro.yml -f docker-compose.pilot.yml run --rm --entrypoint python macro-pilot -m macro_collectors.sources deactivate --indicator VN_CPI_YOY
```

Deactivation is atomic and audited. It retains all published observations; Go's
priority rules may still select those valid published values. It does not erase
data or restore a byte-for-byte pre-activation database.

## Database roles and validation

The shadow role additionally needs SELECT/INSERT/UPDATE on `macro_source_values`
and USAGE on its sequence, plus SELECT on source policies/batches/links.
For activation/automatic refresh, use a separately provisioned role that can
SELECT/INSERT/UPDATE production raw documents, observations and attempts,
SELECT/INSERT/UPDATE canonical batches and links, UPDATE source policies, and
INSERT selection events, with sequence USAGE. Credentials are never printed.
A role lacking production writes can continue inventory, coverage and preview.

CI exercises source-vintage deduplication/revisions, archive parsing and partial
warnings, overlap/gaps/freshness, blocked activation, production provenance,
idempotent activation, retained prior batches, deactivation, and Go API selection
and health. No live database/backfill or activation was performed in development.
Runtime source access and actual coverage must be checked in deployment.

## Rerun identity and duplicate diagnostics

A source revision is identified by binding + period and a change to
value, unit, period end or method. URL, publication metadata, evidence text and
paragraph positions are retained in raw/run audit snapshots; those crawl metadata
changes alone no longer create a new source or production value revision.
Existing pre-fix hashes are handled by comparing the stored semantic fields,
so the first rerun after upgrade does not insert another metadata-only revision.
The first value's original evidence/raw provenance stays intact.

An actual A -> B -> A change still creates three revisions. Different source
series for the same indicator/month are also retained. The default API history
and active canonical selection choose one observation per period; revision history
and per-run pilot audit snapshots intentionally contain multiple rows.

Use the read-only queries in `sql/check_duplicates.sql` to distinguish consecutive
identical source/production vintages from expected collection-run snapshots.
The patch prevents new metadata-only duplicates; it does not delete existing
source revisions, raw evidence, or immutable canonical batches. No production DB
inspection or destructive cleanup has been performed from development.

Canonical activation also retries the whole transaction (up to three attempts) on
serialization/deadlock/unique-revision races. A Go writer can otherwise commit a
revision after Python's repeatable-read snapshot but before its value lock. Each
retry rolls back the incomplete batch and reads a fresh snapshot; it does not
skip the gate, steal leases, or retry one INSERT against a stale snapshot.
