package wyckoff

import "testing"

func TestSpringRetestSequence(t *testing.T) {
	rc := rangeCandidate{support: 100, resistance: 110, valid: true}
	base := []Candle{
		{Open: 101, High: 103, Low: 99, Close: 102, Volume: 100},
		{Open: 102, High: 104, Low: 101, Close: 103, Volume: 100},
		{Open: 100, High: 102, Low: 100, Close: 101.5, Volume: 50},
	}
	if got := findSpringTest(base, 0, rc, 20); got != 2 {
		t.Fatalf("valid retest: got %d, want 2", got)
	}
	if got := findSpringTest(base[:2], 0, rc, 20); got != -1 {
		t.Fatal("must not see a future retest")
	}
	broken := append([]Candle(nil), base...)
	broken[1].Low = 98
	if got := findSpringTest(broken, 0, rc, 20); got != -1 {
		t.Fatal("lower low must invalidate before a later retest")
	}
	belowSupport := append([]Candle(nil), base...)
	belowSupport[2] = Candle{Open: 99.5, High: 100, Low: 99, Close: 99.8, Volume: 50}
	if got := findSpringTest(belowSupport, 0, rc, 20); got != -1 {
		t.Fatal("retest must close back above original support")
	}
	highVolume := append([]Candle(nil), base...)
	highVolume[2].Volume = 150
	if got := findSpringTest(highVolume, 0, rc, 20); got != -1 {
		t.Fatal("high-volume retest must not qualify as reduced supply")
	}
}
