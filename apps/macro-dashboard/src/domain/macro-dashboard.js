export const GROUPS = ["Liquidity", "Credit", "Inflation", "Growth", "External"];

export function calculateTrend(history) {
  const delta = history.at(-1) - history.at(-2);
  if (Math.abs(delta) < 0.15) return "Stable";
  return delta > 0 ? "Rising" : "Falling";
}

export function determineRegime(score) {
  if (score >= 75) return "Supportive";
  if (score >= 55) return "Balanced";
  if (score >= 40) return "Cautious";
  return "Restrictive";
}

export function buildMacroDashboard(indicators, scoreHistory) {
  if (indicators.length !== 15) throw new Error("Macro Dashboard requires exactly 15 indicators.");
  const enriched = indicators.map((indicator) => ({ ...indicator, trend: calculateTrend(indicator.history) }));
  const groupScores = GROUPS.map((name) => {
    const members = enriched.filter((indicator) => indicator.group === name);
    if (!members.length) throw new Error(`Missing ${name} indicators.`);
    return { name, score: Math.round(members.reduce((total, indicator) => total + indicator.score, 0) / members.length), count: members.length };
  });
  const macroScore = Math.round(groupScores.reduce((total, group) => total + group.score, 0) / groupScores.length);
  return { indicators: enriched, groupScores, macroScore, regime: determineRegime(macroScore), scoreHistory, macroTrend: calculateTrend(scoreHistory.map((point) => point.score)) };
}
