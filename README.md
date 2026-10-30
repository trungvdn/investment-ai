# Investment AI

## Vision

Build an autonomous investment intelligence platform that continuously collects,
understands, reasons about, and explains financial markets to help investors
make better investment decisions.

## Mission
Empower investors with explainable investment intelligence.

## Product Goals
See `docs/vision/product-goals.md`.

## Design Principles
- Business-first
- Capability-driven
- Workflow-first
- AI as an enabler
- Explainability by design

## High-Level Architecture
(To be completed in Day 3)

## Business Capabilities
See `ROADMAP.md` (to be completed).

## Project Roadmap
See `ROADMAP.md`.

## Documentation Index
- docs/vision/
- docs/workflows/
- ADRs/

## Run local services

With Go 1.26, Node.js, npm dependencies installed for `apps/macro-dashboard`,
and PostgreSQL configured for the Macro API, use the root Makefile:

```sh
make run-macro       # Macro API + macro dashboard
make run-all         # Macro API + macro dashboard + stock-analysis API
make run-macro-docker # PostgreSQL + Macro API using Docker Compose
```

The individual services can also be started with `make run-macro-api`,
`make run-macro-dashboard`, or `make run-stock-analysis`. The Macro API
requires PostgreSQL and its usual environment configuration; see
[`apps/macro-api/README.md`](apps/macro-api/README.md). The dashboard runs
through Vite and proxies requests to the local APIs.

`make run-macro-docker` starts the PostgreSQL and Macro API services defined
in `docker-compose.macro.yml` in the foreground. Set `MACRO_DB_PASSWORD`
before running it; `FRED_API_KEY` is optional. Stop the stack with Ctrl+C.
The dashboard and stock-analysis API are not included in the current Compose
file and can be started separately with their individual Make targets.

On Windows, PowerShell does not include `make` by default. Install GNU Make to
use the targets above, or start each service from the repository root in a
separate PowerShell terminal:

```powershell
go -C .\apps\macro-api run .\cmd\server
npm --prefix .\apps\macro-dashboard run dev
go -C .\apps\stock-analysis run .\cmd\server
```

The Macro API still requires a configured PostgreSQL connection; the
PowerShell commands do not start PostgreSQL.
