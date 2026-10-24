import {
  escapeHTML as e,
  number,
  dateLabel,
} from "../domain/wyckoff-report.js";
export function renderChart(report) {
  const candles = report.candles;
  if (!candles.length)
    return '<p class="chart-note">Không có dữ liệu nến để hiển thị.</p>';
  const W = 920,
    H = 450,
    left = 14,
    right = 82,
    top = 40,
    bottom = 310;
  const range = report.tradingRange;
  const low = Math.min(
    ...candles.map((c) => c.low),
    range?.support ?? Infinity,
  );
  const high = Math.max(
    ...candles.map((c) => c.high),
    range?.resistance ?? -Infinity,
  );
  const span = Math.max(high - low, high * 0.02, 0.01);
  const min = low - span * 0.1,
    max = high + span * 0.1;
  const y = (p) => top + ((max - p) / (max - min)) * (bottom - top);
  const step = (W - left - right) / candles.length;
  const x = (i) => left + (i + 0.5) * step;
  const bw = Math.max(0.6, Math.min(9, step * 0.65));
  const maxVolume = Math.max(1, ...candles.map((c) => c.volume));
  let svg = `<svg class="price-chart" viewBox="0 0 ${W} ${H}" role="img" aria-label="Biểu đồ nến giá và khối lượng ${e(report.symbol)}. Giá tính theo nghìn đồng."><title>Giá và khối lượng ${e(report.symbol)}</title>`;
  for (let i = 0; i <= 5; i++) {
    const price = min + ((max - min) * i) / 5;
    svg += `<line x1="${left}" x2="${W - right}" y1="${y(price)}" y2="${y(price)}" stroke="#edf1f5"/><text x="${W - right + 12}" y="${y(price) + 4}" fill="#8895a4" font-size="11">${number(price)}</text>`;
  }
  if (range)
    svg += `<rect x="${left}" y="${y(range.resistance)}" width="${W - left - right}" height="${Math.max(0, y(range.support) - y(range.resistance))}" fill="#19866a" opacity=".035"/>`;
  candles.forEach((c, i) => {
    const color = c.close >= c.open ? "#258d77" : "#d57a70";
    svg += `<g><title>${e(dateLabel(c.date))} · O ${number(c.open)} · H ${number(c.high)} · L ${number(c.low)} · C ${number(c.close)} · KL ${number(c.volume, 0)}</title><line x1="${x(i)}" x2="${x(i)}" y1="${y(c.high)}" y2="${y(c.low)}" stroke="${color}"/><rect x="${x(i) - bw / 2}" y="${Math.min(y(c.open), y(c.close))}" width="${bw}" height="${Math.max(1, Math.abs(y(c.open) - y(c.close)))}" fill="${color}"/><rect x="${x(i) - bw / 2}" y="${405 - (c.volume / maxVolume) * 60}" width="${bw}" height="${(c.volume / maxVolume) * 60}" fill="${color}" opacity=".5"/></g>`;
  });
  if (range)
    for (const [value, label] of [
      [range.support, "HT"],
      [range.resistance, "KC"],
    ])
      svg += `<line x1="${left}" x2="${W - right}" y1="${y(value)}" y2="${y(value)}" stroke="#70a997" stroke-dasharray="5 5"/><text x="${left + 5}" y="${y(value) - 6}" fill="#4d8a78" font-size="10">${label} ${number(value)}</text>`;
  report.events.slice(-12).forEach((event, index) => {
    const i = candles.findIndex(
      (c) => c.date.slice(0, 10) === event.date.slice(0, 10),
    );
    if (i < 0) return;
    const yy = 8 + (index % 2) * 18,
      xx = Math.min(W - right - 35, Math.max(30, x(i)));
    svg += `<g class="event-marker" data-event="${report.events.length - Math.min(12, report.events.length) + index}" tabindex="0" role="button" aria-label="Xem sự kiện ${e(event.type)} ngày ${e(dateLabel(event.date))}"><line x1="${xx}" x2="${x(i)}" y1="${yy + 17}" y2="${y(candles[i].high) - 3}" stroke="#9bb1bf" stroke-dasharray="2 3"/><rect x="${xx - 24}" y="${yy}" width="48" height="17" rx="4" fill="#294b63"/><text x="${xx}" y="${yy + 12}" text-anchor="middle" fill="white" font-size="9">${e(event.type)}</text></g>`;
  });
  for (let j = 0; j < 5; j++) {
    const i = Math.round(((candles.length - 1) * j) / 4);
    svg += `<text x="${x(i)}" y="432" text-anchor="middle" fill="#8895a4" font-size="10">${e(dateLabel(candles[i].date))}</text>`;
  }
  svg +=
    '<text x="14" y="337" fill="#8895a4" font-size="10">KHỐI LƯỢNG (cổ phiếu)</text></svg>';
  return `<div class="chart-scroll">${svg}</div><div class="chart-legend"><span>● Nến tăng</span><span>● Nến giảm</span><span>┄ Vùng giá ${range?.status === "IDENTIFIED" ? "được nhận diện" : "tạm thời"}</span><span>Đơn vị giá: nghìn đồng</span></div><p class="chart-note">${candles.length} nến · Khung ${report.timeframe === "1W" ? "tuần (tuần cuối có thể chưa kết thúc)" : "ngày"}. Chọn nhãn sự kiện để xem bằng chứng. Tối đa 12 nhãn gần nhất; danh sách đầy đủ ở tab Sự kiện.</p><div id="chart-event" class="chart-tooltip" aria-live="polite">Các sự kiện là giả thuyết theo quy tắc, cần xác nhận bằng diễn biến tiếp theo.</div>`;
}
