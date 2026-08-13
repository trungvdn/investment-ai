export const asOf = "13 August 2026";

const indicator = (name, group, value, unit, score, history, assessment) => ({ name, group, value, unit, score, history, assessment });

export const mockIndicators = [
  indicator("M2 Growth", "Liquidity", 9.6, "% YoY", 72, [8.7, 8.9, 9.1, 9.2, 9.4, 9.6], "Supportive liquidity"),
  indicator("Credit Growth", "Credit", 10.2, "% YTD", 68, [8.5, 8.9, 9.4, 9.7, 10.0, 10.2], "Credit expansion"),
  indicator("Interbank O/N Rate", "Liquidity", 3.1, "%", 61, [4.2, 3.8, 3.4, 3.2, 3.0, 3.1], "Normal funding conditions"),
  indicator("SBV T-bill", "Liquidity", 3.8, "%", 54, [2.9, 3.1, 3.4, 3.6, 3.7, 3.8], "Absorbing excess liquidity"),
  indicator("USD/VND", "External", 25880, "VND", 43, [25280, 25320, 25410, 25540, 25710, 25880], "Depreciation pressure"),
  indicator("CPI", "Inflation", 3.4, "% YoY", 61, [3.7, 3.5, 3.4, 3.3, 3.3, 3.4], "Within policy comfort zone"),
  indicator("Core CPI", "Inflation", 3.2, "% YoY", 63, [3.5, 3.4, 3.3, 3.2, 3.1, 3.2], "Underlying inflation contained"),
  indicator("PPI", "Inflation", 2.8, "% YoY", 58, [1.9, 2.1, 2.3, 2.5, 2.7, 2.8], "Pipeline price pressure rising"),
  indicator("PMI", "Growth", 51.8, "index", 67, [49.8, 50.4, 50.9, 51.2, 51.5, 51.8], "Manufacturing expansion"),
  indicator("IIP", "Growth", 8.6, "% YoY", 71, [6.2, 6.9, 7.4, 7.8, 8.2, 8.6], "Industrial momentum improving"),
  indicator("Retail Sales", "Growth", 9.1, "% YoY", 69, [7.2, 7.7, 8.1, 8.5, 8.8, 9.1], "Domestic demand resilient"),
  indicator("Export Growth", "External", 7.4, "% YoY", 65, [4.8, 5.3, 5.9, 6.4, 6.9, 7.4], "External demand recovering"),
  indicator("Fed Funds Rate", "External", 4.5, "%", 48, [5.25, 5.0, 4.75, 4.5, 4.5, 4.5], "US policy remains restrictive"),
  indicator("DXY", "External", 103.4, "index", 45, [101.2, 101.8, 102.1, 102.6, 103.0, 103.4], "Dollar strength persists"),
  indicator("US 10Y", "External", 4.28, "%", 49, [4.02, 4.08, 4.11, 4.16, 4.23, 4.28], "Global yields elevated")
];

export const mockScoreHistory = [
  { label: "Mar", score: 55 }, { label: "Apr", score: 57 }, { label: "May", score: 59 }, { label: "Jun", score: 61 }, { label: "Jul", score: 60 }, { label: "Aug", score: 59 }
];
