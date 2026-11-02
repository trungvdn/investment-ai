package wyckoff

import (
	"errors"
	"math"
	"sort"
)

var ErrInsufficientData = errors.New("insufficient valid candles")

// EventLookback is the maximum history considered for the active Wyckoff cycle.
const EventLookback = 70

type Config struct {
	RangeWindow, TrendWindow, VolumeWindow int
}

func DefaultConfig() Config {
	return Config{RangeWindow: 40, TrendWindow: 20, VolumeWindow: 20}
}

type Analyzer struct {
	cfg Config
}

func NewAnalyzer(cfg Config) *Analyzer {
	if cfg.RangeWindow < 20 {
		cfg.RangeWindow = 40
	}
	if cfg.TrendWindow < 5 {
		cfg.TrendWindow = 20
	}
	if cfg.VolumeWindow < 5 {
		cfg.VolumeWindow = 20
	}
	return &Analyzer{cfg: cfg}
}

type candleStats struct {
	spread, closeLocation, relativeVolume float64
}

func stats(c Candle, avgVol float64) candleStats {
	spread := c.High - c.Low
	clv := 0.5
	if spread > 0 {
		clv = (c.Close - c.Low) / spread
	}
	rv := 0.0
	if avgVol > 0 {
		rv = c.Volume / avgVol
	}
	return candleStats{spread: spread, closeLocation: clv, relativeVolume: rv}
}
func avgVolume(cs []Candle) float64 {
	total := 0.0
	for _, c := range cs {
		total += c.Volume
	}
	if len(cs) == 0 {
		return 0
	}
	return total / float64(len(cs))
}
func bounds(cs []Candle) (float64, float64) {
	lo, hi := math.MaxFloat64, -math.MaxFloat64
	for _, c := range cs {
		if c.Low < lo {
			lo = c.Low
		}
		if c.High > hi {
			hi = c.High
		}
	}
	return lo, hi
}

// orderEventsForEvidence returns a copy ordered for human-readable Wyckoff evidence.
// Unknown event types follow the canonical sequence; same-type events are chronological.
func orderEventsForEvidence(events []Event) []Event {
	order := map[string]int{"PS": 0, "SC": 1, "AR": 2, "ST": 3, "SPRING": 4, "SOS": 5, "LPS": 6}
	ordered := append([]Event(nil), events...)
	sort.SliceStable(ordered, func(i, j int) bool {
		iOrder, iOK := order[ordered[i].Type]
		jOrder, jOK := order[ordered[j].Type]
		if !iOK {
			iOrder = len(order)
		}
		if !jOK {
			jOrder = len(order)
		}
		if iOrder != jOrder {
			return iOrder < jOrder
		}
		return ordered[i].Date.Before(ordered[j].Date)
	})
	return ordered
}

func hasEvent(es []Event, kind string) bool {
	for _, e := range es {
		if e.Type == kind {
			return true
		}
	}
	return false
}
func eventIndex(es []Event, kind string) int {
	for i := range es {
		if es[i].Type == kind {
			return i
		}
	}
	return -1
}
func abs(v float64) float64 {
	return math.Abs(v)
}

type rangeCandidate struct {
	start, end                             int
	support, resistance, widthPct, quality float64
	preTrend                               float64
	valid                                  bool
}

// detectTradingRange looks for a reasonably stable recent range after a directional
// move. This is a heuristic candidate detector, not a discretionary chart label.
func (a *Analyzer) detectTradingRange(cs []Candle, trendChange float64) rangeCandidate {
	n := len(cs)
	// Exclude the last five bars from range construction so a fresh breakout
	// does not erase the range it is breaking out of.
	end := n - 5
	if end < 1 {
		end = n
	}
	w := a.cfg.RangeWindow
	if w > end {
		w = end
	}
	start := end - w
	window := cs[start:end]
	support, resistance := bounds(window)
	mid := (support + resistance) / 2
	width := (resistance - support) / math.Max(mid, 0.00001) * 100
	// Count how many candles overlap the inner 80% of the candidate range and
	// how often price reacts near either edge.
	innerLow := support + (resistance-support)*0.1
	innerHigh := resistance - (resistance-support)*0.1
	inside, lowTouches, highTouches := 0, 0, 0
	for _, c := range window {
		if c.Close >= innerLow && c.Close <= innerHigh {
			inside++
		}
		if c.Low <= support+(resistance-support)*0.12 {
			lowTouches++
		}
		if c.High >= resistance-(resistance-support)*0.12 {
			highTouches++
		}
	}
	insideRatio := float64(inside) / float64(len(window))
	touchScore := math.Min(1, float64(lowTouches+highTouches)/6)
	// A range needs multiple reactions and price overlap; extremes alone are not enough.
	quality := 0.65*insideRatio + 0.35*touchScore
	preStart := start - w
	preTrend := 0.0
	if preStart < 0 {
		preStart = 0
	}
	if start > preStart && cs[preStart].Close > 0 {
		preTrend = (cs[start].Close - cs[preStart].Close) / cs[preStart].Close
	}
	valid := width >= 4 && width <= 35 && insideRatio >= 0.35 && lowTouches >= 2 && highTouches >= 2 &&
		abs(trendChange) <= 0.20
	// If the range is actively expanding into a breakout, don't call it a range.
	return rangeCandidate{start: start, end: end - 1, support: support, resistance: resistance, widthPct: width, quality: quality, preTrend: preTrend, valid: valid}
}

// Analyze identifies a candidate range, then evaluates events within that
// recent context and checks their temporal order. Confidence is heuristic only.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
