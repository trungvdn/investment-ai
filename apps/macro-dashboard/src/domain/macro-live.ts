import type { MacroIndicator } from "../lib/types";
export const statusLabels: Record<string, string> = {
  AVAILABLE: "Có dữ liệu",
  STALE: "Kỳ dữ liệu đã cũ",
  SOURCE_ERROR: "Nguồn đang lỗi",
  MISSING: "Chờ thu thập",
  NOT_CONFIGURED: "Chưa cấu hình nguồn",
  ERROR: "Chưa lấy được dữ liệu",
  NOT_IMPLEMENTED: "Chưa tích hợp",
  NOT_COMPARABLE: "Chưa đủ kỳ cùng phương pháp",
  AWAITING_RELEASE: "Chờ công bố kỳ mới",
  NO_PUBLISHED_RATE: "Phiên không công bố lãi suất tín phiếu",
};
export const groupLabels: Record<string, string> = {
  Liquidity: "Thanh khoản",
  Credit: "Tín dụng",
  Inflation: "Lạm phát",
  Growth: "Tăng trưởng",
  External: "Quốc tế",
};
function periodStart(period: string, frequency: string) {
  if (frequency === "QUARTERLY") {
    const match = /^(\d{4})-Q([1-4])$/.exec(period);
    if (!match) return null;
    return new Date(Date.UTC(Number(match[1]), (Number(match[2]) - 1) * 3, 1));
  }
  const date = new Date(`${period}-01T00:00:00Z`);
  return Number.isNaN(date.getTime()) ? null : date;
}
function periodKey(date: Date, frequency: string) {
  return frequency === "QUARTERLY"
    ? `${date.getUTCFullYear()}-Q${Math.floor(date.getUTCMonth() / 3) + 1}`
    : date.toISOString().slice(0, 7);
}
export function historyPoints(indicator: MacroIndicator, months: number) {
  const history = indicator.history;
  if (!history.length) return [];
  const last = new Date(history.at(-1)!.periodEnd);
  const cutoff = new Date(last);
  cutoff.setUTCMonth(cutoff.getUTCMonth() - months);
  const chosen = history.filter((o) => new Date(o.periodEnd) >= cutoff);
  if (!["MONTHLY", "QUARTERLY"].includes(indicator.frequency))
    return chosen.map((o) => ({ period: o.period, value: o.value }));
  if (!chosen.length) return [];
  const byPeriod = new Map(chosen.map((o) => [o.period, o.value]));
  const out: { period: string; value: number | null }[] = [];
  const date = periodStart(chosen[0].period, indicator.frequency);
  if (!date) return [];
  const stop = chosen.at(-1)!.period;
  const step = indicator.frequency === "QUARTERLY" ? 3 : 1;
  while (periodKey(date, indicator.frequency) <= stop) {
    const period = periodKey(date, indicator.frequency);
    out.push({ period, value: byPeriod.get(period) ?? null });
    date.setUTCMonth(date.getUTCMonth() + step);
  }
  return out;
}
export function previousChange(i: MacroIndicator) {
  if (i.code === "VN_SBV_TBILL") return null; // The tenor basket can change between auctions.
  const h = i.history;
  if (h.length < 2) return null;
  const a = h.at(-2)!,
    b = h.at(-1)!;
  if (
    i.comparisonBasis === "YTD" &&
    a.period.slice(0, 4) !== b.period.slice(0, 4)
  )
    return null;
  if (["MONTHLY", "QUARTERLY"].includes(i.frequency)) {
    const next = periodStart(a.period, i.frequency);
    if (!next) return null;
    next.setUTCMonth(
      next.getUTCMonth() + (i.frequency === "QUARTERLY" ? 3 : 1),
    );
    if (periodKey(next, i.frequency) !== b.period) return null;
  }
  return b.value - a.value;
}
export const changeUnit = (i: MacroIndicator) =>
  i.unit.startsWith("%")
    ? "điểm %"
    : ["tỷ đồng", "VND/USD"].includes(i.unit)
      ? i.unit
      : "đơn vị";
export function sourceMessage(i: MacroIndicator) {
  if (i.status === "NOT_CONFIGURED") return "Nguồn dữ liệu chưa được cấu hình.";
  return (
    i.lastAttempt?.message.replace(
      /^(?:NOT_COMPARABLE|NO_PUBLISHED_RATE|AWAITING_RELEASE):\s*/,
      "",
    ) || "Chờ thu thập từ nguồn chính thức."
  );
}

export const sourceErrorLabels: Record<string, string> = {
  ACCESS_DENIED: "Nguồn từ chối truy cập",
  RATE_LIMIT: "Nguồn giới hạn tần suất",
  UPSTREAM: "Nguồn tạm thời lỗi",
  DNS: "Không phân giải được tên miền",
  TLS: "Lỗi chứng chỉ kết nối",
  TIMEOUT: "Hết thời gian chờ",
  BODY_READ: "Không đọc hết phản hồi",
  TRANSPORT: "Lỗi kết nối",
  HTTP: "Nguồn trả lỗi HTTP",
  DATA_QUALITY: "Dữ liệu không đạt kiểm tra",
  NOT_CONFIGURED: "Chưa cấu hình nguồn",
  STORAGE: "Lỗi kho dữ liệu",
};
