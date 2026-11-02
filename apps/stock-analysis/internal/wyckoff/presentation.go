package wyckoff

import (
	"math"
	"sort"
)

// presentationEvents selects a conservative, single-range timeline. Candidate
// inventories remain in Analysis.Events/Structures for evaluation and inspection.
func presentationEvents(a Analysis, cs []Candle) ([]Event, string) {
	out := []Event{}
	if a.TradingRange == nil || len(cs) == 0 {
		return out, ""
	}
	support, resistance := a.TradingRange.Support, a.TradingRange.Resistance
	width := resistance - support
	if width <= 0 {
		return out, ""
	}
	primary := ""
	for _, s := range a.Structures {
		if s.Kind != "context" && math.Abs(s.Support-support) < width*.001 && math.Abs(s.Resistance-resistance) < width*.001 {
			if primary == "" || s.ID == a.ActiveStructureID {
				primary = s.ID
			}
		}
	}
	byDate := map[string]int{}
	for i, c := range cs {
		byDate[c.Date.Format("2006-01-02")] = i
	}
	candidates := append([]Event(nil), a.Events...)
	sort.SliceStable(candidates, func(i, j int) bool { return candidates[i].Date.Before(candidates[j].Date) })
	seen := map[string]bool{}
	var valid []Event
	for _, e := range candidates {
		if e.StructureID != "" && e.StructureID != primary {
			continue
		}
		if e.Status == "invalidated" || (e.ObservedAt != nil && e.ObservedAt.After(a.AsOf)) {
			continue
		}
		i, ok := byDate[e.Date.Format("2006-01-02")]
		if !ok || i == 0 {
			continue
		}
		c := cs[i]
		st := stats(c, avgVolume(cs[maxInt(0, i-20):i]))
		keep := false
		switch e.Type {
		case "PS":
			keep = c.Low <= support+width*.25
		case "SC":
			keep = c.Close < cs[i-1].Close && c.Low <= support+width*.2
		case "AR":
			keep = c.High >= resistance-width*.2
		case "ST":
			keep = c.Low <= support+width*.2 && st.relativeVolume < 1
		case "SPRING":
			if c.Low >= support {
				break
			}
			for j := i; j <= i+3 && j < len(cs); j++ {
				if cs[j].Close > support && stats(cs[j], 1).closeLocation >= .5 {
					keep = true
					break
				}
			}
			for _, s := range a.Structures {
				if s.ID == primary {
					for _, attempt := range s.SpringAttempts {
						if attempt.Date.Equal(e.Date) && attempt.Status == "INVALIDATED" {
							keep = false
						}
					}
				}
			}
		case "SOS":
			keep = c.Close > resistance && st.closeLocation >= .65 && st.relativeVolume >= 1.2
			crossing := false
			for j := maxInt(0, i-3); j < i; j++ {
				if cs[j].Close <= resistance {
					crossing = true
				}
			}
			keep = keep && crossing
		case "LPS":
			keep = c.Close >= resistance && c.Low >= resistance-width*.15 && c.Low <= resistance+width*.35 && c.Close < cs[i-1].Close && st.relativeVolume < 1
		case "TEST":
			keep = c.Close >= support && c.Low <= support+width*.2 && st.relativeVolume < .8 && st.closeLocation >= .55
		case "BC":
			keep = c.High >= resistance-width*.2 && st.relativeVolume >= 1.5
		case "UTAD", "UT":
			keep = c.High > resistance && c.Close < resistance
		case "SOW", "LPSY":
			keep = a.Structure == StructureDistribution && c.Close < support
		}
		key := e.Type + "|" + e.Date.Format("2006-01-02")
		if keep && !seen[key] {
			seen[key] = true
			valid = append(valid, e)
		}
	}
	sc, ar, spring, sos := -1, -1, -1, -1
	for i, e := range valid {
		switch e.Type {
		case "SC":
			sc = i
		case "AR":
			if sc >= 0 && ar < 0 && e.Date.After(valid[sc].Date) && byDate[e.Date.Format("2006-01-02")]-byDate[valid[sc].Date.Format("2006-01-02")] <= 15 {
				ar = i
			}
		case "SPRING":
			spring = i
		case "SOS":
			if sos < 0 {
				sos = i
			} else if byDate[e.Date.Format("2006-01-02")]-byDate[valid[sos].Date.Format("2006-01-02")] <= 3 {
				sos = i
			}
		}
	}
	// Only the latest recovered Spring preceding this range's SOS is relevant.
	if sos >= 0 {
		spring = -1
		for i, e := range valid {
			if e.Type == "SPRING" && e.Date.Before(valid[sos].Date) {
				spring = i
			}
		}
	}
	test, lps := -1, -1
	for i, e := range valid {
		idx := byDate[e.Date.Format("2006-01-02")]
		switch e.Type {
		case "PS":
			if sc >= 0 && e.Date.Before(valid[sc].Date) {
				out = append(out, e)
			}
		case "SC":
			if i == sc {
				out = append(out, e)
			}
		case "AR":
			if i == ar {
				out = append(out, e)
			}
		case "ST":
			if sc >= 0 && ar >= 0 && e.Date.After(valid[ar].Date) && cs[idx].Low >= cs[byDate[valid[sc].Date.Format("2006-01-02")]].Low {
				out = append(out, e)
			}
		case "SPRING":
			if i == spring {
				out = append(out, e)
			}
		case "SOS":
			if i == sos {
				out = append(out, e)
			}
		case "TEST":
			if spring < 0 || test >= 0 || !e.Date.After(valid[spring].Date) || (sos >= 0 && !e.Date.Before(valid[sos].Date)) {
				continue
			}
			start := byDate[valid[spring].Date.Format("2006-01-02")]
			if idx-start > 15 {
				continue
			}
			low := cs[start].Low
			for j := start + 1; j <= start+3 && j < len(cs); j++ {
				if cs[j].Low < low {
					low = cs[j].Low
				}
				if cs[j].Close > support {
					break
				}
			}
			held := true
			for j := start + 1; j <= idx; j++ {
				if cs[j].Low < low {
					held = false
				}
			}
			if held && cs[idx].Low <= low+width*.35 {
				test = i
				out = append(out, e)
			}
		case "LPS":
			if sos < 0 || lps >= 0 || !e.Date.After(valid[sos].Date) || idx-byDate[valid[sos].Date.Format("2006-01-02")] > 15 {
				continue
			}
			// A current down candle alone cannot establish a successful pullback.
			if idx+1 < len(cs) && cs[idx+1].Close > cs[idx].Close {
				d := cs[idx+1].Date
				e.ObservedAt = &d
				lps = i
				out = append(out, e)
			}
		default:
			out = append(out, e)
		}
	}
	return out, primary
}
