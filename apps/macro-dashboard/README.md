# Investment AI — React Research Workspace

Both Wyckoff Research (`/`) and the macro dashboard (`/macro`) are React pages in one application. The legacy `/macro.html` address redirects through React Router. No iframe or imperative HTML report renderer remains.

## Stack

React + TypeScript + Vite, Tailwind CSS, shadcn-style locally owned Button/Tabs built on Radix and CVA (`components.json`), TanStack Query, React Router, TradingView Lightweight Charts and Recharts. Package versions are locked in `package-lock.json`.

## Development

From the repository root (Node 24 and Go 1.26):

```sh
npm ci --prefix apps/macro-dashboard
npm run dev
```

Run `go run ./cmd/server` separately from `apps/stock-analysis`. Vite proxies `/api` to `http://127.0.0.1:8081`; override with the `STOCK_API_URL` environment variable. Real stock data is the backend default. For offline inspection only, explicitly run `MARKET_DATA_PROVIDER=demo go run ./cmd/server`; report banners identify synthetic data. Provider failures never trigger a silent demo fallback.

## Production

```sh
npm ci --prefix apps/macro-dashboard
npm run build
cd apps/stock-analysis
go run ./cmd/server
```

Open `http://127.0.0.1:8081`. Go serves the built `../macro-dashboard/dist` directory and the stock API together. Override the directory with `DASHBOARD_DIR` (relative to the server process working directory). Build before starting the production server; source TSX files cannot be served directly.

For separate frontend hosting, serve `dist`, proxy `/api/stocks/` to Go, and rewrite `/macro` to `index.html`. API and missing asset requests must return their real statuses rather than the SPA entry point.

## Experience

- Shared collapsible navy sidebar, mobile menu and persistent light/dark preference. Lazy-loaded routes share one React shell.
- Wyckoff: normalized search, daily/weekly windows, query cancellation/timeout, explicit unknown/insufficient/provider-error states and URL-based analysis settings.
- Interactive OHLCV chart: zoom, pan, crosshair, event markers, support/resistance and a shaded range. Toggle events/range/volume. Click an event date on the chart or choose one in the timeline to inspect evidence and focus the relevant interval.
- Keyboard-accessible Radix report tabs preserve evidence, risks, watch conditions and quality metadata. Confidence remains rule strength, never a calibrated probability.
- Macro: live observations from the Macro API, source/period/status filters and per-indicator history. Coverage cards replace market scores.
- Macro source definitions and partial coverage are documented in the Macro API README. Stock summary translations supplement the preserved original backend text; no company metadata is invented.
- TradingView attribution remains enabled on its chart.

## Checks

```sh
npm test
npm run build
cd apps/stock-analysis
go test ./...
go vet ./...
```

Root tests run the existing Node domain/API tests followed by React component tests. CI also type-checks and builds the production bundle.

## Live macro data (FRED + NSO)

The macro page now reads `/api/macro/dashboard` through TanStack Query, without mock fallback. It shows per-series periods, source links, collection/publication times, status and revision provenance; coverage replaces Macro Score. The initial eight series have been extended with NHNN/PPI; the current catalog and availability policies come from the API. Broad USD is explicitly different from ICE DXY. Missing monthly values remain gaps in the chart.

Run PostgreSQL + macro-api with `docker compose -f docker-compose.macro.yml up --build -d` after setting `MACRO_DB_PASSWORD` and optionally `FRED_API_KEY`. Vite proxies `/api/macro` to `MACRO_API_URL` (default :3000); stock API requests still go to :8081. See `apps/macro-api/README.md` for exact source series, refresh policies and operational commands. Production stock-analysis also forwards `/api/macro/` to the Macro API.


The NHNN/PPI extension adds central FX, VND overnight interbank rates, aggregate credit YTD, separate M2 level/YTD/YoY, and quarterly industrial PPI. The catalog expands to 17 entries (15 enabled), with PMI/NHNN bills added in phase 5 (17 enabled definitions; source configuration and availability still apply). The page shows three source families (FRED, NSO, NHNN), explicit M2 comparability status, true quarterly gaps and YTD year-reset handling. See the Macro API README for the October 2025 M2 methodology boundary and exact primary-source definitions.

Phase 5 adds the PMI 50-point reference line and S&P Global source links. NHNN bill yields are weighted by awarded volumes; changing tenor baskets are not displayed as a comparable daily yield change. A valid OMO result without a published bill yield displays `NO_PUBLISHED_RATE`, preserving any previous observation with its own date. PMI public release URLs are configured on the backend, not in the browser. See the Macro API README for source setup and live-access limitations.

The operations extension adds a persisted per-source schedule/status panel (Vietnam time), safe source-error details, consecutive failure counts and overdue-worker warnings from `/api/macro/operations`. The catalog contains 19 definitions: actual ICE spot DXY is separate from FRED Broad USD and requires a configured authorized export. PMI now discovers explicit Vietnam manufacturing links from S&P's public release index; explicit release URLs remain an override. See the Macro API operations section for source cadences, retries and access limitations.

## Troubleshoot Macro API while running npm run dev

The browser calls the Vite origin at /api/macro/dashboard. Vite forwards that
request to MACRO_API_URL (default http://127.0.0.1:3000), matching the IPv4 Docker
port binding. npm run dev does not start PostgreSQL or Macro API.

From the repository root:

```sh
docker compose -f docker-compose.macro.yml up -d --build postgres macro-api
curl -i http://127.0.0.1:3000/healthz
curl -i http://127.0.0.1:3000/api/macro/dashboard
MACRO_API_URL=http://127.0.0.1:3000 npm run dev
```

Set MACRO_DB_PASSWORD in the environment/root .env before Compose. If the API
does not answer, inspect `docker compose -f docker-compose.macro.yml ps` and
`docker compose -f docker-compose.macro.yml logs --tail=80 macro-api`.
Through Vite, test http://127.0.0.1:5173/api/macro/dashboard (use Vite's actual port).
If direct API works but Vite does not, check/restart the Vite proxy configuration.

MACRO_API_URL and STOCK_API_URL may also be set in
apps/macro-dashboard/.env.local; shell settings take precedence. Restart Vite
after changing these server-only values. Do not expose API credentials via VITE_
variables. In a container, use the backend's reachable service name instead of
that container's own loopback address.

A fetch/network error differs from HTTP 502 (proxy cannot reach API), HTTP 503
(storage unavailable), and a provenance-validation error. New PMI observations
from MBS and Trading Economics are recognized for VN_PMI; primary GSO/NSO aliases
are supported. HTTPS and exact-host validation remain enabled, without mock
fallback. Validation errors identify the offending indicator.

The deposit-rate indicator (VN_DEPOSIT_RATE) displays NHNN's monthly VND 6–12 month range in Indicator Monitor. Both bounds are plotted and retained; a midpoint is not used. See ../macro-collectors/HISTORY.md for source discovery, PDF extraction and ingestion commands.
