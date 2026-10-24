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

Run `go run ./cmd/server` separately from `apps/stock-analysis`. Vite proxies `/api` to `http://localhost:8081`; override with the `STOCK_API_URL` environment variable. Real stock data is the backend default. For offline inspection only, explicitly run `MARKET_DATA_PROVIDER=demo go run ./cmd/server`; report banners identify synthetic data. Provider failures never trigger a silent demo fallback.

## Production

```sh
npm ci --prefix apps/macro-dashboard
npm run build
cd apps/stock-analysis
go run ./cmd/server
```

Open `http://localhost:8081`. Go serves the built `../macro-dashboard/dist` directory and the stock API together. Override the directory with `DASHBOARD_DIR` (relative to the server process working directory). Build before starting the production server; source TSX files cannot be served directly.

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
