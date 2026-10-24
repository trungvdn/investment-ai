import { normalizeSymbol, validateReport } from "../domain/wyckoff-report.js";
export async function fetchStockReport({
  symbol,
  timeframe,
  from,
  to,
  signal,
  fetchImpl = fetch,
}) {
  symbol = normalizeSymbol(symbol);
  const query = new URLSearchParams({ timeframe, from, to });
  let response;
  try {
    response = await fetchImpl(
      `/api/stocks/${encodeURIComponent(symbol)}/report?${query}`,
      { signal, headers: { Accept: "application/json" } },
    );
  } catch (error) {
    if (signal?.aborted) throw error;
    throw new Error(
      "Không kết nối được dịch vụ phân tích. Hãy kiểm tra backend và thử lại.",
    );
  }
  const json = (response.headers.get("content-type") || "").includes(
    "application/json",
  );
  const body = json ? await response.json() : null;
  if (!response.ok) {
    if (response.status === 422)
      throw new Error(
        "Chưa đủ dữ liệu hợp lệ để phân tích. Khung tuần cần ít nhất 60 nến; hãy chọn khoảng 2 năm hoặc khung ngày.",
      );
    if (response.status === 502)
      throw new Error(
        "Nguồn dữ liệu thị trường hiện không khả dụng hoặc không có dữ liệu cho mã này. Kiểm tra mã và thử lại.",
      );
    throw new Error(
      body?.message || `Không tải được báo cáo (HTTP ${response.status}).`,
    );
  }
  if (!json)
    throw new Error(
      "Không tìm thấy API phân tích. Hãy mở UI từ dịch vụ stock-analysis hoặc cấu hình proxy /api.",
    );
  return validateReport(body, symbol, timeframe);
}
