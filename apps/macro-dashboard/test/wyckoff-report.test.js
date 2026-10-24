import test from "node:test";
import assert from "node:assert/strict";
import {
  normalizeSymbol,
  dateRange,
  validateReport,
  escapeHTML,
  scenarios,
} from "../src/domain/wyckoff-report.js";
import { fetchStockReport } from "../src/data/stock-api-client.js";
const report = {
  schemaVersion: "1.0",
  symbol: "FPT",
  structure: "UNCLEAR",
  phase: "UNKNOWN",
  confidence: 0.4,
  events: [],
  evidence: [],
  dataQuality: { mode: "demo" },
  timeframe: "1D",
  candles: [
    { date: "2026-01-01", open: 10, high: 12, low: 9, close: 11, volume: 100 },
  ],
};
test("normalizes symbols and rejects injected or empty values", () => {
  assert.equal(normalizeSymbol(" fpt "), "FPT");
  for (const v of ["", "<script>", "FPT/../VCB", "A"])
    assert.throws(() => normalizeSymbol(v));
});
test("calendar month range clamps month-end instead of overflowing", () =>
  assert.deepEqual(dateRange(6, new Date("2026-08-31T12:00:00Z")), {
    from: "2026-02-28",
    to: "2026-08-31",
  }));
test("requires same-symbol report, correct timeframe and valid candles", () => {
  assert.equal(validateReport(report, "FPT", "1D"), report);
  assert.throws(() => validateReport(report, "VCB", "1D"));
  assert.throws(() => validateReport(report, "FPT", "1W"));
  assert.throws(() =>
    validateReport(
      { ...report, candles: [{ ...report.candles[0], high: 1 }] },
      "FPT",
      "1D",
    ),
  );
});
test("fetch preserves requested window and reports source failure without demo fallback", async () => {
  let url;
  const result = await fetchStockReport({
    symbol: " fpt ",
    timeframe: "1D",
    from: "2026-01-01",
    to: "2026-10-01",
    fetchImpl: async (u) => {
      url = u;
      return new Response(JSON.stringify(report), {
        headers: { "content-type": "application/json" },
      });
    },
  });
  assert.equal(result.symbol, "FPT");
  assert.match(url, /timeframe=1D&from=2026-01-01&to=2026-10-01/);
  await assert.rejects(
    fetchStockReport({
      symbol: "FPT",
      fetchImpl: async () => new Response("unavailable", { status: 502 }),
    }),
    /không khả dụng/,
  );
});
test("static-host fallback cannot masquerade as a report", async () =>
  assert.rejects(
    fetchStockReport({
      symbol: "FPT",
      fetchImpl: async () =>
        new Response("<html>", { headers: { "content-type": "text/html" } }),
    }),
    /Không tìm thấy API/,
  ));
test("scenarios require actual range and labels are escaped", () => {
  assert.deepEqual(scenarios(null), []);
  assert.equal(scenarios({ support: 10, resistance: 12 }).length, 3);
  assert.equal(
    escapeHTML('<img onerror="bad">'),
    "&lt;img onerror=&quot;bad&quot;&gt;",
  );

});
