import type { MacroIndicator } from "../lib/types";
export const statusLabels: Record<string, string> = {
  AVAILABLE: "Có dữ liệu",
  STALE: "Kỳ dữ liệu đã cũ",
  SOURCE_ERROR: "Nguồn đang lỗi",
  MISSING: "Chờ thu thập",
  NOT_CONFIGURED: "Chưa cấu hình nguồn",
  ERROR: "Chưa lấy được dữ liệu",
  NOT_IMPLEMENTED: "Chưa tích hợp",
};
export const groupLabels: Record<string, string> = {
  Liquidity: "Thanh khoản",
  Credit: "Tín dụng",
  Inflation: "Lạm phát",
  Growth: "Tăng trưởng",
  External: "Quốc tế",
};
export function historyPoints(indicator: MacroIndicator, months: number) {
  const history = indicator.history;
  if (!history.length) return [];
  const last = new Date(history.at(-1)!.periodEnd);
  const cutoff = new Date(last);
  cutoff.setUTCMonth(cutoff.getUTCMonth() - months);
  const chosen = history.filter((o) => new Date(o.periodEnd) >= cutoff);
  if (indicator.frequency !== "MONTHLY")
    return chosen.map((o) => ({ period: o.period, value: o.value }));
  if (!chosen.length) return [];
  const byPeriod = new Map(chosen.map((o) => [o.period, o.value]));
  const out: { period: string; value: number | null }[] = [];
  const date = new Date(`${chosen[0].period}-01T00:00:00Z`);
  const stop = chosen.at(-1)!.period;
  while (date.toISOString().slice(0, 7) <= stop) {
    const period = date.toISOString().slice(0, 7);
    out.push({ period, value: byPeriod.get(period) ?? null });
    date.setUTCMonth(date.getUTCMonth() + 1);
  }
  return out;
}
export function previousChange(i: MacroIndicator) {
  const h = i.history;
  if (h.length < 2) return null;
  const a = h.at(-2)!,
    b = h.at(-1)!;
  if (i.frequency === "MONTHLY") {
    const next = new Date(`${a.period}-01T00:00:00Z`);
    next.setUTCMonth(next.getUTCMonth() + 1);
    if (next.toISOString().slice(0, 7) !== b.period) return null;
  }
  return b.value - a.value;
}
