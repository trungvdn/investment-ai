import { buildMacroDashboard } from "../domain/macro-dashboard.js";
import { asOf, mockIndicators, mockScoreHistory } from "../data/mock-macro-data.js";

const dashboard = buildMacroDashboard(mockIndicators, mockScoreHistory);
const scoreTone = (score) => score >= 65 ? "positive" : score >= 50 ? "neutral" : "negative";
const trendSymbol = (trend) => trend === "Rising" ? "↑" : trend === "Falling" ? "↓" : "→";

document.querySelector("#as-of").textContent = `As of ${asOf} · Mock data`;
document.querySelector("#macro-summary").innerHTML = `<article class="score-card"><p class="eyebrow">Macro score</p><strong>${dashboard.macroScore}</strong><span class="score-out-of">/ 100</span><p class="score-note">${dashboard.macroTrend} over the latest month</p></article><article class="regime-card"><p class="eyebrow">Macro regime</p><strong>${dashboard.regime}</strong><p class="score-note">Growth is supportive; external conditions remain a constraint.</p></article>`;
document.querySelector("#group-scores").innerHTML = dashboard.groupScores.map((group) => `<article class="group-card"><div><p>${group.name}</p><span>${group.count} indicator${group.count === 1 ? "" : "s"}</span></div><strong class="${scoreTone(group.score)}">${group.score}</strong><div class="progress"><i class="${scoreTone(group.score)}" style="width:${group.score}%"></i></div></article>`).join("");

const points = dashboard.scoreHistory.map((point, index) => `${index * 20},${100 - point.score}`).join(" ");
document.querySelector("#chart").innerHTML = `<svg viewBox="0 0 100 100" preserveAspectRatio="none" aria-hidden="true"><line x1="0" y1="25" x2="100" y2="25"/><line x1="0" y1="50" x2="100" y2="50"/><line x1="0" y1="75" x2="100" y2="75"/><polyline points="${points}" /></svg><div class="chart-labels">${dashboard.scoreHistory.map((point) => `<span>${point.label}<b>${point.score}</b></span>`).join("")}</div>`;
document.querySelector("#trend-copy").textContent = `${dashboard.macroTrend} · ${dashboard.scoreHistory.at(-1).score}/100`;
const alerts = dashboard.indicators.filter((indicator) => indicator.score < 55).sort((a, b) => a.score - b.score);
document.querySelector("#alerts").innerHTML = alerts.map((alert) => `<div class="alert"><span class="alert-dot"></span><div><strong>${alert.name}</strong><p>${alert.assessment}</p></div><b>${alert.score}</b></div>`).join("");
document.querySelector("#indicator-table").innerHTML = dashboard.indicators.map((indicator) => `<tr><td><strong>${indicator.name}</strong></td><td><span class="tag">${indicator.group}</span></td><td>${indicator.value.toLocaleString("en-US")} <small>${indicator.unit}</small></td><td class="${indicator.trend === "Falling" ? "down" : indicator.trend === "Rising" ? "up" : "flat"}">${trendSymbol(indicator.trend)} ${indicator.trend}</td><td><span class="assessment ${scoreTone(indicator.score)}">${indicator.assessment}</span></td></tr>`).join("");
