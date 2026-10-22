package wyckoff

import (
	"errors"
	"fmt"
	"math"
	"sort"
)

var ErrInsufficientData = errors.New("at least 60 valid candles are required")

type Config struct { RangeWindow, TrendWindow, VolumeWindow int }
func DefaultConfig() Config { return Config{RangeWindow:40, TrendWindow:20, VolumeWindow:20} }
type Analyzer struct { cfg Config }
func NewAnalyzer(cfg Config) *Analyzer {
	if cfg.RangeWindow<20 { cfg.RangeWindow=40 }; if cfg.TrendWindow<5 {cfg.TrendWindow=20}; if cfg.VolumeWindow<5 {cfg.VolumeWindow=20}
	return &Analyzer{cfg:cfg}
}
type candleStats struct { spread, closeLocation, relativeVolume, change float64 }
func stats(c Candle, avgVol float64) candleStats {
	spread:=c.High-c.Low; clv:=0.5
	if spread>0 {clv=(c.Close-c.Low)/spread}
	rv:=0.0; if avgVol>0 {rv=c.Volume/avgVol}
	return candleStats{spread:spread,closeLocation:clv,relativeVolume:rv,change:(c.Close-c.Open)}
}
func avgVolume(cs []Candle) float64 { total:=0.0; for _,c:=range cs {total+=c.Volume}; if len(cs)==0{return 0}; return total/float64(len(cs)) }
func bounds(cs []Candle)(float64,float64){lo,hi:=math.MaxFloat64,-math.MaxFloat64;for _,c:=range cs{if c.Low<lo{lo=c.Low};if c.High>hi{hi=c.High}};return lo,hi}
func hasEvent(es []Event, kind string) bool {for _,e:=range es{if e.Type==kind{return true}};return false}

// Analyze evaluates price/volume context and event sequences. Confidence is a
// heuristic evidence score, not a calibrated probability or investment advice.
func (a *Analyzer) Analyze(symbol string, candles []Candle) (Analysis,error) {
	if len(candles)<60{return Analysis{},ErrInsufficientData}
	cs:=append([]Candle(nil),candles...)
	sort.Slice(cs,func(i,j int)bool{return cs[i].Date.Before(cs[j].Date)})
	for i,c:=range cs {if c.High<c.Low||c.High<c.Open||c.High<c.Close||c.Low>c.Open||c.Low>c.Close||c.Close<=0||c.Volume<0{return Analysis{},fmt.Errorf("invalid candle at index %d",i)}}
	last:=cs[len(cs)-1]; n:=len(cs); tw:=a.cfg.TrendWindow; rw:=a.cfg.RangeWindow
	if tw>n/2{tw=n/2}; if rw>n-tw-1{rw=n-tw-1}
	priorRange:=cs[n-rw-1:n-1]; support,resistance:=bounds(priorRange)
	widthPct:=(resistance-support)/math.Max(0.00001,(resistance+support)/2)*100
	vw:=a.cfg.VolumeWindow; if vw>n-1{vw=n-1}; avgVol:=avgVolume(cs[n-vw-1:n-1])
	lastStats:=stats(last,avgVol)
	priorClose:=cs[n-tw-1].Close; trendChange:=(last.Close-priorClose)/priorClose
	preTrendStart:=cs[n-2*tw-1].Close; preTrendChange:=(cs[n-tw-1].Close-preTrendStart)/preTrendStart
	// Long enough to identify context: accumulation usually follows markdown;
	// distribution usually follows markup. Range shape alone is not sufficient.
	result:=Analysis{Symbol:symbol,AsOf:last.Date,Structure:StructureUnclear,Phase:PhaseUnknown,Confidence:0.20,TradingRange:&TradingRange{Support:support,Resistance:resistance,WidthPct:widthPct},Events:[]Event{},Evidence:[]string{},DataPoints:n,DataMode:"real"}
	add:=func(kind string, idx int, conf float64, why string){result.Events=append(result.Events,Event{Type:kind,Date:cs[idx].Date,Confidence:conf,Evidence:why});result.Evidence=append(result.Evidence,fmt.Sprintf("%s (%s): %s",kind,cs[idx].Date.Format("2006-01-02"),why))}
	// Detect candidate events over the last 12 candles, with event levels formed
	// by earlier candles to avoid look-ahead contamination.
	start:=n-12; if start<tw{start=tw}
	for i:=start;i<n;i++ {
		windowStart:=i-rw; if windowStart<0{windowStart=0}; if windowStart>=i{continue}
		lo,hi:=bounds(cs[windowStart:i]); c:=cs[i]; st:=stats(c,avgVolume(cs[maxInt(0,i-vw):i]))
		prevTrendStart:=i-tw; if prevTrendStart<0{prevTrendStart=0}
		ctxChange:=(cs[i-1].Close-cs[prevTrendStart].Close)/cs[prevTrendStart].Close
		if ctxChange < -0.08 && c.Low<lo && c.Close>lo && st.closeLocation>=0.55 && st.relativeVolume<1.25 {
			add("SPRING",i,0.72,"After a declining context, price undercut prior support and recovered into the range without exceptional selling volume.")
		}
		if ctxChange > 0.08 && c.High>hi && c.Close<hi && st.closeLocation<=0.45 && st.relativeVolume>0.8 {
			add("UTAD",i,0.70,"After a rising context, price exceeded prior resistance and closed back inside the range.")
		}
		if c.Close>hi && st.closeLocation>=0.70 && st.relativeVolume>=1.2 {
			add("SOS",i,0.70,"Close broke above prior resistance with a strong close and expanding volume.")
		}
		if c.Close<lo && st.closeLocation<=0.30 && st.relativeVolume>=1.2 {
			add("SOW",i,0.70,"Close broke below prior support with a weak close and expanding volume.")
		}
		if ctxChange < -0.08 && math.Abs(c.Low-lo)/math.Max(lo,0.00001) < 0.025 && st.relativeVolume < 0.85 && st.spread < (hi-lo)*0.65 {
			add("ST",i,0.58,"A secondary test approached support with reduced volume and a relatively narrow spread.")
		}
		if ctxChange > 0.08 && c.High>hi && c.Close<hi && st.closeLocation<0.45 && st.relativeVolume>1.0 {
			add("UT",i,0.60,"Price briefly exceeded resistance and was rejected back into the range.")
		}
		if ctxChange < -0.08 && st.relativeVolume >= 1.5 && st.spread < (hi-lo)*0.5 && c.Close >= c.Open {
			add("PS",i,0.55,"Potential preliminary support: elevated volume with limited downside result.")
		}
		// Selling/buying climax candidates: wide spread, elevated volume and
		// rejection/reversal close after an established directional move.
		if ctxChange < -0.12 && st.relativeVolume>=1.5 && st.spread>0 && st.closeLocation>=0.55 { add("SC",i,0.62,"High-volume wide-spread decline was rejected and closed in the upper part of the bar.") }
		if ctxChange > 0.12 && st.relativeVolume>=1.5 && st.spread>0 && st.closeLocation<=0.45 { add("BC",i,0.62,"High-volume wide-spread advance was rejected and closed in the lower part of the bar.") }
	}
	// Automatic Rally / Reaction: meaningful countertrend response following a
	// climax, only report if the move is large enough to be observable.
	if hasEvent(result.Events,"SC") && trendChange>0.05 { add("AR",n-1,0.58,"A meaningful rally followed a selling-climax candidate.") }
	if hasEvent(result.Events,"BC") && trendChange < -0.05 { add("AR",n-1,0.58,"A meaningful reaction followed a buying-climax candidate.") }
	// Last Point of Support / Supply candidates after confirmed range breaks.
	for i:=0;i<n-1;i++ {
		if cs[i].Close>resistance && stats(cs[i],avgVol).relativeVolume>=1.2 {
			for j:=i+1;j<n;j++ { st:=stats(cs[j],avgVolume(cs[maxInt(0,j-vw):j])); if cs[j].Low>=resistance*0.98 && cs[j].Close>=resistance && st.relativeVolume<0.9 { add("LPS",j,0.62,"Pullback held the former resistance as support with reduced volume after a breakout."); break } }
			break
		}
		if cs[i].Close<support && stats(cs[i],avgVol).relativeVolume>=1.2 {
			for j:=i+1;j<n;j++ { st:=stats(cs[j],avgVolume(cs[maxInt(0,j-vw):j])); if cs[j].High<=support*1.02 && cs[j].Close<=support && st.relativeVolume<0.9 { add("LPSY",j,0.62,"Rally failed near former support/resistance with reduced demand after a breakdown."); break } }
			break
		}
	}
	// Test: after a Spring, require low supply and no close below the spring low.
	for i:=len(result.Events)-1;i>=0;i-- {
		e:=result.Events[i]; if e.Type!="SPRING"{continue}
		springIdx:=0; for j:=range cs{if cs[j].Date.Equal(e.Date){springIdx=j;break}}
		for j:=springIdx+1;j<n;j++ {st:=stats(cs[j],avgVolume(cs[maxInt(0,j-vw):j])); if cs[j].Low>=cs[springIdx].Low && st.relativeVolume<0.8 && st.closeLocation>=0.55 {add("TEST",j,0.68,"Retest held above the Spring low with below-average volume and a firm close.");break}}
		break
	}
	// Structural interpretation uses context + sequence rather than one event.
	switch {
	case trendChange>0.08 && widthPct>15 && (hasEvent(result.Events,"SOS") || (trendChange>0.15 && last.Close>resistance)):
		result.Structure=StructureMarkup; result.Phase=PhaseE
	case trendChange < -0.08 && widthPct>15 && (hasEvent(result.Events,"SOW") || (trendChange < -0.15 && last.Close<support)):
		result.Structure=StructureMarkdown; result.Phase=PhaseE
	case hasEvent(result.Events,"SPRING") || hasEvent(result.Events,"SC"):
		result.Structure=StructureAccumulation
	case hasEvent(result.Events,"UTAD") || hasEvent(result.Events,"BC"):
		result.Structure=StructureDistribution
	case widthPct<=30 && preTrendChange < -0.08:
		result.Structure=StructureAccumulation
	case widthPct<=30 && preTrendChange > 0.08:
		result.Structure=StructureDistribution
	}
	if result.Phase==PhaseUnknown {
		switch {
		case hasEvent(result.Events,"SPRING") || hasEvent(result.Events,"UTAD"): result.Phase=PhaseC
		case hasEvent(result.Events,"SOS") || hasEvent(result.Events,"SOW") || hasEvent(result.Events,"TEST"): result.Phase=PhaseD
		case result.Structure==StructureAccumulation || result.Structure==StructureDistribution:
			if hasEvent(result.Events,"SC")||hasEvent(result.Events,"BC"){result.Phase=PhaseA}else{result.Phase=PhaseB}
		}
	}
	if result.Structure==StructureAccumulation {
		if hasEvent(result.Events,"SOS") || hasEvent(result.Events,"TEST") {result.Phase=PhaseD}
		if hasEvent(result.Events,"SPRING") && !hasEvent(result.Events,"SOS") {result.Phase=PhaseC}
	}
	if result.Structure==StructureDistribution {
		if hasEvent(result.Events,"SOW") {result.Phase=PhaseD}
		if hasEvent(result.Events,"UTAD") && !hasEvent(result.Events,"SOW") {result.Phase=PhaseC}
	}
	if result.Phase!=PhaseUnknown {result.Confidence=0.42;if len(result.Events)>0{result.Confidence=0.55};if len(result.Events)>=2{result.Confidence=0.68};if hasEvent(result.Events,"TEST")&&(hasEvent(result.Events,"SPRING")||hasEvent(result.Events,"SOS")){result.Confidence=0.76}}
	result.Evidence=append(result.Evidence,fmt.Sprintf("Prior range support %.2f / resistance %.2f; width %.1f%%. Recent trend change %.1f%%.",support,resistance,widthPct,trendChange*100))
	if result.Phase==PhaseUnknown {result.Evidence=append(result.Evidence,"Phase remains UNKNOWN: evidence is insufficient or conflicting; wait for confirmation rather than anticipate.")}
	return result,nil
}
func maxInt(a,b int)int{if a>b{return a};return b}
