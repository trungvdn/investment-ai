import { it, expect } from "vitest";
import { historyPoints, previousChange } from "./macro-live";
import type { MacroIndicator } from "../lib/types";
it("leaves a gap in monthly charts and does not call nonconsecutive observations a monthly change", () => {
  const indicator = {
    frequency: "MONTHLY",
    history: [
      { period: "2026-01", periodEnd: "2026-01-31", value: 3 },
      { period: "2026-03", periodEnd: "2026-03-31", value: 4 },
    ],
  } as MacroIndicator;
  expect(historyPoints(indicator, 12)).toEqual([
    { period: "2026-01", value: 3 },
    { period: "2026-02", value: null },
    { period: "2026-03", value: 4 },
  ]);
  expect(previousChange(indicator)).toBeNull();
});

it("preserves quarterly gaps and does not resample PPI into months", () => {
  const i = {
    frequency: "QUARTERLY",
    comparisonBasis: "YoY",
    unit: "% YoY",
    history: [
      { period: "2025-Q4", periodEnd: "2025-12-31", value: 3 },
      { period: "2026-Q2", periodEnd: "2026-06-30", value: 4 },
    ],
  } as MacroIndicator;
  expect(historyPoints(i, 12)).toEqual([
    { period: "2025-Q4", value: 3 },
    { period: "2026-Q1", value: null },
    { period: "2026-Q2", value: 4 },
  ]);
  expect(previousChange(i)).toBeNull();
});
it("does not treat the year-end YTD reset as a fall in growth", () => {
  const i = {
    frequency: "MONTHLY",
    comparisonBasis: "YTD",
    history: [
      { period: "2025-12", value: 15 },
      { period: "2026-01", value: 1 },
    ],
  } as MacroIndicator;
  expect(previousChange(i)).toBeNull();
});
