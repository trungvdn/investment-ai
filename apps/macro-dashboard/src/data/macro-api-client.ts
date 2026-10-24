import type {
  MacroDashboard,
  MacroIndicator,
  MacroObservation,
} from "../lib/types";
const safeURL = (value: unknown) => {
  try {
    const u = new URL(String(value));
    return (
      u.protocol === "https:" &&
      [
        "fred.stlouisfed.org",
        "nsdp.nso.gov.vn",
        "nso.gov.vn",
        "www.nso.gov.vn",
        "sbv.gov.vn",
        "www.sbv.gov.vn",
      ].includes(u.hostname)
    );
  } catch {
    return false;
  }
};
const validObservation = (o: MacroObservation, ind: MacroIndicator) =>
  o &&
  o.indicatorCode === ind.code &&
  Number.isFinite(o.value) &&
  o.unit === ind.unit &&
  typeof o.period === "string" &&
  Number.isFinite(Date.parse(o.periodEnd)) &&
  safeURL(o.sourceUrl);
export function validateMacroDashboard(value: unknown): MacroDashboard {
  const d = value as MacroDashboard;
  if (
    !d ||
    d.mode !== "live" ||
    d.score !== null ||
    !Array.isArray(d.indicators) ||
    !Number.isFinite(d.available) ||
    !Number.isFinite(d.enabled)
  )
    throw new Error("Macro API trả về định dạng không hợp lệ.");
  if (
    d.indicators.some(
      (i) =>
        !i ||
        typeof i.code !== "string" ||
        typeof i.name !== "string" ||
        typeof i.enabled !== "boolean" ||
        !Array.isArray(i.history) ||
        (i.latest !== null && !validObservation(i.latest, i)) ||
        i.history.some((o) => !validObservation(o, i)),
    )
  )
    throw new Error("Dữ liệu chỉ báo không hợp lệ hoặc thiếu nguồn gốc.");
  return d;
}
export async function fetchMacroDashboard({
  signal,
  fetchImpl = fetch,
}: {
  signal?: AbortSignal;
  fetchImpl?: typeof fetch;
} = {}): Promise<MacroDashboard> {
  let response: Response;
  try {
    response = await fetchImpl("/api/macro/dashboard", {
      signal,
      headers: { Accept: "application/json" },
    });
  } catch (error) {
    if (signal?.aborted) throw error;
    throw new Error(
      "Không kết nối được Macro API. Dữ liệu mẫu không được dùng thay thế.",
    );
  }
  if (!response.ok)
    throw new Error(
      response.status === 503
        ? "Kho dữ liệu vĩ mô chưa sẵn sàng. Kiểm tra kết nối PostgreSQL và Macro API."
        : `Không tải được dữ liệu vĩ mô (HTTP ${response.status}).`,
    );
  if (!response.headers.get("content-type")?.includes("application/json"))
    throw new Error("Macro API chưa được cấu hình tại /api/macro.");
  return validateMacroDashboard(await response.json());
}
