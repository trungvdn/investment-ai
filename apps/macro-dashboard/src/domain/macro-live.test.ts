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
