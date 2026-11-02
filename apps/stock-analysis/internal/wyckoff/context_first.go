package wyckoff

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

const BaseAssessmentBars = 40
const MaxAssessmentBars = 160

type AssessmentWindow struct {
	BaseBars int       `json:"baseBars"`
	UsedBars int       `json:"usedBars"`
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	Reasons  []string  `json:"reasons"`
}
type SwingPoint struct {
	Type        string    `json:"type"`
	Label       string    `json:"label"`
	Date        time.Time `json:"date"`
	ConfirmedAt time.Time `json:"confirmedAt"`
	Price       float64   `json:"price"`
	index       int
}
type PriceStructure struct {
	Trend   string       `json:"trend"`
	Pattern string       `json:"pattern"`
	Swings  []SwingPoint `json:"swings"`
}
type FormingContext struct {
	Stage           string   `json:"stage"`
	Status          string   `json:"status"`
	PriorTrend      string   `json:"priorTrend"`
	Rationale       []string `json:"rationale"`
	CounterEvidence []string `json:"counterEvidence"`
	Alternatives    []string `json:"alternatives"`
}
type campaign struct {
	state                  StructureState
	stop, anchor, response int
	prior                  float64
	score                  float64
	phase                  Phase
	structure              Structure
	stage                  string
	counters               []string
}

func (a *Analyzer) Analyze(symbol string, candles []Candle) (Analysis, error) {
	if len(candles) < BaseAssessmentBars {
		return Analysis{}, ErrInsufficientData
	}
	cs := append([]Candle(nil), candles...)
	sort.Slice(cs, func(i, j int) bool { return cs[i].Date.Before(cs[j].Date) })
	for i, c := range cs {
		if c.Date.IsZero() || c.High < c.Low || c.High < math.Max(c.Open, c.Close) || c.Low > math.Min(c.Open, c.Close) || c.Close <= 0 || c.Volume < 0 || math.IsNaN(c.Open+c.High+c.Low+c.Close+c.Volume) || math.IsInf(c.Open+c.High+c.Low+c.Close+c.Volume, 0) {
			return Analysis{}, fmt.Errorf("invalid candle at index %d", i)
		}
		if i > 0 && !c.Date.After(cs[i-1].Date) {
			return Analysis{}, fmt.Errorf("duplicate candle date at index %d", i)
		}
	}
	return analyzeCampaigns(strings.ToUpper(strings.TrimSpace(symbol)), cs), nil
}

// Confirm pivots only after two later observations and discard tiny rotations.
// The last unfinished swing is never advertised as a confirmed HH/HL.
func confirmedSwings(cs []Candle, start int) PriceStructure {
	ps := PriceStructure{Trend: "SIDEWAYS", Pattern: "UNRESOLVED", Swings: []SwingPoint{}}
	for i := maxInt(2, start); i+2 < len(cs); i++ {
		hi, lo := true, true
		for j := i - 2; j <= i+2; j++ {
			if j == i {
				continue
			}
			if cs[j].High > cs[i].High || (j > i && cs[j].High == cs[i].High) {
				hi = false
			}
			if cs[j].Low < cs[i].Low || (j > i && cs[j].Low == cs[i].Low) {
				lo = false
			}
		}
		if hi && lo {
			continue
		}
		typ := ""
		p := 0.0
		if hi {
			typ = "HIGH"
			p = cs[i].High
		}
		if lo {
			typ = "LOW"
			p = cs[i].Low
		}
		if typ == "" {
			continue
		}
		s := SwingPoint{Type: typ, Date: cs[i].Date, ConfirmedAt: cs[i+2].Date, Price: p, index: i}
		if len(ps.Swings) > 0 {
			prev := ps.Swings[len(ps.Swings)-1]
			if prev.Type == typ {
				if (typ == "HIGH" && p >= prev.Price) || (typ == "LOW" && p <= prev.Price) {
					ps.Swings[len(ps.Swings)-1] = s
				}
				continue
			}
			if i-prev.index < 2 || math.Abs(p-prev.Price)/prev.Price < .02 {
				continue
			}
		}
		ps.Swings = append(ps.Swings, s)
	}
	var highs, lows []float64
	for i := range ps.Swings {
		s := &ps.Swings[i]
		if s.Type == "HIGH" {
			s.Label = "H"
			if len(highs) > 0 {
				s.Label = "EH"
				if s.Price > highs[len(highs)-1]*1.003 {
					s.Label = "HH"
				} else if s.Price < highs[len(highs)-1]*.997 {
					s.Label = "LH"
				}
			}
			highs = append(highs, s.Price)
		} else {
			s.Label = "L"
			if len(lows) > 0 {
				s.Label = "EL"
				if s.Price > lows[len(lows)-1]*1.003 {
					s.Label = "HL"
				} else if s.Price < lows[len(lows)-1]*.997 {
					s.Label = "LL"
				}
			}
			lows = append(lows, s.Price)
		}
	}
	if len(highs) >= 2 && len(lows) >= 2 {
		hh := highs[len(highs)-1] > highs[len(highs)-2]*1.003
		hl := lows[len(lows)-1] > lows[len(lows)-2]*1.003
		lh := highs[len(highs)-1] < highs[len(highs)-2]*.997
		ll := lows[len(lows)-1] < lows[len(lows)-2]*.997
		if hh && hl {
			ps.Trend = "UP"
			ps.Pattern = "HH_HL"
		} else if lh && ll {
			ps.Trend = "DOWN"
			ps.Pattern = "LH_LL"
		} else {
			ps.Pattern = "MIXED_SWINGS"
		}
	}
	return ps
}
func addCampaignEvent(c *campaign, cs []Candle, kind string, idx, observed int, why string) {
	if idx < 0 || observed < idx || observed >= len(cs) {
		return
	}
	for _, e := range c.state.Events {
		if e.Type == kind && e.Date.Equal(cs[idx].Date) {
			return
		}
	}
	d := cs[observed].Date
	c.state.Events = append(c.state.Events, Event{Type: kind, Date: cs[idx].Date, ObservedAt: &d, Status: "candidate", Confidence: .6, StructureID: c.state.ID, Evidence: why})
}
func moveBefore(cs []Candle, idx, window int) float64 {
	if idx < 1 {
		return 0
	}
	start := maxInt(0, idx-window)
	return (cs[idx].Close - cs[start].Close) / cs[start].Close
}
func reactionEnd(cs []Candle, anchor int, accum bool) int {
	// Establish the first response from a fifteen-session horizon. Its endpoint
	// needs two later observations; a later breakout never raises this boundary.
	end := minInt(len(cs)-3, anchor+15)
	if end <= anchor {
		return anchor
	}
	best := anchor
	for i := anchor + 1; i <= end; i++ {
		if (accum && cs[i].High >= cs[best].High) || (!accum && cs[i].Low <= cs[best].Low) {
			best = i
		}
	}
	return best
}
func stoppingCampaigns(symbol string, cs []Candle, ps PriceStructure) []campaign {
	var out []campaign
	start := maxInt(20, len(cs)-MaxAssessmentBars)
	for _, p := range ps.Swings {
		i := p.index
		if i < start {
			continue
		}
		accum := p.Type == "LOW"
		prior := moveBefore(cs, maxInt(0, i-3), 40)
		short := moveBefore(cs, maxInt(0, i-1), 8)
		if accum {
			prior = math.Min(prior, short)
		} else {
			prior = math.Max(prior, short)
		}
		if (accum && prior > -.08) || (!accum && prior < .08) {
			continue
		}
		anchor := i
		recentLo, recentHi := bounds(cs[maxInt(0, i-40):i])
		if (accum && cs[i].Low > recentLo*1.01) || (!accum && cs[i].High < recentHi*.99) {
			continue
		}
		// Treat adjacent extreme/recovery bars as one stopping cluster.
		stop := anchor
		bestEffort := 0.0
		for j := maxInt(start, anchor-3); j <= anchor; j++ {
			st := stats(cs[j], avgVolume(cs[maxInt(0, j-20):j]))
			daily := (cs[j].Close - cs[j-1].Close) / cs[j-1].Close
			effort := math.Abs(daily) * st.relativeVolume
			if accum && cs[j].Low <= cs[anchor].Low*1.05 && cs[j].Close < cs[j-1].Close && effort > bestEffort {
				bestEffort = effort
				stop = j
			}
			if !accum && cs[j].High >= cs[anchor].High*.95 && cs[j].Close > cs[j-1].Close && effort > bestEffort {
				bestEffort = effort
				stop = j
			}
		}
		ast := stats(cs[anchor], avgVolume(cs[maxInt(0, anchor-20):anchor]))
		if ast.relativeVolume >= 2 && ((accum && ast.closeLocation >= .5) || (!accum && ast.closeLocation <= .5)) {
			stop = anchor
		}
		st := stats(cs[stop], avgVolume(cs[maxInt(0, stop-20):stop]))
		daily := math.Abs((cs[stop].Close - cs[stop-1].Close) / cs[stop-1].Close)
		if st.relativeVolume < 1.2 && daily < .035 {
			continue
		}
		if !accum && (st.relativeVolume < 1.5 || (st.closeLocation > .65 && st.relativeVolume < 3)) {
			continue
		}
		ar := reactionEnd(cs, anchor, accum)
		if ar <= anchor+1 || ar+2 >= len(cs) {
			continue
		}
		support, resistance := cs[anchor].Low, cs[ar].High
		if !accum {
			support, resistance = cs[ar].Low, cs[anchor].High
		}
		width := (resistance - support) / ((resistance + support) / 2)
		if width < .04 || width > .40 {
			continue
		}
		c := campaign{stop: stop, anchor: anchor, response: ar, prior: prior, phase: PhaseA, structure: StructureAccumulation, stage: "FORMING_ACCUMULATION_A"}
		kind, label := "accumulation", "SC"
		if !accum {
			kind, label = "distribution", "BC"
			c.structure = StructureDistribution
			c.stage = "FORMING_DISTRIBUTION_A"
		}
		c.state = StructureState{ID: fmt.Sprintf("%s_%s_%s", symbol, strings.ToUpper(kind), cs[stop].Date.Format("20060102")), Kind: kind, StartDate: cs[stop].Date, RangeEnd: cs[len(cs)-1].Date, IdentifiedAt: cs[ar+2].Date, LastObserved: cs[len(cs)-1].Date, Support: support, Resistance: resistance, Status: "RANGE", Window: ar - stop + 1, Events: []Event{}, Transitions: []StructureTransition{}}
		addCampaignEvent(&c, cs, label, stop, ar+2, "Stopping cluster after a preceding directional move; selected in the context of its subsequent automatic response.")
		addCampaignEvent(&c, cs, "AR", ar, ar+2, "Automatic response of this stopping cluster defines the opposite range boundary; later breakout highs do not replace it.")
		// Preliminary support/supply is optional, and must precede the climax.
		for j := maxInt(start, stop-8); j < stop; j++ {
			s := stats(cs[j], avgVolume(cs[maxInt(0, j-20):j]))
			if s.relativeVolume >= 1.3 && ((accum && s.closeLocation >= .5) || (!accum && s.closeLocation <= .5)) {
				k := "PS"
				if !accum {
					k = "PSY"
				}
				addCampaignEvent(&c, cs, k, j, ar+2, "Preliminary stopping effort before the climax in the same campaign.")
				break
			}
		}
		evaluateCampaign(&c, cs)
		scoreCampaign(&c, cs)
		out = append(out, c)
	}
	return out
}

// Continuation candidates are built from at least forty observations before
// their exit. A preceding advance alone is not proof of distribution.
func continuationCampaigns(symbol string, cs []Candle) []campaign {
	var out []campaign
	start := maxInt(59, len(cs)-MaxAssessmentBars+39)
	for end := start; end < len(cs)-3; end++ {
		lo, hi, prior, ok := lifecycleRange(cs, end, 40)
		if !ok {
			continue
		}
		duplicate := false
		for _, old := range out {
			w := old.state.Resistance - old.state.Support
			if math.Abs(old.state.Support-lo) < w*.15 && math.Abs(old.state.Resistance-hi) < w*.15 {
				duplicate = true
				break
			}
		}
		if duplicate {
			continue
		}
		c := campaign{stop: end - 39, anchor: end - 39, response: end, prior: prior, structure: StructureUnclear, phase: PhaseUnknown, stage: "FORMING_CONTINUATION_RANGE"}
		c.state = StructureState{ID: fmt.Sprintf("%s_CONTINUATION_%s", symbol, cs[end-39].Date.Format("20060102")), Kind: "range", StartDate: cs[end-39].Date, RangeEnd: cs[len(cs)-1].Date, IdentifiedAt: cs[end].Date, LastObserved: cs[len(cs)-1].Date, Support: lo, Resistance: hi, Status: "RANGE", Window: 40, Events: []Event{}, Transitions: []StructureTransition{}}
		if prior > .08 {
			c.state.Kind = "continuation"
		}
		evaluateCampaign(&c, cs)
		scoreCampaign(&c, cs)
		out = append(out, c)
	}
	return out
}

// Narrow only after a larger campaign has been established on the >=40-bar
// context. A short range never replaces that initial context by itself.
func innerCampaigns(symbol string, cs []Candle, parents []campaign) []campaign {
	var out []campaign
	n := len(cs)
	for _, parent := range parents {
		if parent.state.Kind != "accumulation" || cs[n-1].Close < parent.state.Support || cs[n-1].Close > parent.state.Resistance {
			continue
		}
		for _, w := range []int{12, 20} {
			for end := maxInt(n-40, w+20); end < n-3; end++ {
				lo, hi, _, ok := lifecycleRange(cs, end, w)
				if !ok || lo < parent.state.Support || hi > parent.state.Resistance || hi-lo > (parent.state.Resistance-parent.state.Support)*.6 || !cs[end].Date.After(parent.state.IdentifiedAt) {
					continue
				}
				duplicate := false
				for _, old := range out {
					if math.Abs(old.state.Support-lo) < (hi-lo)*.1 && math.Abs(old.state.Resistance-hi) < (hi-lo)*.1 {
						duplicate = true
						break
					}
				}
				if duplicate {
					continue
				}
				c := campaign{stop: end - w + 1, anchor: end - w + 1, response: end, prior: parent.prior, structure: StructureAccumulation, phase: PhaseB, stage: "FORMING_INNER_ACCUMULATION_B"}
				c.state = StructureState{ID: fmt.Sprintf("%s_INNER_%s", symbol, cs[end-w+1].Date.Format("20060102")), ParentID: parent.state.ID, Kind: "accumulation", StartDate: cs[end-w+1].Date, RangeEnd: cs[n-1].Date, IdentifiedAt: cs[end].Date, LastObserved: cs[n-1].Date, Support: lo, Resistance: hi, Status: "RANGE", Window: w, Events: []Event{}, Transitions: []StructureTransition{}}
				evaluateCampaign(&c, cs)
				broken := false
				for i := end + 2; i < n; i++ {
					if c.state.BreakoutDate != nil && !cs[i].Date.Before(*c.state.BreakoutDate) {
						break
					}
					if cs[i].Close < lo && cs[i-1].Close < lo {
						broken = true
						break
					}
				}
				if broken {
					continue
				}
				scoreCampaign(&c, cs)
				if hasEvent(c.state.Events, "SPRING") && hasEvent(c.state.Events, "SOS") {
					c.score += 1.5
				}
				out = append(out, c)
			}
		}
	}
	return out
}
func evaluateCampaign(c *campaign, cs []Candle) {
	accum := c.state.Kind != "distribution"
	lo, hi := c.state.Support, c.state.Resistance
	w := hi - lo
	n := len(cs)
	last := cs[n-1]
	stCount, spring, springLow, springRecovery, breakout, lastST := 0, -1, 0.0, -1, -1, -100
	for i := c.response + 1; i < n; i++ {
		if !cs[i].Date.After(c.state.IdentifiedAt) {
			continue
		}
		bar := cs[i]
		s := stats(bar, avgVolume(cs[maxInt(0, i-20):i]))
		if breakout < 0 {
			if c.state.Kind == "continuation" && bar.Close < lo && s.closeLocation <= .35 && s.relativeVolume >= 1.1 {
				accum = false
				c.state.Kind = "distribution"
				c.structure = StructureDistribution
			}
			near := bar.Low <= lo+w*.25 && bar.Low >= lo
			if !accum {
				near = bar.High >= hi-w*.25 && bar.High <= hi
			}
			quiet := bar.Volume <= cs[c.stop].Volume*1.3 && s.spread < w*.7
			// A test is confirmed only by an observed reaction, not by its own candle.
			rebound := false
			if i+1 < n {
				rebound = cs[i+1].Close > bar.Close
				if !accum {
					rebound = cs[i+1].Close < bar.Close
				}
			}
			if rebound && i+1 < n {
				if accum {
					rebound = cs[i+1].Low > bar.Low
				} else {
					rebound = cs[i+1].High < bar.High
				}
			}
			if near && quiet && rebound && i-lastST >= 5 && spring < 0 && (hasEvent(c.state.Events, "SC") || hasEvent(c.state.Events, "BC")) {
				addCampaignEvent(c, cs, "ST", i, i+1, "Secondary test of this campaign's stopping boundary with reduced selling/buying effort and an observed reaction.")
				stCount++
				lastST = i
			}
			penetration := bar.Low < lo && bar.Low >= lo-w*.8
			if !accum {
				penetration = bar.High > hi && bar.High <= hi+w*.8
			}
			if penetration && (spring < 0 || (accum && bar.Low < springLow) || (!accum && bar.High > springLow)) {
				recovered := -1
				extreme := bar.Low
				if !accum {
					extreme = bar.High
				}
				for j := i; j < n && j <= i+3; j++ {
					if accum {
						extreme = math.Min(extreme, cs[j].Low)
					} else {
						extreme = math.Max(extreme, cs[j].High)
					}
					loc := stats(cs[j], 1).closeLocation
					if (accum && cs[j].Close > lo && loc >= .5) || (!accum && cs[j].Close < hi && loc <= .5) {
						recovered = j
						break
					}
				}
				if recovered >= 0 {
					kind := "SPRING"
					if !accum {
						kind = "UT"
						if stCount >= 2 {
							kind = "UTAD"
						}
					}
					addCampaignEvent(c, cs, kind, i, recovered, "Penetration and timely recovery of a boundary established by this campaign before the event.")
					spring, springLow, springRecovery = i, extreme, recovered
					i = recovered
				}
			}
			if accum && bar.Close > hi && s.closeLocation >= .65 && s.relativeVolume >= 1.2 {
				// Require price to have been inside this range shortly before breakout.
				crossed := false
				for j := maxInt(c.response, i-3); j < i; j++ {
					if cs[j].Close <= hi {
						crossed = true
					}
				}
				if crossed {
					breakout = i
					for j := i + 1; j < n && j <= i+2; j++ {
						js := stats(cs[j], avgVolume(cs[maxInt(0, j-20):j]))
						if cs[j].Close > cs[j-1].Close && cs[j].Close > hi && js.closeLocation >= .65 && js.relativeVolume >= 1.2 {
							breakout = j
						} else {
							break
						}
					}
					addCampaignEvent(c, cs, "SOS", breakout, breakout, "Range breakout with demand and observed follow-through; dated at the final qualifying breakout bar.")
				}
			}
			if !accum && bar.Close < lo && s.closeLocation <= .35 && s.relativeVolume >= 1.1 {
				crossed := false
				for j := maxInt(c.response, i-3); j < i; j++ {
					if cs[j].Close >= lo {
						crossed = true
					}
				}
				if crossed {
					breakout = i
					for j := i + 1; j < n && j <= i+2; j++ {
						js := stats(cs[j], avgVolume(cs[maxInt(0, j-20):j]))
						if cs[j].Close < cs[j-1].Close && cs[j].Close < lo && js.closeLocation <= .35 && js.relativeVolume >= 1.1 {
							breakout = j
						} else {
							break
						}
					}
					addCampaignEvent(c, cs, "SOW", breakout, breakout, "Weakness breaks this distribution range after BC/AR; no SOW is inferred from an accumulation failure alone.")
				}
			}
		}
		if spring >= 0 && breakout < 0 && i > springRecovery && i <= springRecovery+15 {
			holds := bar.Low >= springLow && bar.Close >= lo && bar.Low <= lo+w*.2 && bar.Low <= springLow+w*.35 && s.relativeVolume < .8 && s.closeLocation >= .55
			if !accum {
				holds = bar.High <= springLow && bar.Close <= hi && bar.High >= hi-w*.2 && s.relativeVolume < .8 && s.closeLocation <= .45
			}
			valid := true
			for j := springRecovery + 1; j <= i; j++ {
				if (accum && cs[j].Low < springLow) || (!accum && cs[j].High > springLow) {
					valid = false
				}
			}
			if holds && valid {
				addCampaignEvent(c, cs, "TEST", i, i, "Reduced-effort Test holds the same Spring/UTAD extreme and original range boundary.")
				springRecovery = -100
			}
		}
	}
	if stCount > 0 {
		c.phase = PhaseB
		c.stage = "FORMING_ACCUMULATION_B"
		if !accum {
			c.stage = "FORMING_DISTRIBUTION_B"
		}
	}
	if spring >= 0 {
		c.phase = PhaseC
		c.stage = "FORMING_ACCUMULATION_C"
		if !accum {
			c.stage = "FORMING_DISTRIBUTION_C"
		}
	}
	if !accum && hasEvent(c.state.Events, "UT") && !hasEvent(c.state.Events, "UTAD") {
		c.phase = PhaseB
		c.stage = "FORMING_DISTRIBUTION_B_UPTHRUST"
	}
	if breakout >= 0 {
		c.phase = PhaseD
		c.stage = "ACCUMULATION_D_BREAKOUT_TESTING"
		if !accum {
			c.stage = "DISTRIBUTION_D_BREAKDOWN_TESTING"
		}
		d := cs[breakout].Date
		c.state.BreakoutDate = &d
		c.state.RangeEnd = cs[breakout-1].Date
		c.state.Status = "BREAKOUT_UP"
		if !accum {
			c.state.Status = "BREAKDOWN"
		}
		best := -1
		for i := breakout + 1; i < n && i <= breakout+15; i++ {
			s := stats(cs[i], avgVolume(cs[maxInt(0, i-20):i]))
			ok := cs[i].Low >= hi-w*.15 && cs[i].Low <= hi+w*.35 && cs[i].Close >= hi && cs[i].Close < cs[i-1].Close && s.relativeVolume < 1.05
			if !accum {
				ok = cs[i].High <= lo+w*.15 && cs[i].High >= lo-w*.35 && cs[i].Close <= lo && cs[i].Close > cs[i-1].Close && s.relativeVolume < 1.05
			}
			if ok && i+2 < n {
				if (accum && cs[i+1].Close > cs[i].Close && cs[i+2].Close >= cs[i+1].Close && cs[i+1].Close >= hi && cs[i+2].Close >= hi) || (!accum && cs[i+1].Close < cs[i].Close && cs[i+2].Close <= cs[i+1].Close && cs[i+1].Close <= lo && cs[i+2].Close <= lo) {
					if best < 0 || (accum && cs[i].Low < cs[best].Low) || (!accum && cs[i].High > cs[best].High) {
						best = i
					}
				}
			}
		}
		if best >= 0 {
			kind := "LPS"
			if !accum {
				kind = "LPSY"
			}
			addCampaignEvent(c, cs, kind, best, best+2, "Confirmed reaction after a pullback to this structure's original breakout/breakdown boundary.")
		}
		ps := confirmedSwings(cs, maxInt(c.response, n-BaseAssessmentBars))
		if accum && last.Close > hi && n-1-breakout >= 5 && (ps.Pattern == "HH_HL" || (last.Close > hi+w*.4 && best >= 0)) {
			c.structure = StructureMarkup
			c.phase = PhaseE
			c.stage = "MARKUP_HH_HL"
			c.state.Status = "MARKUP"
			if c.prior > .08 {
				c.state.Kind = "reaccumulation"
			}
		}
		if !accum && last.Close < lo && n-1-breakout >= 5 && (ps.Pattern == "LH_LL" || (last.Close < lo-w*.4 && best >= 0)) {
			c.structure = StructureMarkdown
			c.phase = PhaseE
			c.stage = "MARKDOWN_LH_LL"
			c.state.Status = "MARKDOWN"
		}
		if accum && last.Close < hi {
			c.counters = append(c.counters, "Price has returned below original breakout resistance; Phase D remains under test, not a confirmed LPS or new distribution.")
		}
		if !accum && last.Close > lo {
			c.counters = append(c.counters, "Price has reclaimed the breakdown boundary; distribution continuation is unconfirmed.")
		}
	}
	failed := false
	if accum {
		for i := c.response + 2; i < n; i++ {
			if cs[i].Close < lo*.995 && cs[i-1].Close < lo*.995 {
				failed = true
				break
			}
		}
	}
	if failed {
		c.structure = StructureMarkdown
		c.phase = PhaseE
		c.stage = "MARKDOWN_AFTER_FAILED_ACCUMULATION"
		c.state.Status = "FAILED_ACCUMULATION"
		c.counters = append(c.counters, "Two closes below the stopping support invalidate the accumulation hypothesis; this is not automatically distribution SOW.")
	}
	if (c.state.Kind == "continuation" || c.state.Kind == "range") && breakout < 0 {
		c.structure = StructureUnclear
		c.phase = PhaseUnknown
		c.stage = "FORMING_CONTINUATION_RANGE"
	}
	sort.SliceStable(c.state.Events, func(i, j int) bool { return c.state.Events[i].Date.Before(c.state.Events[j].Date) })
}
func pendingCampaigns(symbol string, cs []Candle) []campaign {
	var out []campaign
	n := len(cs)
	for i := maxInt(20, n-5); i < n; i++ {
		st := stats(cs[i], avgVolume(cs[maxInt(0, i-20):i]))
		daily := (cs[i].Close - cs[i-1].Close) / cs[i-1].Close
		prior := moveBefore(cs, i-1, 8)
		accum := prior < -0.08
		distribution := prior > .08
		if !accum && !distribution {
			continue
		}
		shock := st.relativeVolume >= 3 || st.relativeVolume >= 2 && math.Abs(daily) >= .04
		if !shock {
			continue
		}
		if accum && st.closeLocation < .5 && daily > -0.04 {
			continue
		}
		if distribution && st.closeLocation > .5 {
			continue
		}
		support, resistance := bounds(cs[maxInt(0, i-39) : i+1])
		kind, label, stage := "accumulation", "SC", "POTENTIAL_STOPPING_ACCUMULATION_A"
		structure := StructureAccumulation
		if distribution {
			kind, label, stage = "distribution", "BC", "POTENTIAL_STOPPING_DISTRIBUTION_A"
			structure = StructureDistribution
		}
		c := campaign{stop: i, anchor: i, response: i, prior: prior, structure: structure, phase: PhaseA, stage: stage, counters: []string{"Climactic effort is provisional; an automatic response and subsequent test have not yet established a range."}}
		c.state = StructureState{ID: fmt.Sprintf("%s_%s_%s", symbol, strings.ToUpper(kind), cs[i].Date.Format("20060102")), Kind: kind, StartDate: cs[i].Date, RangeEnd: cs[n-1].Date, IdentifiedAt: cs[i].Date, LastObserved: cs[n-1].Date, Support: support, Resistance: resistance, Status: "PENDING_RESPONSE", Window: 40, Events: []Event{}, Transitions: []StructureTransition{}}
		addCampaignEvent(&c, cs, label, i, i, "Potential stopping climax after a directional leg; AR and range boundaries are not yet confirmed.")
		c.score = 6 + math.Min(1, st.relativeVolume/10)
		out = append(out, c)
	}
	return out
}
func scoreCampaign(c *campaign, cs []Candle) {
	n := len(cs)
	lo, hi := c.state.Support, c.state.Resistance
	w := hi - lo
	inside := 0
	for _, b := range cs[maxInt(0, n-BaseAssessmentBars):] {
		if b.Close >= lo-w*.1 && b.Close <= hi+w*.1 {
			inside++
		}
	}
	c.score = float64(inside) / float64(minInt(BaseAssessmentBars, n)) * 3
	if hasEvent(c.state.Events, "SC") || hasEvent(c.state.Events, "BC") {
		c.score += 1
	} else {
		c.score -= .75
	}
	if hasEvent(c.state.Events, "ST") {
		c.score += .8
	}
	if hasEvent(c.state.Events, "SPRING") || hasEvent(c.state.Events, "UTAD") {
		c.score += .3
	}
	if c.state.BreakoutDate != nil {
		c.score += 1.2
		for i := n - 1; i >= maxInt(0, n-11); i-- {
			if cs[i].Date.Equal(*c.state.BreakoutDate) {
				c.score += 1
				break
			}
		}
	}
	if hasEvent(c.state.Events, "SOW") {
		c.score += 1
	}
	if c.structure == StructureMarkup || c.structure == StructureMarkdown {
		c.score += 1
	}
	// Prefer a complete current campaign to an isolated newer high-volume candle.
	c.score += math.Min(1, float64(n-c.response)/40)
	if n-c.response < 8 {
		c.score -= 1.5
	}
	if n-c.response > 120 {
		c.score -= 1
	}
	if c.state.Status == "FAILED_ACCUMULATION" {
		c.score -= 3
	}
	latest := c.response
	for _, e := range c.state.Events {
		for i := n - 1; i >= 0; i-- {
			if cs[i].Date.Equal(e.Date) {
				if i > latest {
					latest = i
				}
				break
			}
		}
	}
	if n-1-latest > 40 {
		current := confirmedSwings(cs, maxInt(0, n-40))
		if c.structure == StructureMarkdown && hasEvent(c.state.Events, "SOW") && current.Pattern == "LH_LL" {
			c.score -= 1
		} else {
			c.score -= 4
		}
	}
	if c.state.Kind == "accumulation" && c.state.Status != "FAILED_ACCUMULATION" && c.response >= n-40 {
		effort := stats(cs[c.stop], avgVolume(cs[maxInt(0, c.stop-20):c.stop]))
		if effort.relativeVolume >= 3 && cs[n-1].Close >= lo && cs[n-1].Close <= hi {
			c.score += 2
		}
	}
}
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func analyzeCampaigns(symbol string, cs []Candle) Analysis {
	n := len(cs)
	start := maxInt(0, n-BaseAssessmentBars)
	ps := confirmedSwings(cs, maxInt(0, n-MaxAssessmentBars))
	candidates := stoppingCampaigns(symbol, cs, ps)
	candidates = append(candidates, pendingCampaigns(symbol, cs)...)
	candidates = append(candidates, continuationCampaigns(symbol, cs)...)
	candidates = append(candidates, innerCampaigns(symbol, cs, candidates)...)
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].score == candidates[j].score {
			return candidates[i].state.StartDate.After(candidates[j].state.StartDate)
		}
		return candidates[i].score > candidates[j].score
	})
	lo, hi := bounds(cs[start:])
	a := Analysis{Symbol: symbol, AsOf: cs[n-1].Date, Structure: StructureUnclear, Phase: PhaseUnknown, Confidence: .25, TradingRange: &TradingRange{Support: lo, Resistance: hi, WidthPct: 100 * (hi - lo) / ((hi + lo) / 2)}, DataPoints: n, DataMode: "real", Events: []Event{}, PresentationEvents: []Event{}, Structures: []StructureState{}, Evidence: []string{}}
	a.PriceStructure = confirmedSwings(cs, start)
	a.Context = FormingContext{Stage: "FORMING_RANGE_UNRESOLVED", Status: "developing", PriorTrend: "unknown", Rationale: []string{}, CounterEvidence: []string{}, Alternatives: []string{"Range hypothesis has insufficient stopping/response evidence."}}
	if len(candidates) > 0 {
		c := candidates[0]
		a.Structure, a.Phase = c.structure, c.phase
		a.Confidence = .45 + math.Min(.25, c.score*.035)
		a.ActiveStructureID = c.state.ID
		a.PresentationStructureID = c.state.ID
		a.TradingRange = &TradingRange{Support: c.state.Support, Resistance: c.state.Resistance, WidthPct: 100 * (c.state.Resistance - c.state.Support) / ((c.state.Resistance + c.state.Support) / 2)}
		a.Events = append([]Event{}, c.state.Events...)
		a.PresentationEvents = append([]Event{}, a.Events...)
		start = minInt(start, maxInt(0, c.stop-40))
		for _, e := range a.Events {
			for j := start - 1; j >= 0; j-- {
				if cs[j].Date.Equal(e.Date) {
					start = j
					break
				}
			}
		}
		a.Context.Stage = c.stage
		a.Context.PriorTrend = "unknown"
		if c.prior < -0.08 {
			a.Context.PriorTrend = "down"
		}
		if c.prior > .08 {
			a.Context.PriorTrend = "up"
		}
		a.Context.Rationale = []string{"Selected one campaign from stopping context, automatic response and subsequent tests before assigning its phase.", fmt.Sprintf("Candidate uses support %.3f and resistance %.3f established before breakout evaluation.", c.state.Support, c.state.Resistance)}
		a.Context.CounterEvidence = c.counters
		a.Context.Alternatives = []string{"The structure remains a hypothesis; boundary failure or a conflicting swing sequence requires reassessment."}
		if a.Structure == StructureUnclear && a.PriceStructure.Pattern == "LH_LL" && moveBefore(cs, n-1, 40) < -0.08 {
			a.Structure = StructureMarkdown
			a.Phase = PhaseE
			a.Context.Stage = "MARKDOWN_LH_LL_WITH_UNRESOLVED_RANGE"
		}

		for i, s := range candidates {
			if i < 12 || s.state.ID == c.state.ParentID {
				a.Structures = append(a.Structures, s.state)
			}
			if s.state.ID == c.state.ParentID {
				start = minInt(start, maxInt(0, s.stop-40))
			}
		}
		// Retain a broad swing envelope separately from the first automatic
		// response; it is context, never a source of extra SC/BC/SOS labels.
		for _, parent := range candidates {
			if parent.state.Kind != "accumulation" || !hasEvent(parent.state.Events, "SC") {
				continue
			}
			broadHi := parent.state.Resistance
			for j := parent.anchor; j < n && j <= parent.anchor+40; j++ {
				broadHi = math.Max(broadHi, cs[j].High)
			}
			if broadHi <= parent.state.Resistance {
				continue
			}
			b := parent.state
			b.ID = parent.state.ID + "_CONTEXT"
			b.Kind = "context"
			b.Resistance = broadHi
			b.Events = []Event{}
			b.Status = "CONTEXT"
			a.Structures = append(a.Structures, b)
		}
	} else {
		if a.PriceStructure.Pattern == "HH_HL" {
			a.Structure = StructureMarkup
			a.Phase = PhaseE
			a.Context.Stage = "MARKUP_HH_HL_WITHOUT_CONFIRMED_RANGE"
			a.Confidence = .45
		}
		if a.PriceStructure.Pattern == "LH_LL" {
			a.Structure = StructureMarkdown
			a.Phase = PhaseE
			a.Context.Stage = "MARKDOWN_LH_LL_WITHOUT_CONFIRMED_RANGE"
			a.Confidence = .45
		}
		if a.PriceStructure.Pattern == "UNRESOLVED" {
			change := moveBefore(cs, n-1, 39)
			if change > .08 {
				a.Structure = StructureMarkup
				a.Phase = PhaseE
				a.Context.Stage = "FORMING_MARKUP_UNCONFIRMED_SWINGS"
				a.PriceStructure.Pattern = "RISING_SWINGS_UNCONFIRMED"
				a.PriceStructure.Trend = "UP_DEVELOPING"
				a.Confidence = .35
			}
			if change < -.08 {
				a.Structure = StructureMarkdown
				a.Phase = PhaseE
				a.Context.Stage = "FORMING_MARKDOWN_UNCONFIRMED_SWINGS"
				a.PriceStructure.Pattern = "FALLING_SWINGS_UNCONFIRMED"
				a.PriceStructure.Trend = "DOWN_DEVELOPING"
				a.Confidence = .35
			}
		}
	}
	a.Window = AssessmentWindow{BaseBars: BaseAssessmentBars, UsedBars: n - start, Start: cs[start].Date, End: cs[n-1].Date, Reasons: []string{"Begin with the latest 40 observations; confirmed pivots need two later observations."}}
	if n-start > BaseAssessmentBars {
		a.Window.Reasons = append(a.Window.Reasons, "Expanded to retain the selected stopping cluster and its automatic response; do not redefine support using a later Spring low or resistance using a breakout high.")
	}
	if len(candidates) > 0 && candidates[0].state.ParentID != "" {
		a.Window.Reasons = append(a.Window.Reasons, "Narrowed the event range inside an established parent campaign; the assessment still includes >=40 observations and retains parent context.")
	}
	a.Evidence = append(a.Evidence, fmt.Sprintf("Current stage: %s; base=%d, used=%d; swing pattern=%s.", a.Context.Stage, a.Window.BaseBars, a.Window.UsedBars, a.PriceStructure.Pattern))
	for _, e := range a.Events {
		a.Evidence = append(a.Evidence, fmt.Sprintf("%s (%s) [%s]: %s", e.Type, e.Date.Format("2006-01-02"), e.StructureID, e.Evidence))
	}
	a.Evidence = append(a.Evidence, a.Context.CounterEvidence...)
	return a
}
