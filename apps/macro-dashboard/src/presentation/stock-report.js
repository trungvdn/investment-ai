const $ = (selector) => document.querySelector(selector);

function escapeHTML(value) {
  return String(value ?? "").replace(/[&<>"']/g, (char) => ({
    "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;",
  })[char]);
}
function pretty(value) {
  if (value === null || value === undefined || value === "") return "—";
  return String(value).replaceAll("_", " ").toLowerCase().replace(/\b\w/g, (c) => c.toUpperCase());
}
function listItems(items, emptyText) {
  if (!Array.isArray(items) || items.length === 0) return `<p class="report-empty">${escapeHTML(emptyText)}</p>`;
  return `<ul class="report-list">${items.map((item) => {
    if (typeof item === "string") return `<li>${escapeHTML(item)}</li>`;
    const title = item.title || item.name || item.type || item.event || item.code || "Detail";
    const detail = item.description || item.evidence || item.reason || item.message || item.date || "";
    return `<li><strong>${escapeHTML(pretty(title))}</strong>${detail ? `<p>${escapeHTML(detail)}</p>` : ""}</li>`;
  }).join("")}</ul>`;
}
function panel(title, body) {
  return `<article class="panel report-panel"><h3>${escapeHTML(title)}</h3>${body}</article>`;
}
function renderReport(report) {
  const range = report.tradingRange || {};
  const metrics = [
    ["Structure", pretty(report.structure)],
    ["Wyckoff phase", pretty(report.phase)],
    ["Confidence", report.confidence == null ? "—" : `${report.confidence}%`],
    ["Trading range", pretty(range.status)],
  ];
  const events = (report.events || []).map((event) => ({
    title: event.type || event.name || event.event || "Event",
    description: [event.date, event.confidence == null ? "" : `Confidence: ${event.confidence}%`, event.evidence].filter(Boolean).join(" · "),
  }));
  const dataMode = report.dataQuality?.mode || report.dataQuality?.dataMode || "unknown";
  const pointCount = report.dataQuality?.pointCount ?? report.dataQuality?.historyCount;
  const warnings = [];
  if (String(dataMode).toLowerCase() === "demo") warnings.push("This report uses demo data. Do not use it to make investment decisions.");
  if (range.status === "PROVISIONAL") warnings.push("The trading range is provisional, not a confirmed structure.");
  const quality = `<p><strong>Data mode:</strong> ${escapeHTML(pretty(dataMode))}${pointCount == null ? "" : ` · <strong>History points:</strong> ${escapeHTML(pointCount)}`}</p>`;
  $("#stock-report-result").innerHTML = `
    <div class="stock-report-hero">
      <article class="panel">
        <p class="eyebrow">${escapeHTML(report.symbol || "Stock")} · ${escapeHTML(report.asOf || "Latest analysis")}</p>
        <h2 class="stock-headline">${escapeHTML(report.headline || "Stock analysis")}</h2>
        <p class="stock-summary">${escapeHTML(report.summary || "No summary was returned.")}</p>
        <div style="margin-top:14px"><span class="report-pill">Schema ${escapeHTML(report.schemaVersion || "unknown")}</span></div>
      </article>
      <article class="panel"><div class="stock-metrics">${metrics.map(([label, value]) => `<div class="stock-metric"><span>${escapeHTML(label)}</span><strong>${escapeHTML(value)}</strong></div>`).join("")}</div></article>
    </div>
    ${warnings.length ? `<div class="report-warning">${warnings.map(escapeHTML).join("<br>")}</div>` : ""}
    <div class="stock-report-grid">
      ${panel("Trading range", `<p><strong>Status:</strong> ${escapeHTML(pretty(range.status))}</p><p style="margin-top:8px"><strong>Range:</strong> ${escapeHTML(range.low ?? "—")} – ${escapeHTML(range.high ?? "—")}</p>${range.description ? `<p class="stock-summary" style="margin-top:8px">${escapeHTML(range.description)}</p>` : ""}`)}
      ${panel("Wyckoff events", listItems(events, "No events were reported."))}
      ${panel("Evidence", listItems(report.evidence, "No evidence details were reported."))}
      ${panel("Risks", listItems(report.risks, "No explicit risks were reported."))}
      ${panel("Watchlist", listItems(report.watchlist, "No follow-up conditions were reported."))}
      ${panel("Data quality", `${quality}${listItems(report.dataQuality?.warnings || report.dataQuality?.issues, "No additional data-quality warnings.")}`)}
    </div>
    <p class="report-disclaimer">${escapeHTML(report.disclaimer || "For research purposes only; not investment advice.")}</p>`;
}
export function initStockReport() {
  const form = $("#stock-report-form");
  if (!form) return;
  const status = $("#stock-report-status");
  const submit = $("#stock-report-submit");
  form.addEventListener("submit", async (event) => {
    event.preventDefault();
    const symbol = $("#stock-symbol").value.trim().toUpperCase();
    const apiBase = $("#stock-api-base").value.trim().replace(/\/+$/, "");
    const from = $("#stock-from").value;
    const to = $("#stock-to").value;
    $("#stock-symbol").value = symbol;
    $("#stock-report-result").replaceChildren();
    if (!/^[A-Z0-9.^_-]{1,12}$/.test(symbol)) {
      status.dataset.state = "error";
      status.textContent = "Enter a valid stock symbol (letters, numbers, dot, underscore or hyphen).";
      return;
    }
    if (from && to && from > to) {
      status.dataset.state = "error";
      status.textContent = "The start date must not be after the end date.";
      return;
    }
    let base;
    try { base = new URL(apiBase); } catch {
      status.dataset.state = "error";
      status.textContent = "Enter a valid API URL, for example http://localhost:8081.";
      return;
    }
    if (!["http:", "https:"].includes(base.protocol)) {
      status.dataset.state = "error";
      status.textContent = "The API URL must use HTTP or HTTPS.";
      return;
    }
    const query = new URLSearchParams();
    if (from) query.set("from", from);
    if (to) query.set("to", to);
    const url = `${base.origin}/api/stocks/${encodeURIComponent(symbol)}/report${query.size ? `?${query}` : ""}`;
    submit.disabled = true;
    submit.textContent = "Analyzing…";
    status.dataset.state = "loading";
    status.textContent = `Requesting report for ${symbol}…`;
    try {
      const response = await fetch(url, { headers: { Accept: "application/json" } });
      const payload = await response.json().catch(() => ({}));
      if (!response.ok) {
        const detail = payload.message || payload.error || `HTTP ${response.status}`;
        throw new Error(detail);
      }
      if (!payload || typeof payload !== "object" || !payload.schemaVersion) {
        throw new Error("The API response does not match report schema v1.0.");
      }
      renderReport(payload);
      status.dataset.state = "success";
      status.textContent = `Report loaded for ${symbol}.`;
    } catch (error) {
      status.dataset.state = "error";
      status.textContent = `Unable to load report: ${error.message}. Check that the Go API is running and reachable from this browser.`;
    } finally {
      submit.disabled = false;
      submit.textContent = "Analyze stock";
    }
  });
}
