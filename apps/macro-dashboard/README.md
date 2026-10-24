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
- Macro: all 15 existing mock indicators and five equal-weight group scores preserved. Filter/search/sort the table, select group cards, choose 3/6-month score history, or inspect a specific indicator's six-month history.
- Macro data remains the explicitly labeled **August 2026 mock dataset**. This migration does not add a live macro feed. Stock summary translations supplement the preserved original backend text; no company metadata is invented.
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
