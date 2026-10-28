import { describe, it, expect, vi, afterEach } from "vitest";
import { fetchMacroScore, validateMacroScore } from "./macro-score-client";
export function scoreFixture() {
  const codes = ["VN_DEPOSIT_RATE", "VN_PMI", "VN_CPI_YOY", "VN_CREDIT_GROWTH", "VN_IIP_YOY", "US_EFFR", "VN_PPI_YOY", "VN_M2_GROWTH"];
  const weights = [20, 20, 15, 15, 10, 10, 5, 5];
  const hash = "a".repeat(64);
  const rules = codes.map((code, i) => ({ code, weight: weights[i], critical: i < 3, mode: "test" }));
  const current = {
    asOf: "2026-10-28T12:00:00Z", modelVersion: "macro-v1.0.0", modelHash: hash, inputHash: hash,
    score: 50, status: "READY", coverage: 100,
    components: rules.map(r => ({ ...r, name: r.code, group: "Liquidity", score: 50, contribution: r.weight / 2, effectiveWeight: r.weight, reason: "ELIGIBLE", inputs: [{ indicatorCode: r.code, value: 3, period: "2026-09", source: "test", firstSeenAt: "2026-10-28T11:00:00Z", revision: 1, rawId: 1 }], warnings: [] })),
    groups: ["Liquidity", "Production", "Inflation"].map(name => ({ name, weight: 30, coverage: 100, score: 50 })),
  };
  return { model: { version: current.modelVersion, hash, minCoverage: 80, rules }, frequency: "daily", current, history: [] };
}
afterEach(() => vi.unstubAllGlobals());
describe("macro score response", () => {
  it("accepts the versioned model and rejects inconsistent score/coverage", () => {
    const d = scoreFixture(); expect(validateMacroScore(d).current?.score).toBe(50);
    d.current.coverage = 80; expect(() => validateMacroScore(d)).toThrow();
    d.current.coverage = 100; d.current.score = 90; expect(() => validateMacroScore(d)).toThrow();
  });
  it("rejects duplicate components, wrong versions and fabricated READY", () => {
    const d = scoreFixture(); d.current.components[1].code = d.current.components[0].code; expect(() => validateMacroScore(d)).toThrow();
    const e = scoreFixture(); e.current.modelHash = "b".repeat(64); expect(() => validateMacroScore(e)).toThrow();
    const f = scoreFixture(); (f.current.components[0] as any).score = null; f.current.components[0].contribution = 0; f.current.coverage = 80; f.current.score = 40; expect(() => validateMacroScore(f)).toThrow();
  });
  it("accepts an empty deployment and never uses a score fallback", async () => {
    const d = scoreFixture(); expect(validateMacroScore({ ...d, current: null }).current).toBeNull();
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response("unavailable", { status: 503 })));
    await expect(fetchMacroScore("daily")).rejects.toThrow("Chưa tải được điểm");
  });
});
