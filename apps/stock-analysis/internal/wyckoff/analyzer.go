package wyckoff

import (
	"errors"
	"fmt"
	"math"
	"sort"
)

var ErrInsufficientData = errors.New("at least 60 valid candles are required")

// EventLookback is the maximum history considered for the active Wyckoff cycle.
const EventLookback = 70

type Config struct { RangeWindow, TrendWindow, VolumeWindow int }
func DefaultConfig() Config { return Config{RangeWindow:40, TrendWindow:20, VolumeWindow:20} }
type Analyzer struct { cfg Config }
func NewAnalyzer(cfg Config) *Analyzer {
	if cfg.RangeWindow < 20 { cfg.RangeWindow = 40 }
	if cfg.TrendWindow < 5 { cfg.TrendWindow = 20 }
	if cfg.VolumeWindow < 5 { cfg.VolumeWindow = 20 }
	return &Analyzer{cfg:cfg}
}
type candleStats struct { spread, closeLocation, relativeVolume float64 }
func stats(c Candle, avgVol float64) candleStats {
	spread:=c.High-c.Low; clv:=0.5
	if spread>0 { clv=(c.Close-c.Low)/spread }
	rv:=0.0; if avgVol>0 { rv=c.Volume/avgVol }
	return candleStats{spread:spread,closeLocation:clv,relativeVolume:rv}
}
func avgVolume(cs []Candle) float64 {
	total:=0.0; for _,c:=range cs { total+=c.Volume }
	if len(cs)==0 { return 0 }; return total/float64(len(cs))
}
func bounds(cs []Candle)(float64,float64) {
	lo,hi:=math.MaxFloat64,-math.MaxFloat64
	for _,c:=range cs { if c.Low<lo {lo=c.Low}; if c.High>hi {hi=c.High} }
	return lo,hi
}
func hasEvent(es []Event, kind string) bool { for _,e:=range es {if e.Type==kind{return true}}; return false }
func eventIndex(es []Event, kind string) int { for i:=range es {if es[i].Type==kind{return i}}; return -1 }
func abs(v float64)float64 { return math.Abs(v) }

type rangeCandidate struct {
	start, end int
	support, resistance, widthPct, quality float64
	preTrend float64
	valid bool
}

// detectTradingRange looks for a reasonably stable recent range after a directional
// move. This is a heuristic candidate detector, not a discretionary chart label.
func (a *Analyzer) detectTradingRange(cs []Candle, trendChange float64) rangeCandidate {
	n:=len(cs)
	// Exclude the last five bars from range construction so a fresh breakout
	// does not erase the range it is breaking out of.
	end := n - 5
	if end < 1 { end = n }
	w:=a.cfg.RangeWindow
	if w>end {w=end}
	start:=end-w
	window:=cs[start:end]
	support,resistance:=bounds(window)
	mid:=(support+resistance)/2
	width:=(resistance-support)/math.Max(mid,0.00001)*100
	// Count how many candles overlap the inner 80% of the candidate range and
	// how often price reacts near either edge.
	innerLow:=support+(resistance-support)*0.1
	innerHigh:=resistance-(resistance-support)*0.1
	inside,lowTouches,highTouches:=0,0,0
	for _,c:=range window {
		if c.Close>=innerLow && c.Close<=innerHigh {inside++}
		if c.Low<=support+(resistance-support)*0.12 {lowTouches++}
		if c.High>=resistance-(resistance-support)*0.12 {highTouches++}
	}
	insideRatio:=float64(inside)/float64(len(window))
	touchScore:=math.Min(1,float64(lowTouches+highTouches)/6)
	// A range needs multiple reactions and price overlap; extremes alone are not enough.
	quality:=0.65*insideRatio+0.35*touchScore
	preStart:=start-a.cfg.TrendWindow
	preTrend:=0.0
	if preStart<0 {preStart=0}
	if start>preStart && cs[preStart].Close>0 {
		preTrend=(cs[start].Close-cs[preStart].Close)/cs[preStart].Close
	}
	valid:=width>=4 && width<=35 && insideRatio>=0.35 && lowTouches>=2 && highTouches>=2
	// If the range is actively expanding into a breakout, don't call it a range.
	_ = trendChange
	return rangeCandidate{start:start,end:end-1,support:support,resistance:resistance,widthPct:width,quality:quality,preTrend:preTrend,valid:valid}
}

// Analyze identifies a candidate range, then evaluates events within that
// recent context and checks their temporal order. Confidence is heuristic only.
func (a *Analyzer) Analyze(symbol string, candles []Candle) (Analysis,error) {
	if len(candles)<60 { return Analysis{},ErrInsufficientData }
	cs:=append([]Candle(nil),candles...)
	sort.Slice(cs,func(i,j int)bool{return cs[i].Date.Before(cs[j].Date)})
	for i,c:=range cs {
		if c.High<c.Low || c.High<c.Open || c.High<c.Close || c.Low>c.Open || c.Low>c.Close || c.Close<=0 || c.Volume<0 {
			return Analysis{},fmt.Errorf("invalid candle at index %d",i)
		}
	}
	n:=len(cs); last:=cs[n-1]; tw:=a.cfg.TrendWindow; vw:=a.cfg.VolumeWindow
	if tw>n/2 {tw=n/2}; if vw>n-1 {vw=n-1}
	priorClose:=cs[n-tw-1].Close
	trendChange:=0.0; if priorClose>0 {trendChange=(last.Close-priorClose)/priorClose}
	lookback:=EventLookback; if lookback>n {lookback=n}
	recentStart:=n-lookback
	recent:=cs[recentStart:]
	rc:=a.detectTradingRange(cs,trendChange)
	// Use the detected candidate's bounds if valid; otherwise expose recent
	// bounds as provisional levels, explicitly retaining UNKNOWN structure.
	support,resistance:=rc.support,rc.resistance
	if !rc.valid { support,resistance=bounds(recent) }
	widthPct:=(resistance-support)/math.Max((resistance+support)/2,0.00001)*100
	result:=Analysis{Symbol:symbol,AsOf:last.Date,Structure:StructureUnclear,Phase:PhaseUnknown,Confidence:0.2,
		TradingRange:&TradingRange{Support:support,Resistance:resistance,WidthPct:widthPct},
		Events:[]Event{},Evidence:[]string{},DataPoints:n,DataMode:"real"}
	add:=func(kind string,idx int,conf float64,why string) {
		result.Events=append(result.Events,Event{Type:kind,Date:cs[idx].Date,Confidence:conf,Evidence:why})
		result.Evidence=append(result.Evidence,fmt.Sprintf("%s (%s): %s",kind,cs[idx].Date.Format("2006-01-02"),why))
	}
	// Event candidates are evaluated only inside the recent cycle. Before a
	// valid range exists, do not label range-dependent Spring/UTAD/SOS/SOW.
	rangeStart:=rc.start
	if !rc.valid { rangeStart=recentStart }
	if rangeStart<1 {rangeStart=1}
	for i:=rangeStart;i<n;i++ {
		historyStart:=i-a.cfg.RangeWindow
		if historyStart<rangeStart {historyStart=rangeStart}
		if historyStart>=i {continue}
		lo,hi:=bounds(cs[historyStart:i])
		// Range-dependent events use the established candidate boundaries when
		// available, not a moving window that changes the definition every day.
		if rc.valid {
			lo,hi = support,resistance
		}
		c:=cs[i]
		volStart:=i-vw; if volStart<0 {volStart=0}
		st:=stats(c,avgVolume(cs[volStart:i]))
		trendStart:=i-tw; if trendStart<0 {trendStart=0}
		ctxChange:=0.0
		if cs[trendStart].Close>0 {ctxChange=(cs[i-1].Close-cs[trendStart].Close)/cs[trendStart].Close}
		rangeWidth:=math.Max(hi-lo,0.00001)
		// Preliminary support and climax candidates require an established decline,
		// relative volume expansion, and meaningful result-vs-effort evidence.
		if ctxChange < -0.08 && st.relativeVolume>=1.3 && st.spread<=rangeWidth*0.7 && st.closeLocation>=0.45 {
			add("PS",i,0.55,"Potential preliminary support after a decline: elevated volume with limited downside result.")
		}
		if ctxChange < -0.12 && st.relativeVolume>=1.5 && st.spread>=rangeWidth*0.18 && st.closeLocation>=0.55 {
			add("SC",i,0.62,"Selling-climax candidate after a decline: elevated volume and rejection from the lower part of the bar.")
		}
		if ctxChange > 0.12 && st.relativeVolume>=1.5 && st.spread>=rangeWidth*0.18 && st.closeLocation<=0.45 {
			add("BC",i,0.62,"Buying-climax candidate after an advance: elevated volume and rejection from the upper part of the bar.")
		}
		if rc.valid {
			if ctxChange < -0.05 && c.Low<support && c.Close>support && st.closeLocation>=0.55 && st.relativeVolume<1.35 {
				add("SPRING",i,0.72,"Price undercut the established range support and closed back inside on non-exceptional volume.")
			}
			if ctxChange > 0.05 && c.High>resistance && c.Close<resistance && st.closeLocation<=0.45 && st.relativeVolume>0.8 {
				add("UTAD",i,0.70,"Price exceeded established range resistance and closed back inside after an advance.")
			}
			if c.Close>resistance && st.closeLocation>=0.70 && st.relativeVolume>=1.2 {
				add("SOS",i,0.70,"Close broke above established range resistance with a strong close and expanding volume.")
			}
			if c.Close<support && st.closeLocation<=0.30 && st.relativeVolume>=1.2 {
				add("SOW",i,0.70,"Close broke below established range support with a weak close and expanding volume.")
			}
			if ctxChange < -0.05 && abs(c.Low-support)/math.Max(support,0.00001)<0.025 && st.relativeVolume<0.85 && st.spread<rangeWidth*0.65 {
				add("ST",i,0.58,"Secondary test approached established support with reduced volume and a relatively narrow spread.")
			}
		}
	}
	// Sequence-aware AR: a rally/reaction must occur after a climax and be
	// observable in subsequent closes, rather than assigning AR to the latest day.
	scIdx,bcIdx:=eventIndex(result.Events,"SC"),eventIndex(result.Events,"BC")
	if scIdx>=0 {
		scDate:=result.Events[scIdx].Date
		scBar:=0; for i:=range cs {if cs[i].Date.Equal(scDate){scBar=i;break}}
		end:=scBar+15; if end>=n {end=n-1}
		if end>scBar {
			lo,_:=bounds(cs[scBar:end+1])
			hi:=cs[scBar].High
			for _,c:=range cs[scBar:end+1] {if c.High>hi {hi=c.High}}
			if (hi-lo)/math.Max(cs[scBar].Close,0.00001)>=0.05 {
				arIdx:=scBar
				for i:=scBar+1;i<=end;i++ {if cs[i].High==hi {arIdx=i}}
				add("AR",arIdx,0.58,"Meaningful automatic rally followed the selling-climax candidate.")
			}
		}
	}
	if bcIdx>=0 {
		bcDate:=result.Events[bcIdx].Date
		bcBar:=0; for i:=range cs {if cs[i].Date.Equal(bcDate){bcBar=i;break}}
		end:=bcBar+15; if end>=n {end=n-1}
		if end>bcBar {
			lo,hi:=bounds(cs[bcBar:end+1])
			if (hi-lo)/math.Max(cs[bcBar].Close,0.00001)>=0.05 {
				arIdx:=bcBar
				for i:=bcBar+1;i<=end;i++ {if cs[i].Low==lo {arIdx=i}}
				add("AR",arIdx,0.58,"Meaningful automatic reaction followed the buying-climax candidate.")
			}
		}
	}
	// Sort events chronologically after event candidates have been appended.
	sort.SliceStable(result.Events,func(i,j int)bool{return result.Events[i].Date.Before(result.Events[j].Date)})
	// Validate the accumulation stopping sequence PS -> SC -> AR -> ST. These
	// flags are evidence, not hard requirements for every legitimate schematic.
	ps,sc,ar,stIdx:=eventIndex(result.Events,"PS"),eventIndex(result.Events,"SC"),eventIndex(result.Events,"AR"),eventIndex(result.Events,"ST")
	accumulationSequence:=ps>=0 && sc>ps && ar>sc && stIdx>ar
	if rc.valid && rc.preTrend < -0.08 && (accumulationSequence || (sc>=0 && ar>sc && stIdx>ar)) {
		result.Structure=StructureAccumulation
		switch {
		case hasEvent(result.Events,"SOS") || hasEvent(result.Events,"LPS"): result.Phase=PhaseD
		case hasEvent(result.Events,"SPRING"):
			if hasEvent(result.Events,"TEST") {result.Phase=PhaseC} else {result.Phase=PhaseC}
		case accumulationSequence: result.Phase=PhaseA
		default: result.Phase=PhaseB
		}
	} else if rc.valid && rc.preTrend > 0.08 && (hasEvent(result.Events,"BC") || hasEvent(result.Events,"UTAD") || hasEvent(result.Events,"SOW")) {
		result.Structure=StructureDistribution
		switch {
		case hasEvent(result.Events,"SOW") || hasEvent(result.Events,"LPSY"): result.Phase=PhaseD
		case hasEvent(result.Events,"UTAD"): result.Phase=PhaseC
		case hasEvent(result.Events,"BC"): result.Phase=PhaseA
		default: result.Phase=PhaseB
		}
	} else if trendChange>0.08 && last.Close>resistance*1.01 && hasEvent(result.Events,"SOS") {
		result.Structure=StructureMarkup; result.Phase=PhaseE
	} else if trendChange < -0.08 && last.Close<support*0.99 && hasEvent(result.Events,"SOW") {
		result.Structure=StructureMarkdown; result.Phase=PhaseE
	}
	// Only evaluate Test after a Spring within the same recent range; it must
	// happen shortly afterward, hold above the Spring low, and show lower supply.
	if rc.valid {
		for i:=range result.Events {
			if result.Events[i].Type!="SPRING" {continue}
			springIdx:=0; for j:=range cs {if cs[j].Date.Equal(result.Events[i].Date){springIdx=j;break}}
			end:=springIdx+15; if end>=n {end=n-1}
			for j:=springIdx+2;j<=end;j++ {
				volStart:=j-vw; if volStart<0 {volStart=0}
				ts:=stats(cs[j],avgVolume(cs[volStart:j]))
				if cs[j].Low>=cs[springIdx].Low && ts.relativeVolume<0.8 && ts.closeLocation>=0.55 && cs[j].Low<=support+(resistance-support)*0.2 {
					add("TEST",j,0.68,"Retest followed the Spring within 15 sessions, held above the Spring low, and showed reduced supply.")
					break
				}
			}
			break
		}
	}
	// LPS only follows an SOS in the same active range and within 15 sessions.
	if rc.valid && hasEvent(result.Events,"SOS") {
		for i:=range result.Events {
			if result.Events[i].Type!="SOS" {continue}
			sosIdx:=0; for j:=range cs {if cs[j].Date.Equal(result.Events[i].Date){sosIdx=j;break}}
			end:=sosIdx+15; if end>=n {end=n-1}
			for j:=sosIdx+1;j<=end;j++ {
				volStart:=j-vw; if volStart<0 {volStart=0}
				ls:=stats(cs[j],avgVolume(cs[volStart:j]))
				if cs[j].Low>=resistance*0.98 && cs[j].Close>=resistance && ls.relativeVolume<0.9 {
					add("LPS",j,0.62,"Pullback after SOS held former range resistance as support with reduced volume.")
					if result.Structure==StructureAccumulation {result.Phase=PhaseD}
					break
				}
			}
			break
		}
	}
	sort.SliceStable(result.Events,func(i,j int)bool{return result.Events[i].Date.Before(result.Events[j].Date)})
	// Evidence for sequence quality and candidate range.
	result.Evidence=append(result.Evidence,fmt.Sprintf("Recent range candidate: valid=%t, support=%.2f, resistance=%.2f, width=%.1f%%, quality=%.2f, pre-range trend=%.1f%%.",rc.valid,support,resistance,widthPct,rc.quality,rc.preTrend*100))
	result.Evidence=append(result.Evidence,fmt.Sprintf("Recent trend change %.1f%%; PS->SC->AR->ST sequence=%t.",trendChange*100,accumulationSequence))
	if result.Phase!=PhaseUnknown {
		result.Confidence=0.42
		if rc.valid {result.Confidence+=0.10}
		if accumulationSequence {result.Confidence+=0.12}
		if hasEvent(result.Events,"TEST") {result.Confidence+=0.08}
		if hasEvent(result.Events,"SOS") {result.Confidence+=0.08}
		if result.Confidence>0.82 {result.Confidence=0.82}
	}
	if result.Phase==PhaseUnknown {result.Evidence=append(result.Evidence,"Phase remains UNKNOWN: no sufficiently validated range and event sequence; wait for confirmation rather than anticipate.")}
	return result,nil
}
func maxInt(a,b int)int {if a>b{return a};return b}
