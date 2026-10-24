import { fetchStockReport } from "../data/stock-api-client.js";
import {
  escapeHTML as e,
  normalizeSymbol,
  dateRange,
  number,
  dateLabel,
  structureLabels,
  scenarios,
} from "../domain/wyckoff-report.js";
import { renderChart } from "./wyckoff-chart.js";
const $ = (selector) => document.querySelector(selector);
const form = $("#analysis-form"),
  output = $("#report"),
  status = $("#status");
let activeRequest;
const list = (items, risk = false) =>
  `<ul class="evidence-list ${risk ? "risk-list" : ""}">${(items?.length ? items : ["Chưa có bằng chứng được cung cấp."]).map((v) => `<li>${e(v)}</li>`).join("")}</ul>`;
const panel = (title, content) =>
  `<article class="panel"><h3>${title}</h3>${content}</article>`;
function render(report, request) {
  const r = report,
    q = r.dataQuality,
    range = r.tradingRange,
    last = r.candles.at(-1);
  const phase = /^[A-E]$/.test(r.phase)
    ? `Phase ${r.phase}`
    : "Chưa xác định pha";
  const isUnknown = r.structure === "UNCLEAR" || r.phase === "UNKNOWN";
  const scenarioCards = scenarios(range);
  output.innerHTML = `${q.mode === "demo" ? '<div class="demo-banner"><b>DỮ LIỆU MINH HỌA</b> · Nến tổng hợp để kiểm tra giao diện, không phản ánh cổ phiếu thực tế.</div>' : ""}
    <div class="report-header"><div class="stock-title"><div class="stock-avatar">${e(r.symbol)}</div><div><h2>${e(r.symbol)} <span class="badge">${e(phase)}</span></h2><div class="meta">Báo cáo Wyckoff · Dữ liệu đến ${dateLabel(r.asOf)} · ${e(request.from)} → ${e(request.to)}</div></div></div><div class="price">${number(last?.close)}<small>Giá đóng cửa · nghìn đồng</small></div></div>
    <article class="summary"><div><p class="eyebrow">NHẬN ĐỊNH CHÍNH / GIẢ THUYẾT PHÂN TÍCH</p><h2>${e(structureLabels[r.structure])}</h2><p>${e(r.summary)}</p></div><div class="summary-tags"><span class="badge">${isUnknown ? "Cần thêm bằng chứng" : "Cần diễn biến xác nhận"}</span><span>Khớp quy tắc: <b>${number(r.confidence * 100, 0)}/100</b></span><span class="muted">Không phải xác suất thành công</span></div></article>
    <div class="chart-grid"><article class="panel"><div class="panel-heading"><h2>Giá & khối lượng</h2><span class="caption">${e(r.symbol)} · ${e(r.timeframe)} · ${q.mode === "demo" ? "Minh họa" : "KB Securities"}</span></div>${renderChart(r)}</article><aside class="panel metric-panel"><p class="eyebrow">ĐỌC CẤU TRÚC</p><h2>Bức tranh hiện tại</h2><div class="phase-row">${["A", "B", "C", "D", "E"].map((p) => `<span class="${r.phase === p ? "active" : ""}">${p}</span>`).join("")}</div><p class="caption">${e(phase)} · Nhận diện theo quy tắc</p><div class="metrics"><div class="metric"><span>Cấu trúc</span><strong>${e(structureLabels[r.structure])}</strong></div><div class="metric"><span>Hỗ trợ</span><strong>${number(range?.support)}</strong></div><div class="metric"><span>Kháng cự</span><strong>${number(range?.resistance)}</strong></div><div class="metric"><span>Trạng thái vùng</span><strong>${range ? (range.status === "IDENTIFIED" ? "Được nhận diện" : "Tạm thời") : "Chưa xác định"}</strong></div><div class="metric"><span>Sự kiện gần nhất</span><strong>${e(r.events.at(-1)?.type || "Chưa có")}</strong></div></div><div class="watch-note"><b>Điều kiện cần theo dõi</b>${e(r.watchlist?.[0] || "Chờ thêm dữ liệu giá và khối lượng.")}</div></aside></div>
    <div class="tabs" role="tablist" aria-label="Chi tiết báo cáo">${[
      ["overview", "Tổng quan"],
      ["events", "Sự kiện Wyckoff"],
      ["scenarios", "Kịch bản"],
      ["data", "Dữ liệu & phương pháp"],
    ]
      .map(
        ([id, label], i) =>
          `<button type="button" id="tab-${id}" role="tab" aria-selected="${i === 0}" aria-controls="content-${id}" tabindex="${i === 0 ? 0 : -1}">${label}</button>`,
      )
      .join("")}</div>
    <section id="content-overview" class="tab-content" role="tabpanel" aria-labelledby="tab-overview"><div class="detail-grid">${panel("Bằng chứng hỗ trợ", list(r.evidence))}${panel("Rủi ro & điểm cần thận trọng", list(r.risks, true))}${panel("Theo dõi tiếp theo", list(r.watchlist))}</div></section>
    <section id="content-events" class="tab-content" role="tabpanel" aria-labelledby="tab-events" hidden><p class="chart-note">Sự kiện do bộ quy tắc nhận diện; mức khớp quy tắc không đồng nghĩa với sự kiện đã được xác nhận.</p><div class="event-list">${r.events.length ? r.events.map((ev) => `<details class="event"><summary>${e(ev.type)}<span class="meta">${dateLabel(ev.date)} · Khớp quy tắc ${number(ev.confidence * 100, 0)}/100 · Giả thuyết</span></summary><p>${e(ev.evidence)}</p></details>`).join("") : '<div class="panel">Chưa nhận diện được sự kiện Wyckoff trong khoảng dữ liệu này.</div>'}</div></section>
    <section id="content-scenarios" class="tab-content" role="tabpanel" aria-labelledby="tab-scenarios" hidden><p class="chart-note">Các điều kiện quan sát dựa trên ${range?.status === "IDENTIFIED" ? "vùng được nhận diện" : "vùng tạm thời"}; không phải dự báo hoặc lệnh giao dịch.</p><div class="scenario-grid">${scenarioCards.length ? scenarioCards.map((s) => `<article class="panel scenario ${s.tone}"><h3>${s.title}</h3><p><b>Điều kiện</b><br>${e(s.trigger)}</p><p><b>Theo dõi</b><br>${e(s.watch)}</p><p><b>Vô hiệu khi</b><br>${e(s.invalid)}</p></article>`).join("") : '<div class="panel">Chưa có trading range để xây dựng kịch bản theo mức giá. Chờ vùng hỗ trợ/kháng cự được nhận diện.</div>'}</div></section>
    <section id="content-data" class="tab-content" role="tabpanel" aria-labelledby="tab-data" hidden><div class="detail-grid">${panel("Nguồn & chất lượng dữ liệu", `<p class="meta">Nguồn: ${q.mode === "demo" ? "Nến tổng hợp (demo)" : "KB Securities · endpoint công khai"}<br>Số nến phân tích: ${number(q.dataPoints, 0)}<br>Khung: ${e(r.timeframe)}<br>Trạng thái: ${e(q.status)}<br>Ngày dữ liệu: ${dateLabel(r.asOf)}<br>Schema: ${e(r.schemaVersion)}</p>${list(q.notes)}`)}${panel("Phương pháp & giới hạn", '<p class="meta">Phân tích dựa trên quy tắc cấu trúc giá, khối lượng và trình tự sự kiện Wyckoff. Confidence đo mức khớp quy tắc, không phải xác suất. Có thể không xác định được cấu trúc hoặc pha.</p><p class="meta">Khung tuần tổng hợp OHLCV theo tuần ISO trước khi phân tích; tuần cuối có thể chưa hoàn tất. Cần ít nhất 60 nến.</p><p class="meta">Nguồn giá chưa bảo đảm điều chỉnh đầy đủ quyền/cổ tức. Số nến đủ không bảo đảm dữ liệu liên tục. Nội dung bằng chứng gốc từ bộ phân tích được giữ nguyên.</p>')}</div></section><p class="disclaimer">${e(r.disclaimer)}</p>`;
  output.hidden = false;
  const tabs = [...output.querySelectorAll('[role="tab"]')];
  function selectTab(tab) {
    tabs.forEach((t) => {
      const selected = t === tab;
      t.setAttribute("aria-selected", String(selected));
      t.tabIndex = selected ? 0 : -1;
      document.getElementById(t.getAttribute("aria-controls")).hidden =
        !selected;
    });
  }
  tabs.forEach((tab, i) => {
    tab.addEventListener("click", () => selectTab(tab));
    tab.addEventListener("keydown", (ev) => {
      let next;
      if (ev.key === "ArrowRight") next = (i + 1) % tabs.length;
      if (ev.key === "ArrowLeft") next = (i + tabs.length - 1) % tabs.length;
      if (ev.key === "Home") next = 0;
      if (ev.key === "End") next = tabs.length - 1;
      if (next !== undefined) {
        ev.preventDefault();
        selectTab(tabs[next]);
        tabs[next].focus();
      }
    });
  });
  output.querySelectorAll("[data-event]").forEach((marker) => {
    const show = () => {
      const ev = r.events[Number(marker.dataset.event)];
      $("#chart-event").textContent =
        `${ev.type} · ${dateLabel(ev.date)} · ${ev.evidence}`;
    };
    marker.addEventListener("click", show);
    marker.addEventListener("keydown", (ev) => {
      if (ev.key === "Enter" || ev.key === " ") {
        ev.preventDefault();
        show();
      }
    });
  });
}
form.addEventListener("submit", async (event) => {
  event.preventDefault();
  let symbol;
  try {
    symbol = normalizeSymbol($("#symbol").value);
  } catch (error) {
    status.className = "error";
    status.textContent = error.message;
    return;
  }
  activeRequest?.abort();
  const controller = new AbortController();
  activeRequest = controller;
  const timeout = setTimeout(() => controller.abort("timeout"), 45000);
  const request = {
    symbol,
    timeframe: $("#timeframe").value,
    ...dateRange($("#period").value),
  };
  $("#symbol").value = symbol;
  output.hidden = true;
  $("#empty").hidden = true;
  $("#analyze").disabled = true;
  output.setAttribute("aria-busy", "true");
  status.className = "";
  status.textContent = `Đang lấy dữ liệu ${symbol} và phân tích cấu trúc giá, khối lượng…`;
  try {
    const report = await fetchStockReport({
      ...request,
      signal: controller.signal,
    });
    if (activeRequest !== controller) return;
    render(report, request);
    status.textContent = "";
    history.replaceState(
      null,
      "",
      `?symbol=${encodeURIComponent(symbol)}&timeframe=${request.timeframe}&period=${$("#period").value}`,
    );
  } catch (error) {
    if (activeRequest !== controller) return;
    status.className = "error";
    status.textContent = controller.signal.aborted
      ? "Yêu cầu quá thời gian chờ. Vui lòng thử lại."
      : error.message;
    $("#empty").hidden = false;
  } finally {
    clearTimeout(timeout);
    if (activeRequest === controller) {
      $("#analyze").disabled = false;
      output.setAttribute("aria-busy", "false");
    }
  }
});
document.querySelectorAll("[data-symbol]").forEach((button) =>
  button.addEventListener("click", () => {
    $("#symbol").value = button.dataset.symbol;
    form.requestSubmit();
  }),
);
$("#timeframe").addEventListener("change", () => {
  if ($("#timeframe").value === "1W") $("#period").value = "24";
});
const params = new URLSearchParams(location.search);
if (params.has("symbol")) {
  $("#symbol").value = params.get("symbol");
  if (["1D", "1W"].includes(params.get("timeframe")))
    $("#timeframe").value = params.get("timeframe");
  if (["6", "12", "24"].includes(params.get("period")))
    $("#period").value = params.get("period");
  form.requestSubmit();
}
