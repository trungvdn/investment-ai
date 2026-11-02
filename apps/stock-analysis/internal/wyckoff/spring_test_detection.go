package wyckoff

// findSpringTest uses only the captured range and bars available to the caller.
// A breach of the Spring low invalidates this sequence before any later retest.
func findSpringTest(cs []Candle, springIdx int, rc rangeCandidate, volumeWindow int) int {
	if springIdx < 0 || springIdx >= len(cs) || !rc.valid || volumeWindow < 1 {
		return -1
	}
	end := springIdx + 15
	if end >= len(cs) {
		end = len(cs) - 1
	}
	for j := springIdx + 1; j <= end; j++ {
		if cs[j].Low < cs[springIdx].Low {
			return -1
		}
		if j < springIdx+2 {
			continue
		}
		start := j - volumeWindow
		if start < 0 {
			start = 0
		}
		ts := stats(cs[j], avgVolume(cs[start:j]))
		if cs[j].Close >= rc.support && ts.relativeVolume < 0.8 && ts.closeLocation >= 0.55 && cs[j].Low <= rc.support+(rc.resistance-rc.support)*0.2 {
			return j
		}
	}
	return -1
}
