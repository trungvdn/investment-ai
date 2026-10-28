import type { MacroObservation } from "../lib/types";
export type ScoreComponent = {
  code: string; name: string; group: string; weight: number; critical: boolean;
  score: number | null; contribution: number; effectiveWeight: number;
  reason: string; inputs: MacroObservation[]; warnings: string[];
};
export type ScoreSnapshot = {
  asOf: string; modelVersion: string; modelHash: string; inputHash: string;
  score: number | null; status: "READY" | "PROVISIONAL" | "INSUFFICIENT";
  coverage: number; components: ScoreComponent[];
  groups: { name: string; weight: number; coverage: number; score: number | null }[];
};
export type ScoreResponse = {
  model: { version: string; hash: string; minCoverage: number; rules: { code: string; weight: number; critical: boolean; mode: string }[] };
  frequency: "daily" | "weekly"; current: ScoreSnapshot | null;
  history: (ScoreSnapshot & { kind: "DAILY" | "WEEKLY" })[];
};
const bounded = (v: unknown): v is number => typeof v === "number" && Number.isFinite(v) && v >= 0 && v <= 100;
const nullable = (v: unknown) => v === null || bounded(v);
const hash = (v: unknown) => typeof v === "string" && /^[a-f0-9]{64}$/.test(v);
export function validateMacroScore(value: unknown): ScoreResponse {
  const d = value as ScoreResponse;
  const fail = () => { throw new Error("Dữ liệu điểm vĩ mô không hợp lệ."); };
  if (!d || !d.model || typeof d.model.version !== "string" || !hash(d.model.hash)
    || !bounded(d.model.minCoverage) || !["daily", "weekly"].includes(d.frequency)
    || !Array.isArray(d.model.rules) || !Array.isArray(d.history)) return fail();
  const rules = d.model.rules;
  if (new Set(rules.map(r => r.code)).size !== 8 || rules.some(r => !bounded(r.weight) || r.weight <= 0 || typeof r.code !== "string" || typeof r.critical !== "boolean")
    || Math.abs(rules.reduce((s, r) => s + r.weight, 0) - 100) > 0.001) return fail();
  const valid = (s: ScoreSnapshot) => {
    if (!s || !Number.isFinite(Date.parse(s.asOf)) || s.modelHash !== d.model.hash || s.modelVersion !== d.model.version || !hash(s.inputHash)
      || !nullable(s.score) || !bounded(s.coverage) || !["READY", "PROVISIONAL", "INSUFFICIENT"].includes(s.status)
      || !Array.isArray(s.components) || s.components.length !== rules.length || !Array.isArray(s.groups) || s.groups.length !== 3) return false;
    if (new Set(s.components.map(c => c.code)).size !== rules.length) return false;
    if (s.components.some(c => {
      const r = rules.find(r => r.code === c.code);
      return !r || c.weight !== r.weight || c.critical !== r.critical || typeof c.name !== "string" || typeof c.reason !== "string"
        || !nullable(c.score) || !bounded(c.contribution) || !bounded(c.effectiveWeight) || !Array.isArray(c.inputs) || !Array.isArray(c.warnings)
        || c.inputs.some(o => !o || o.indicatorCode !== c.code || !Number.isFinite(o.value) || typeof o.source !== "string" || !Number.isFinite(Date.parse(o.firstSeenAt)) || !Number.isInteger(o.revision) || o.revision < 1 || !Number.isInteger(o.rawId) || o.rawId <= 0);
    })) return false;
    const coverage = s.components.reduce((a, c) => a + (c.score === null ? 0 : c.weight), 0);
    const sum = s.components.reduce((a, c) => a + c.contribution, 0);
    if (Math.abs(coverage - s.coverage) > 0.001 || (s.score === null ? coverage !== 0 : Math.abs(sum - s.score) > 0.001)) return false;
    const ready = coverage >= d.model.minCoverage && s.components.every(c => !c.critical || c.score !== null);
    if (s.status !== (coverage === 0 ? "INSUFFICIENT" : ready ? "READY" : "PROVISIONAL")) return false;
    return s.groups.every(g => typeof g.name === "string" && bounded(g.weight) && bounded(g.coverage) && nullable(g.score));
  };
  if ((d.current !== null && !valid(d.current)) || d.history.some(s => !valid(s) || s.kind !== (d.frequency === "daily" ? "DAILY" : "WEEKLY"))) return fail();
  return d;
}
export async function fetchMacroScore(frequency: "daily" | "weekly", signal?: AbortSignal): Promise<ScoreResponse> {
  const response = await fetch(`/api/macro/score?frequency=${frequency}`, { signal, headers: { Accept: "application/json" } });
  if (!response.ok || !response.headers.get("content-type")?.includes("application/json")) throw new Error("Chưa tải được điểm vĩ mô. Kiểm tra Macro API và migration 011.");
  const result = validateMacroScore(await response.json());
  if (result.frequency !== frequency) throw new Error("Kỳ lịch sử điểm không khớp yêu cầu.");
  return result;
}
