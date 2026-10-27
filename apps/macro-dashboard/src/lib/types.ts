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

export type MacroObservation = {
  upperValue?: number;
  indicatorCode: string;
  period: string;
  periodEnd: string;
  value: number;
  unit: string;
  source: string;
  sourceUrl: string;
  sourceSeriesId: string;
  method: string;
  publishedAt: string | null;
  firstSeenAt: string;
  collectedAt: string;
  revision: number;
  rawId: number;
};
export type MacroIndicator = {
  code: string;
  name: string;
  group: string;
  country: string;
  unit: string;
  frequency: string;
  comparisonBasis: string;
  enabled: boolean;
  maxAgeDays: number;
  description: string;
  status: string;
  latest: MacroObservation | null;
  history: MacroObservation[];
  lastAttempt: {
    indicatorCode: string;
    checkedAt: string;
    success: boolean;
    message: string;
  } | null;
};
export type MacroDashboard = {
  mode: "live";
  generatedAt: string;
  available: number;
  enabled: number;
  indicators: MacroIndicator[];
  score: null;
  scoreStatus: string;
};

export type MacroSourceState = {
  provider: string;
  nextDue: string;
  leaseUntil: string | null;
  lastChecked: string | null;
  lastSuccess: string | null;
  consecutiveFailures: number;
  errorKind: string;
  message: string;
};
export type MacroOperations = {
  generatedAt: string;
  sources: MacroSourceState[];
  alerts: {
    code?: string;
    provider?: string;
    status: string;
    message: string;
  }[];
};
