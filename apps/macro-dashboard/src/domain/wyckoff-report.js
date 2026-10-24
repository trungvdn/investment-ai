export const structureLabels = {
  ACCUMULATION: "Tích lũy tiềm năng",
  DISTRIBUTION: "Phân phối tiềm năng",
  MARKUP: "Xu hướng tăng tiềm năng",
  MARKDOWN: "Xu hướng giảm tiềm năng",
  UNCLEAR: "Chưa xác định cấu trúc",
};
export const escapeHTML = (value) =>
  String(value ?? "").replace(
    /[&<>"']/g,
    (c) =>
      ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[
        c
      ],
  );
export const number = (value, digits = 2) =>
  Number.isFinite(value)
    ? value.toLocaleString("vi-VN", { maximumFractionDigits: digits })
    : "—";
export const dateLabel = (value) => {
  const d = new Date(value);
  return Number.isNaN(d.getTime())
    ? "—"
    : d.toLocaleDateString("vi-VN", { timeZone: "Asia/Ho_Chi_Minh" });
};
export function normalizeSymbol(value) {
  const symbol = value.trim().toUpperCase();
  if (!/^[A-Z][A-Z0-9]{1,9}$/.test(symbol))
    throw new Error(
      "Nhập mã cổ phiếu hợp lệ, chỉ gồm 2–10 chữ cái hoặc chữ số.",
    );
  return symbol;
}
export function dateRange(months, now = new Date()) {
  const to = new Date(now);
  const from = new Date(now);
  const day = from.getUTCDate();
  from.setUTCDate(1);
  from.setUTCMonth(from.getUTCMonth() - Number(months));
  const last = new Date(
    Date.UTC(from.getUTCFullYear(), from.getUTCMonth() + 1, 0),
  ).getUTCDate();
  from.setUTCDate(Math.min(day, last));
  return {
    from: from.toISOString().slice(0, 10),
    to: to.toISOString().slice(0, 10),
  };
}
export function validateReport(report, symbol, timeframe) {
  if (
    !report ||
    report.schemaVersion !== "1.0" ||
    report.symbol !== symbol ||
    !structureLabels[report.structure] ||
    !Array.isArray(report.events) ||
    !Array.isArray(report.evidence) ||
    !report.dataQuality ||
    !Number.isFinite(report.confidence) ||
    report.confidence < 0 ||
    report.confidence > 1
  )
    throw new Error("Báo cáo trả về không đúng định dạng được hỗ trợ.");
  if (!Array.isArray(report.candles) || report.timeframe !== timeframe)
    throw new Error(
      "Backend chưa hỗ trợ biểu đồ/khung thời gian. Vui lòng cập nhật stock-analysis cùng phiên bản UI.",
    );
  if (
    report.candles.some(
      (c) =>
        !Number.isFinite(Date.parse(c.date)) ||
        ![c.open, c.high, c.low, c.close, c.volume].every(Number.isFinite) ||
        c.volume < 0 ||
        c.low > Math.min(c.open, c.close) ||
        c.high < Math.max(c.open, c.close),
    )
  )
    throw new Error("Dữ liệu nến không hợp lệ.");
  return report;
}
export function scenarios(range) {
  if (!range) return [];
  const support = number(range.support),
    resistance = number(range.resistance);
  return [
    {
      title: "01 / Vượt kháng cự",
      tone: "",
      trigger: `Giá đóng cửa vượt ${resistance}, kèm khối lượng mở rộng và lực cầu duy trì.`,
      watch:
        "Theo dõi nhịp kiểm tra lại vùng vượt; không xem một phiên vượt là xác nhận đầy đủ.",
      invalid: `Giá quay lại trong vùng ${support}–${resistance} và không phục hồi.`,
    },
    {
      title: "02 / Tiếp tục trong vùng",
      tone: "sideways",
      trigger: `Giá tiếp tục dao động giữa ${support} và ${resistance}.`,
      watch:
        "Quan sát thay đổi cung cầu ở hai biên, đặc biệt khối lượng trong các nhịp kiểm tra.",
      invalid: "Giá rời vùng và duy trì ngoài biên với bằng chứng tiếp diễn.",
    },
    {
      title: "03 / Mất hỗ trợ",
      tone: "negative",
      trigger: `Giá đóng cửa dưới ${support}, kèm áp lực bán tiếp diễn.`,
      watch:
        "Kiểm tra khả năng phục hồi vào vùng; phân biệt phá vỡ thật với Spring tiềm năng.",
      invalid:
        "Giá phục hồi vào vùng và giữ được hỗ trợ trong lần kiểm tra tiếp theo.",
    },
  ];
}
