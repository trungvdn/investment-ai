package wyckoff

import "time"

// Candle is a single OHLCV observation. Volume is expressed in shares.
type Candle struct {
	Date   time.Time `json:"date"`
	Open   float64   `json:"open"`
	High   float64   `json:"high"`
	Low    float64   `json:"low"`
	Close  float64   `json:"close"`
	Volume float64   `json:"volume"`
}

type Structure string

const (
	StructureAccumulation  Structure = "ACCUMULATION"
	StructureDistribution  Structure = "DISTRIBUTION"
	StructureMarkup        Structure = "MARKUP"
	StructureMarkdown      Structure = "MARKDOWN"
	StructureUnclear       Structure = "UNCLEAR"
)

type Phase string

const (
	PhaseA       Phase = "A"
	PhaseB       Phase = "B"
	PhaseC       Phase = "C"
	PhaseD       Phase = "D"
	PhaseE       Phase = "E"
	PhaseUnknown Phase = "UNKNOWN"
)

type Event struct {
	Type       string    `json:"type"`
	Date       time.Time `json:"date"`
	Confidence float64   `json:"confidence"`
	Evidence   string    `json:"evidence"`
}

type Analysis struct {
	Symbol       string        `json:"symbol"`
	AsOf         time.Time     `json:"asOf"`
	Structure    Structure     `json:"structure"`
	Phase        Phase         `json:"phase"`
	Confidence   float64       `json:"confidence"`
	TradingRange *TradingRange `json:"tradingRange,omitempty"`
	Events       []Event       `json:"events"`
	Evidence     []string      `json:"evidence"`
	DataPoints   int           `json:"dataPoints"`
	DataMode     string        `json:"dataMode"`
}

type TradingRange struct {
	Support    float64 `json:"support"`
	Resistance float64 `json:"resistance"`
	WidthPct   float64 `json:"widthPct"`
}
