package wyckoff

import (
	"errors"
	"fmt"
	"math"
	"sort"
)

var ErrInsufficientData = errors.New("at least 30 valid candles are required")

type Config struct {
	RangeWindow int
	TrendWindow int
	VolumeWindow int
}

func DefaultConfig() Config { return Config{RangeWindow: 20, TrendWindow: 10, VolumeWindow: 20} }

type Analyzer struct { cfg Config }

func NewAnalyzer(cfg Config) *Analyzer {
	if cfg.RangeWindow < 5 { cfg.RangeWindow = 20 }
	if cfg.TrendWindow < 3 { cfg.TrendWindow = 10 }
	if cfg.VolumeWindow < 5 { cfg.VolumeWindow = 20 }
	return &Analyzer{cfg: cfg}
}

// Analyze applies transparent heuristic rules to OHLCV. Its confidence is a
// rule-strength score, not a statistically calibrated probability.
func (a *Analyzer) Analyze(symbol string, candles []Candle) (Analysis, error) {
	if len(candles) < 30 { return Analysis{}, ErrInsufficientData }
	for i, c := range candles {
		if c.High < c.Low || c.High < c.Open || c.High < c.Close || c.Low > c.Open || c.Low > c.Close || c.Close <= 0 || c.Volume < 0 {
			return Analysis{}, fmt.Errorf("invalid candle at index %d", i)
		}
	}
	candles = append([]Candle(nil), candles...)
	sort.Slice(candles, func(i,j int) bool { return candles[i].Date.Before(candles[j].Date) })
	last := candles[len(candles)-1]
	rw := a.cfg.RangeWindow
	if rw > len(candles)-a.cfg.TrendWindow { rw = len(candles)-a.cfg.TrendWindow }
	window := candles[len(candles)-rw:]
	support, resistance := math.MaxFloat64, -math.MaxFloat64
	for _, c := range window {
		if c.Low < support { support = c.Low }
		if c.High > resistance { resistance = c.High }
	}
	widthPct := (resistance-support)/((resistance+support)/2)*100
	tr := &TradingRange{Support:support, Resistance:resistance, WidthPct:widthPct}
	avgVolume := 0.0
	vw := a.cfg.VolumeWindow
	if vw > len(candles)-1 { vw = len(candles)-1 }
	for _, c := range candles[len(candles)-1-vw:len(candles)-1] { avgVolume += c.Volume }
	avgVolume /= float64(vw)
	prior := candles[len(candles)-1-a.cfg.TrendWindow]
	trendPct := (prior.Close-candles[0].Close)/candles[0].Close
	if len(candles) > a.cfg.TrendWindow { trendPct = (candles[len(candles)-a.cfg.TrendWindow-1].Close-candles[len(candles)-2*a.cfg.TrendWindow-1].Close) / candles[len(candles)-2*a.cfg.TrendWindow-1].Close }
	_ = trendPct
	result := Analysis{Symbol:symbol, AsOf:last.Date, Structure:StructureUnclear, Phase:PhaseUnknown, Confidence:0.25, TradingRange:tr, Events:[]Event{}, Evidence:[]string{}, DataPoints:len(candles), DataMode:"heuristic"}
	addEvent := func(kind string, confidence float64, evidence string) {
		result.Events = append(result.Events, Event{Type:kind, Date:last.Date, Confidence:confidence, Evidence:evidence})
		result.Evidence = append(result.Evidence, evidence)
	}
	// Look for a recent penetration and recovery relative to the range formed
	// before the last candle; this avoids using the current candle to set levels.
	prevWindow := candles[len(candles)-rw-1:len(candles)-1]
	prevSupport, prevResistance := math.MaxFloat64, -math.MaxFloat64
	for _, c := range prevWindow {
		if c.Low < prevSupport { prevSupport = c.Low }
		if c.High > prevResistance { prevResistance = c.High }
	}
	if last.Low < prevSupport && last.Close > prevSupport && last.Close > last.Open && last.Volume < avgVolume {
		addEvent("SPRING", 0.70, "Price pierced prior support and closed back above it on below-average volume.")
	}
	if last.High > prevResistance && last.Close < prevResistance && last.Close < last.Open && last.Volume > avgVolume {
		addEvent("UTAD", 0.68, "Price pierced prior resistance and closed back below it on above-average volume.")
	}
	if last.Close > resistance && last.Volume > avgVolume { addEvent("SOS", 0.65, "Close broke above the recent range on above-average volume.") }
	if last.Close < support && last.Volume > avgVolume { addEvent("SOW", 0.65, "Close broke below the recent range on above-average volume.") }

	// Conservative structure classification: only label a trading range when
	// the observed width is limited; otherwise report directional structure or unclear.
	if widthPct <= 25 {
		if len(result.Events)>0 {
			for _, e := range result.Events {
				if e.Type=="SPRING" || e.Type=="SOS" { result.Structure=StructureAccumulation }
				if e.Type=="UTAD" || e.Type=="SOW" { result.Structure=StructureDistribution }
			}
		} else {
			start := candles[len(candles)-rw-a.cfg.TrendWindow].Close
			change := (last.Close-start)/start
			if change < -0.08 { result.Structure=StructureAccumulation } else if change > 0.08 { result.Structure=StructureDistribution } else { result.Structure=StructureUnclear }
		}
	} else {
		start := candles[len(candles)-a.cfg.TrendWindow-1].Close
		change := (last.Close-start)/start
		if change > 0.08 { result.Structure=StructureMarkup; result.Phase=PhaseE; result.Evidence=append(result.Evidence,"Price has advanced materially over the recent trend window.") }
		if change < -0.08 { result.Structure=StructureMarkdown; result.Phase=PhaseE; result.Evidence=append(result.Evidence,"Price has declined materially over the recent trend window.") }
	}
	if result.Phase==PhaseUnknown {
		switch {
		case hasEvent(result.Events,"SPRING") || hasEvent(result.Events,"UTAD"): result.Phase=PhaseC
		case hasEvent(result.Events,"SOS") || hasEvent(result.Events,"SOW"): result.Phase=PhaseD
		case result.Structure==StructureAccumulation || result.Structure==StructureDistribution: result.Phase=PhaseB
		}
	}
	if result.Phase!=PhaseUnknown {
		result.Confidence=0.45
		if len(result.Events)>0 { result.Confidence=0.65 }
		if len(result.Events)>1 { result.Confidence=0.75 }
	} else {
		result.Evidence=append(result.Evidence,"Available OHLCV does not support a reliable Wyckoff phase classification.")
	}
	result.Evidence=append(result.Evidence,fmt.Sprintf("Recent trading range: support %.2f, resistance %.2f (width %.1f%%).",support,resistance,widthPct))
	return result,nil
}

func hasEvent(events []Event, kind string) bool { for _,e:=range events { if e.Type==kind { return true } }; return false }
