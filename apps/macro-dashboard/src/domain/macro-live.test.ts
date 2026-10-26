import { it, expect } from "vitest";
import {
  historyPoints,
  previousChange,
  sourceMessage,
  statusLabels,
} from "./macro-live";
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

it("does not compare bill yields across changing tenor baskets, or expose status prefixes", () => {
  const i = {
    code: "VN_SBV_TBILL",
    history: [{ value: 3.5 }, { value: 4 }],
    status: "NO_PUBLISHED_RATE",
    lastAttempt: {
      message: "NO_PUBLISHED_RATE: Không có lãi suất trong phiên.",
    },
  } as MacroIndicator;
  expect(previousChange(i)).toBeNull();
  expect(sourceMessage(i)).toBe("Không có lãi suất trong phiên.");
  expect(statusLabels.NO_PUBLISHED_RATE).toContain("không công bố");
});

it("shows a normal publication delay without exposing the manifest status prefix", () => {
  const i = {
    status: "AWAITING_RELEASE",
    lastAttempt: { message: "AWAITING_RELEASE: Chờ PMI kỳ mới" },
  } as MacroIndicator;
  expect(sourceMessage(i)).toBe("Chờ PMI kỳ mới");
  expect(statusLabels.AWAITING_RELEASE).toBe("Chờ công bố kỳ mới");
});
