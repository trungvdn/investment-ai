package wyckoff

import (
	"errors"
	"fmt"
	"math"
	"sort"
)

var ErrInsufficientData = errors.New("at least 30 valid candles are required")

type Config struct { RangeWindow, TrendWindow, VolumeWindow int }
func DefaultConfig() Config { return Config{RangeWindow:20, TrendWindow:10, VolumeWindow:20} }
type Analyzer struct { cfg Config }
func NewAnalyzer(cfg Config) *Analyzer {
	if cfg.RangeWindow < 5 { cfg.RangeWindow=20 }
	if cfg.TrendWindow < 3 { cfg.TrendWindow=10 }
	if cfg.VolumeWindow < 5 { cfg.VolumeWindow=20 }
	return &Analyzer{cfg:cfg}
}

// Analyze applies transparent heuristics to OHLCV. Confidence is a rule-strength
// score, not a statistically calibrated probability or investment recommendation.
func (a *Analyzer) Analyze(symbol string, candles []Candle) (Analysis,error) {
	if len(candles)<30 { return Analysis{},ErrInsufficientData }
	candles=append([]Candle(nil),candles...)
	sort.Slice(candles,func(i,j int)bool{return candles[i].Date.Before(candles[j].Date)})
	for i,c:=range candles {
		if c.High<c.Low || c.High<c.Open || c.High<c.Close || c.Low>c.Open || c.Low>c.Close || c.Close<=0 || c.Volume<0 {
			return Analysis{},fmt.Errorf("invalid candle at index %d",i)
		}
	}
	last:=candles[len(candles)-1]
	rw:=a.cfg.RangeWindow
	if rw>len(candles)-1 { rw=len(candles)-1 }
	prevWindow:=candles[len(candles)-rw-1:len(candles)-1]
	support,resistance:=math.MaxFloat64,-math.MaxFloat64
	for _,c:=range prevWindow { if c.Low<support {support=c.Low}; if c.High>resistance {resistance=c.High} }
	widthPct:=(resistance-support)/((resistance+support)/2)*100
	tr:=&TradingRange{Support:support,Resistance:resistance,WidthPct:widthPct}
	vw:=a.cfg.VolumeWindow
	if vw>len(candles)-1 {vw=len(candles)-1}
	avgVolume:=0.0
	for _,c:=range candles[len(candles)-1-vw:len(candles)-1] {avgVolume+=c.Volume}
	avgVolume/=float64(vw)
	result:=Analysis{Symbol:symbol,AsOf:last.Date,Structure:StructureUnclear,Phase:PhaseUnknown,Confidence:0.25,TradingRange:tr,Events:[]Event{},Evidence:[]string{},DataPoints:len(candles),DataMode:"heuristic"}
	addEvent:=func(kind string,confidence float64,evidence string){
		result.Events=append(result.Events,Event{Type:kind,Date:last.Date,Confidence:confidence,Evidence:evidence})
		result.Evidence=append(result.Evidence,evidence)
	}
	if last.Low<support && last.Close>support && last.Close>last.Open && last.Volume<avgVolume {
		addEvent("SPRING",0.70,"Price pierced prior support and closed back above it on below-average volume.")
	}
	if last.High>resistance && last.Close<resistance && last.Close<last.Open && last.Volume>avgVolume {
		addEvent("UTAD",0.68,"Price pierced prior resistance and closed back below it on above-average volume.")
	}
	// Use prior-window levels, not levels that include the current candle.
	if last.Close>resistance && last.Volume>avgVolume { addEvent("SOS",0.65,"Close broke above the prior range on above-average volume.") }
	if last.Close<support && last.Volume>avgVolume { addEvent("SOW",0.65,"Close broke below the prior range on above-average volume.") }

	trendWindow:=a.cfg.TrendWindow
	start:=candles[len(candles)-1-trendWindow].Close
	change:=(last.Close-start)/start
	switch {
	case widthPct>25 && change>0.08:
		result.Structure=StructureMarkup; result.Phase=PhaseE
		result.Evidence=append(result.Evidence,"Price advanced materially over the recent trend window.")
	case widthPct>25 && change < -0.08:
		result.Structure=StructureMarkdown; result.Phase=PhaseE
		result.Evidence=append(result.Evidence,"Price declined materially over the recent trend window.")
	case hasEvent(result.Events,"SPRING") || hasEvent(result.Events,"SOS"):
		result.Structure=StructureAccumulation
	case hasEvent(result.Events,"UTAD") || hasEvent(result.Events,"SOW"):
		result.Structure=StructureDistribution
	case widthPct<=25 && change < -0.08:
		result.Structure=StructureAccumulation
	case widthPct<=25 && change>0.08:
		result.Structure=StructureDistribution
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
		if len(result.Events)>0 {result.Confidence=0.65}
		if len(result.Events)>1 {result.Confidence=0.75}
	} else {
		result.Evidence=append(result.Evidence,"Available OHLCV does not support a reliable Wyckoff phase classification.")
	}
	result.Evidence=append(result.Evidence,fmt.Sprintf("Prior trading range: support %.2f, resistance %.2f (width %.1f%%).",support,resistance,widthPct))
	return result,nil
}
func hasEvent(events []Event,kind string)bool{for _,e:=range events{if e.Type==kind{return true}};return false}
