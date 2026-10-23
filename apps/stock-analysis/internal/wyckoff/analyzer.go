package wyckoff

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
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
	if w < EventLookback {
		w = EventLookback
	}
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
func (a *Analyzer) Analyze(symbol string, candles []Candle) (Analysis, error) {
	if len(candles) < 60 {
		return Analysis{}, ErrInsufficientData
	}
	cs := append([]Candle(nil), candles...)
	sort.Slice(cs, func(i, j int) bool { return cs[i].Date.Before(cs[j].Date) })
	for i, c := range cs {
		if c.High < c.Low || c.High < c.Open || c.High < c.Close || c.Low > c.Open || c.Low > c.Close || c.Close <= 0 || c.Volume < 0 {
			return Analysis{}, fmt.Errorf("invalid candle at index %d", i)
		}
	}
	n := len(cs)
	last := cs[n-1]
	tw := a.cfg.TrendWindow
	vw := a.cfg.VolumeWindow
	if tw > n/2 {
		tw = n / 2
	}
	if vw > n-1 {
		vw = n - 1
	}
	priorClose := cs[n-tw-1].Close
	trendChange := 0.0
	if priorClose > 0 {
		trendChange = (last.Close - priorClose) / priorClose
	}
	lookback := EventLookback
	if lookback > n {
		lookback = n
	}
	recentStart := n - lookback
	recent := cs[recentStart:]
	rc := a.detectTradingRange(cs, trendChange)
	// Use the detected candidate's bounds if valid; otherwise expose recent
	// bounds as provisional levels, explicitly retaining UNKNOWN structure.
	support, resistance := rc.support, rc.resistance
	useRangeLevels := rc.widthPct >= 4 && rc.widthPct <= 35
	if !useRangeLevels {
		support, resistance = bounds(recent)
	}
	widthPct := (resistance - support) / math.Max((resistance+support)/2, 0.00001) * 100
	result := Analysis{
		Symbol:       symbol,
		AsOf:         last.Date,
		Structure:    StructureUnclear,
		Phase:        PhaseUnknown,
		Confidence:   0.2,
		TradingRange: &TradingRange{Support: support, Resistance: resistance, WidthPct: widthPct},
		Events:       []Event{},
		Evidence:     []string{},
		DataPoints:   n,
		DataMode:     "real",
	}
	add := func(kind string, idx int, conf float64, why string) {
		result.Events = append(result.Events, Event{Type: kind, Date: cs[idx].Date, Confidence: conf, Evidence: why})
		result.Evidence = append(result.Evidence, fmt.Sprintf("%s (%s): %s", kind, cs[idx].Date.Format("2006-01-02"), why))
	}
	// Evaluate the entire active-cycle lookback, not only the shorter range-construction window.
	// This allows a climax near the start of the range to be linked to its later AR/ST events.
	rangeStart := recentStart
	if rangeStart < 1 {
		rangeStart = 1
	}
	lastSTIndex := -1000
	for i := rangeStart; i < n; i++ {
		historyStart := i - a.cfg.RangeWindow
		if historyStart < rangeStart {
			historyStart = rangeStart
		}
		if historyStart >= i {
			continue
		}
		lo, hi := bounds(cs[historyStart:i])
		// Range-dependent events use the established candidate boundaries when
		// available, not a moving window that changes the definition every day.
		if useRangeLevels {
			lo, hi = support, resistance
		}
		c := cs[i]
		volStart := i - vw
		if volStart < 0 {
			volStart = 0
		}
		st := stats(c, avgVolume(cs[volStart:i]))
		trendStart := i - tw
		if trendStart < 0 {
			trendStart = 0
		}
		ctxChange := 0.0
		if cs[trendStart].Close > 0 {
			ctxChange = (cs[i-1].Close - cs[trendStart].Close) / cs[trendStart].Close
		}
		rangeWidth := math.Max(hi-lo, 0.00001)
		// Preliminary support and climax candidates require an established decline,
		// relative volume expansion, and meaningful result-vs-effort evidence.
		if ctxChange < -0.08 && st.relativeVolume >= 1.3 && st.spread <= rangeWidth*0.7 && st.closeLocation >= 0.45 {
			add("PS_CANDIDATE", i, 0.55, "Potential preliminary-support candidate.")
		}
		if ctxChange < -0.12 && st.relativeVolume >= 1.5 && st.spread >= rangeWidth*0.18 && st.closeLocation >= 0.55 {
			add("SC_CANDIDATE", i, 0.62, "Selling-climax candidate after a decline.")
		}
		if ctxChange > 0.12 && st.relativeVolume >= 1.5 && st.spread >= rangeWidth*0.18 && st.closeLocation <= 0.45 {
			add("BC_CANDIDATE", i, 0.62, "Buying-climax candidate after an advance.")
		}
		if useRangeLevels {
			if rc.valid && i >= rc.start && i <= rc.end && ctxChange < -0.05 && c.Low < support && c.Close > support && st.closeLocation >= 0.55 && st.relativeVolume < 1.35 {
				add("SPRING", i, 0.72, "After the trading range formed, price undercut support and closed back inside on non-exceptional volume.")
			}
			if ctxChange > 0.05 && c.High > resistance && c.Close < resistance && st.closeLocation <= 0.45 && st.relativeVolume > 0.8 {
				add("UTAD", i, 0.70, "Price exceeded established range resistance and closed back inside after an advance.")
			}
			if rc.valid && i > rc.end && c.Close > resistance && st.closeLocation >= 0.70 && st.relativeVolume >= 1.2 {
				add("SOS_CANDIDATE", i, 0.70, "Post-range close broke above established resistance with a strong close and expanding volume.")
			}
			// SOW is a post-range weakness signal: do not label a breakdown
			// that occurs before the candidate trading range has completed.
			// detectTradingRange intentionally ends five bars before the latest
			// candle, leaving those bars available to confirm a range breakout.
			if rc.valid && i > rc.end && c.Close < support && st.closeLocation <= 0.30 && st.relativeVolume >= 1.2 {
				add("SOW", i, 0.70, "After the established trading range, price broke below support with a weak close and expanding volume.")
			}
			if abs(c.Low-support)/math.Max(support, 0.00001) < 0.025 && st.relativeVolume < 0.85 && st.spread < rangeWidth*0.65 && i-lastSTIndex >= 3 {
				add("ST", i, 0.58, "Secondary test approached established support with reduced volume and a relatively narrow spread.")
				lastSTIndex = i
			}
		}
	}
	// Consolidate noisy per-bar candidates into one event per stopping action.
	// Keep only the strongest PS candidate before the selected SC, and select one
	// SC from a short climax cluster using the deepest low, breaking ties by volume.
	var psCandidates, scCandidates, bcCandidates, sosCandidates []Event
	kept := result.Events[:0]
	for _, e := range result.Events {
		switch e.Type {
		case "PS_CANDIDATE":
			psCandidates = append(psCandidates, e)
		case "SC_CANDIDATE":
			scCandidates = append(scCandidates, e)
		case "BC_CANDIDATE":
			bcCandidates = append(bcCandidates, e)
		case "SOS_CANDIDATE":
			sosCandidates = append(sosCandidates, e)
		default:
			kept = append(kept, e)
		}
	}
	result.Events = kept
	// Candidate labels are internal diagnostics, not final Wyckoff events.
	cleanEvidence := result.Evidence[:0]
	for _, evidence := range result.Evidence {
		if strings.HasPrefix(evidence, "PS_CANDIDATE ") ||
			strings.HasPrefix(evidence, "SC_CANDIDATE ") ||
			strings.HasPrefix(evidence, "BC_CANDIDATE ") ||
			strings.HasPrefix(evidence, "SOS_CANDIDATE ") {
			continue
		}
		cleanEvidence = append(cleanEvidence, evidence)
	}
	result.Evidence = cleanEvidence
	findBar := func(date time.Time) int {
		for i := range cs {
			if cs[i].Date.Equal(date) {
				return i
			}
		}
		return -1
	}
	if len(scCandidates) > 0 {
		// Consolidate nearby SC candidates into the first stopping-action cluster.
		// Keep the cluster open for up to five trading sessions so a later,
		// stronger decline (rather than the first detected down-spike) can win.
		first := findBar(scCandidates[0].Date)
		if first >= 0 {
			// A strict SC candidate predicate can miss the final climax bar when
			// its contextual trend metric changes on the day. Starting from the
			// first detected candidate, inspect the next five trading sessions for
			// a later bearish sell-off and prefer the latest qualifying sell-off.
			lastSearch := first + 5
			if lastSearch >= n {
				lastSearch = n - 1
			}
			best := first
			for i := first + 1; i <= lastSearch; i++ {
				candle := cs[i]
				avgStart := i - vw
				if avgStart < 0 {
					avgStart = 0
				}
				st := stats(candle, avgVolume(cs[avgStart:i]))
				dayChange := 0.0
				if cs[i-1].Close > 0 {
					dayChange = (candle.Close - cs[i-1].Close) / cs[i-1].Close
				}
				bearishBody := math.Max(0, candle.Open-candle.Close) / math.Max(candle.High-candle.Low, 0.00001)
				// Require a real bearish session with meaningful selling effort.
				// This intentionally lets a later climax bar qualify even if it
				// did not pass the earlier multi-day ctxChange predicate.
				if dayChange < -0.015 && bearishBody >= 0.20 && st.relativeVolume >= 0.8 {
					best = i
				}
			}
			add(
				"SC",
				best,
				0.68,
				"Selected the latest meaningful bearish sell-off within five sessions after the first climax candidate; later down days are checked independently of the multi-day trend predicate.",
			)
			// Emit at most one PS and only if it precedes the selected SC.
			bestPS := -1
			for i, e := range psCandidates {
				idx := findBar(e.Date)
				if idx >= 0 && idx < best && (bestPS < 0 || idx > findBar(psCandidates[bestPS].Date)) {
					bestPS = i
				}
			}
			if bestPS >= 0 {
				add("PS", findBar(psCandidates[bestPS].Date), 0.55, "Preliminary support before the selected selling climax.")
			}
		}
	}
	// Merge consecutive post-range SOS bars into one breakout event.
	// Keep the first bar in each contiguous cluster.
	if len(sosCandidates) > 0 {
		// Consolidate adjacent breakout bars into one SOS and date it at the
		// last qualifying breakout bar. For example, if price breaks resistance,
		// rises for three consecutive sessions, then pulls back to range resistance,
		// the third up/breakout session is the SOS rather than the first session.
		clusterStart := findBar(sosCandidates[0].Date)
		clusterLast := clusterStart
		for _, candidate := range sosCandidates[1:] {
			idx := findBar(candidate.Date)
			if idx >= 0 && idx-clusterLast <= 1 {
				clusterLast = idx
				continue
			}
			add("SOS", clusterLast, 0.70, "Post-range sign of strength; consecutive breakout bars are consolidated and dated at the final qualifying breakout bar.")
			clusterStart, clusterLast = idx, idx
		}
		add("SOS", clusterLast, 0.70, "Post-range sign of strength; consecutive breakout bars are consolidated and dated at the final qualifying breakout bar.")
	}
	if len(bcCandidates) > 0 {
		best := 0
		for i := 1; i < len(bcCandidates); i++ {
			if bcCandidates[i].Confidence > bcCandidates[best].Confidence {
				best = i
			}
		}
		add("BC", findBar(bcCandidates[best].Date), 0.62, "Selected buying-climax candidate from the recent advance.")
	}
	// Sequence-aware AR: a rally/reaction must occur after a climax and be
	// observable in subsequent closes, rather than assigning AR to the latest day.
	scIdx, bcIdx := eventIndex(result.Events, "SC"), eventIndex(result.Events, "BC")
	if scIdx >= 0 {
		scDate := result.Events[scIdx].Date
		scBar := 0
		for i := range cs {
			if cs[i].Date.Equal(scDate) {
				scBar = i
				break
			}
		}
		end := scBar + 15
		if end >= n {
			end = n - 1
		}
		if end > scBar {
			lo, _ := bounds(cs[scBar : end+1])
			hi := cs[scBar].High
			for _, c := range cs[scBar : end+1] {
				if c.High > hi {
					hi = c.High
				}
			}
			if (hi-lo)/math.Max(cs[scBar].Close, 0.00001) >= 0.05 {
				arIdx := scBar
				for i := scBar + 1; i <= end; i++ {
					if cs[i].High == hi {
						arIdx = i
					}
				}
				add("AR", arIdx, 0.58, "Meaningful automatic rally followed the selling-climax candidate.")
			}
		}
	}
	if bcIdx >= 0 {
		bcDate := result.Events[bcIdx].Date
		bcBar := 0
		for i := range cs {
			if cs[i].Date.Equal(bcDate) {
				bcBar = i
				break
			}
		}
		end := bcBar + 15
		if end >= n {
			end = n - 1
		}
		if end > bcBar {
			lo, hi := bounds(cs[bcBar : end+1])
			if (hi-lo)/math.Max(cs[bcBar].Close, 0.00001) >= 0.05 {
				arIdx := bcBar
				for i := bcBar + 1; i <= end; i++ {
					if cs[i].Low == lo {
						arIdx = i
					}
				}
				add("AR", arIdx, 0.58, "Meaningful automatic reaction followed the buying-climax candidate.")
			}
		}
	}
	// Keep ST labels only when they occur after the SC -> AR stopping action.
	// A low-volume touch of support before the automatic rally is not a secondary test.
	scDateForST, arDateForST := time.Time{}, time.Time{}
	for _, e := range result.Events {
		if e.Type == "SC" && (scDateForST.IsZero() || e.Date.Before(scDateForST)) {
			scDateForST = e.Date
		}
		if e.Type == "AR" && (arDateForST.IsZero() || e.Date.Before(arDateForST)) {
			arDateForST = e.Date
		}
	}
	filteredEvents := result.Events[:0]
	for _, e := range result.Events {
		if e.Type == "ST" && (scDateForST.IsZero() || arDateForST.IsZero() || !e.Date.After(arDateForST)) {
			continue
		}
		filteredEvents = append(filteredEvents, e)
	}
	result.Events = filteredEvents
	// Rebuild event evidence from the final event list so filtered-out labels
	// (for example an early ST before AR) do not remain in the API response.
	result.Evidence = result.Evidence[:0]
	for _, e := range result.Events {
		result.Evidence = append(result.Evidence, fmt.Sprintf("%s (%s): %s", e.Type, e.Date.Format("2006-01-02"), e.Evidence))
	}
	// Sort events chronologically after event candidates have been appended.
	sort.SliceStable(result.Events, func(i, j int) bool { return result.Events[i].Date.Before(result.Events[j].Date) })
	// Validate the accumulation stopping sequence PS -> SC -> AR -> ST. These
	// flags are evidence, not hard requirements for every legitimate schematic.
	ps, sc, ar, stIdx := eventIndex(result.Events, "PS"), eventIndex(result.Events, "SC"), eventIndex(result.Events, "AR"), eventIndex(result.Events, "ST")
	// Sequence events must occur on strictly later bars; same-day PS and SC
	// candidates do not establish the stopping sequence.
	accumulationSequence := false
	if ps >= 0 && sc >= 0 && ar >= 0 && stIdx >= 0 {
		psDate, scDate := result.Events[ps].Date, result.Events[sc].Date
		arDate, stDate := result.Events[ar].Date, result.Events[stIdx].Date
		accumulationSequence = psDate.Before(scDate) && scDate.Before(arDate) && arDate.Before(stDate)
	}
	if useRangeLevels && rc.preTrend < -0.08 && (accumulationSequence || (sc >= 0 && ar > sc && stIdx > ar)) {
		result.Structure = StructureAccumulation
		switch {
		case hasEvent(result.Events, "SOS") || hasEvent(result.Events, "LPS"):
			result.Phase = PhaseD
		case hasEvent(result.Events, "SPRING"):
			if hasEvent(result.Events, "TEST") {
				result.Phase = PhaseC
			} else {
				result.Phase = PhaseC
			}
		case accumulationSequence:
			result.Phase = PhaseA
		default:
			result.Phase = PhaseB
		}
	} else if useRangeLevels && rc.preTrend > 0.08 && (hasEvent(result.Events, "BC") || hasEvent(result.Events, "UTAD") || hasEvent(result.Events, "SOW")) {
		result.Structure = StructureDistribution
		switch {
		case hasEvent(result.Events, "SOW") || hasEvent(result.Events, "LPSY"):
			result.Phase = PhaseD
		case hasEvent(result.Events, "UTAD"):
			result.Phase = PhaseC
		case hasEvent(result.Events, "BC"):
			result.Phase = PhaseA
		default:
			result.Phase = PhaseB
		}
	} else if trendChange > 0.08 && last.Close > resistance*1.01 && hasEvent(result.Events, "SOS") {
		result.Structure = StructureMarkup
		result.Phase = PhaseE
	} else if trendChange < -0.08 && last.Close < support*0.99 && hasEvent(result.Events, "SOW") {
		result.Structure = StructureMarkdown
		result.Phase = PhaseE
	}
	// Only evaluate Test after a Spring within the same recent range; it must
	// happen shortly afterward, hold above the Spring low, and show lower supply.
	if useRangeLevels {
		for i := range result.Events {
			if result.Events[i].Type != "SPRING" {
				continue
			}
			springIdx := 0
			for j := range cs {
				if cs[j].Date.Equal(result.Events[i].Date) {
					springIdx = j
					break
				}
			}
			end := springIdx + 15
			if end >= n {
				end = n - 1
			}
			for j := springIdx + 2; j <= end; j++ {
				volStart := j - vw
				if volStart < 0 {
					volStart = 0
				}
				ts := stats(cs[j], avgVolume(cs[volStart:j]))
				if cs[j].Low >= cs[springIdx].Low && ts.relativeVolume < 0.8 && ts.closeLocation >= 0.55 && cs[j].Low <= support+(resistance-support)*0.2 {
					add("TEST", j, 0.68, "Retest followed the Spring within 15 sessions, held above the Spring low, and showed reduced supply.")
					break
				}
			}
			break
		}
	}
	// LPS only follows an SOS in the same active range and within 15 sessions.
	if useRangeLevels && hasEvent(result.Events, "SOS") {
		for i := range result.Events {
			if result.Events[i].Type != "SOS" {
				continue
			}
			sosIdx := 0
			for j := range cs {
				if cs[j].Date.Equal(result.Events[i].Date) {
					sosIdx = j
					break
				}
			}
			end := sosIdx + 15
			if end >= n {
				end = n - 1
			}
			for j := sosIdx + 1; j <= end; j++ {
				volStart := j - vw
				if volStart < 0 {
					volStart = 0
				}
				ls := stats(cs[j], avgVolume(cs[volStart:j]))
				if cs[j].Low >= resistance*0.98 && cs[j].Close >= resistance && ls.relativeVolume < 0.9 {
					add("LPS", j, 0.62, "Pullback after SOS held former range resistance as support with reduced volume.")
					if result.Structure == StructureAccumulation {
						result.Phase = PhaseD
					}
					break
				}
			}
			break
		}
	}
	sort.SliceStable(result.Events, func(i, j int) bool { return result.Events[i].Date.Before(result.Events[j].Date) })
	// Keep event timeline chronological, but render evidence in canonical Wyckoff order.
	result.Evidence = result.Evidence[:0]
	for _, e := range orderEventsForEvidence(result.Events) {
		result.Evidence = append(result.Evidence, fmt.Sprintf("%s (%s): %s", e.Type, e.Date.Format("2006-01-02"), e.Evidence))
	}
	// Evidence for sequence quality and candidate range.
	result.Evidence = append(result.Evidence, fmt.Sprintf(
		"Recent range candidate: valid=%t, support=%.2f, resistance=%.2f, width=%.1f%%, quality=%.2f, pre-range trend=%.1f%%.",
		rc.valid, support, resistance, widthPct, rc.quality, rc.preTrend*100,
	))
	result.Evidence = append(result.Evidence, fmt.Sprintf("Recent trend change %.1f%%; PS->SC->AR->ST sequence=%t.", trendChange*100, accumulationSequence))
	if result.Phase != PhaseUnknown {
		result.Confidence = 0.42
		if rc.valid {
			result.Confidence += 0.10
		}
		if accumulationSequence {
			result.Confidence += 0.12
		}
		if hasEvent(result.Events, "TEST") {
			result.Confidence += 0.08
		}
		if hasEvent(result.Events, "SOS") {
			result.Confidence += 0.08
		}
		if result.Confidence > 0.82 {
			result.Confidence = 0.82
		}
	}
	if result.Phase == PhaseUnknown {
		result.Evidence = append(result.Evidence, "Phase remains UNKNOWN: no sufficiently validated range and event sequence; wait for confirmation rather than anticipate.")
	}
	return result, nil
}
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
