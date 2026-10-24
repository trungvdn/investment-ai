export type Candle = {
  date: string;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
};
export type WyckoffEvent = {
  type: string;
  date: string;
  confidence: number;
  evidence: string;
};
export type StockReport = {
  schemaVersion: string;
  symbol: string;
  asOf: string;
  headline: string;
  summary: string;
  structure:
    "ACCUMULATION" | "DISTRIBUTION" | "MARKUP" | "MARKDOWN" | "UNCLEAR";
  phase: string;
  confidence: number;
  tradingRange?: {
    status: string;
    support: number;
    resistance: number;
    widthPct: number;
  };
  events: WyckoffEvent[];
  evidence: string[];
  risks: string[];
  watchlist: string[];
  dataQuality: {
    mode: string;
    dataPoints: number;
    status: string;
    notes: string[];
  };
  disclaimer: string;
  candles: Candle[];
  timeframe: "1D" | "1W";
};
export type AnalysisRequest = {
  symbol: string;
  timeframe: "1D" | "1W";
  period: string;
  from: string;
  to: string;
};
