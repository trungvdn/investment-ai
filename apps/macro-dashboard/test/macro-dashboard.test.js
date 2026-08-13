import test from "node:test";
import assert from "node:assert/strict";
import { buildMacroDashboard, calculateTrend, determineRegime } from "../src/domain/macro-dashboard.js";
import { mockIndicators, mockScoreHistory } from "../src/data/mock-macro-data.js";

test("builds a macro context from exactly 15 indicators in five groups", () => {
  const dashboard = buildMacroDashboard(mockIndicators, mockScoreHistory);
  assert.equal(dashboard.indicators.length, 15);
  assert.deepEqual(dashboard.groupScores.map((group) => group.name), ["Liquidity", "Credit", "Inflation", "Growth", "External"]);
  assert.equal(dashboard.macroScore, 62);
  assert.equal(dashboard.regime, "Balanced");
});

test("rejects an incomplete indicator set", () => {
  assert.throws(() => buildMacroDashboard(mockIndicators.slice(1), mockScoreHistory), /exactly 15/);
});

test("derives trends and regimes deterministically", () => {
  assert.equal(calculateTrend([1, 1.1]), "Stable");
  assert.equal(calculateTrend([1, 1.2]), "Rising");
  assert.equal(calculateTrend([1.2, 1]), "Falling");
  assert.equal(determineRegime(40), "Cautious");
  assert.equal(determineRegime(39), "Restrictive");
});
