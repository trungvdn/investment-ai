package wyckoff

import (
	"errors"
	"math"
	"testing"
	"time"
)

func candles(n int) []Candle {
	out := make([]Candle, n)
	for i := range out {
		v := 100 + float64(i%5)*0.3
		out[i] = Candle{Date: time.Date(2025, 1, 1+i, 0, 0, 0, 0, time.UTC), Open: v - 0.2, High: v + 0.8, Low: v - 0.8, Close: v, Volume: 1000}
	}
	return out
}
func TestRejectsInsufficientCandles(t *testing.T) {
	_, err := NewAnalyzer(DefaultConfig()).Analyze("TEST", candles(20))
	if !errors.Is(err, ErrInsufficientData) {
		t.Fatalf("got %v", err)
	}
}
func TestAnalysisIncludesRangeAndEvidence(t *testing.T) {
	got, err := NewAnalyzer(DefaultConfig()).Analyze("TEST", candles(90))
	if err != nil {
		t.Fatal(err)
	}
	if got.DataPoints != 90 || got.TradingRange == nil || len(got.Evidence) == 0 {
		t.Fatalf("incomplete result: %+v", got)
	}
}
func TestRejectsInvalidCandle(t *testing.T) {
	cs := candles(90)
	cs[4].High = 1
	_, err := NewAnalyzer(DefaultConfig()).Analyze("TEST", cs)
	if err == nil {
		t.Fatal("expected invalid candle error")
	}
}
func TestUnknownPhaseWhenNoEvidence(t *testing.T) {
	cs := candles(90)
	got, err := NewAnalyzer(DefaultConfig()).Analyze("TEST", cs)
	if err != nil {
		t.Fatal(err)
	}
	if got.Phase != PhaseUnknown && got.Phase != PhaseB {
		t.Fatalf("unexpected phase %s", got.Phase)
	}
}

func TestEventsAreScopedToRecentCycle(t *testing.T) {
	cs := candles(180)
	// Create an old high-volume breakout/retest pattern that should not be
	// reported as a current-cycle event once it falls outside EventLookback.
	for i := 0; i < 45; i++ {
		cs[i].Close = 100 + float64(i)*0.8
		cs[i].Open = cs[i].Close - 0.2
		cs[i].Low = cs[i].Close - 0.5
		cs[i].High = cs[i].Close + 0.5
		cs[i].Volume = 1000
	}
	got, err := NewAnalyzer(DefaultConfig()).Analyze("TEST", cs)
	if err != nil {
		t.Fatal(err)
	}
	cutoff := cs[len(cs)-EventLookback].Date
	for _, event := range got.Events {
		if event.Date.Before(cutoff) {
			t.Fatalf("stale event escaped recent-cycle filter: %s on %s before %s", event.Type, event.Date.Format("2006-01-02"), cutoff.Format("2006-01-02"))
		}
	}
}

func TestUnclearStructureCannotBeAssignedPhaseDFromEventsAlone(t *testing.T) {
	cs := candles(90)
	got, err := NewAnalyzer(DefaultConfig()).Analyze("TEST", cs)
	if err != nil {
		t.Fatal(err)
	}
	if got.Structure == StructureUnclear && got.Phase != PhaseUnknown {
		t.Fatalf("unclear structure must remain UNKNOWN, got phase %s", got.Phase)
	}
}

func TestTradingRangeDetectorRequiresStableRange(t *testing.T) {
	cs := candles(140)
	// Establish a gradual decline, followed by a bounded sideways range.
	for i := 0; i < 70; i++ {
		v := 140 - float64(i)*0.62
		cs[i] = Candle{Date: cs[i].Date, Open: v + 0.2, High: v + 1, Low: v - 1, Close: v, Volume: 2000}
	}
	for i := 70; i < len(cs); i++ {
		v := 96 + float64((i%8)-4)*0.45
		cs[i] = Candle{Date: cs[i].Date, Open: v - 0.1, High: v + 1.2, Low: v - 1.2, Close: v, Volume: 1000}
	}
	a := NewAnalyzer(DefaultConfig())
	rc := a.detectTradingRange(cs, -0.1)
	if !rc.valid {
		t.Fatalf("expected stable candidate range, got %+v", rc)
	}
	if rc.preTrend >= -0.08 {
		t.Fatalf("expected preceding markdown context, got %.3f", rc.preTrend)
	}
}

func TestTradingRangeDetectorRejectsStrongTrend(t *testing.T) {
	cs := candles(100)
	for i := range cs {
		v := 50 + float64(i)*0.8
		cs[i] = Candle{Date: cs[i].Date, Open: v - 0.1, High: v + 0.3, Low: v - 0.3, Close: v, Volume: 1000}
	}
	rc := NewAnalyzer(DefaultConfig()).detectTradingRange(cs, 0.21)
	if rc.valid {
		t.Fatalf("strong trend should not be classified as a stable trading range: %+v", rc)
	}
}

func TestTradingRangeUsesFullActiveCycleAndRetainsSupportTouches(t *testing.T) {
	cs := candles(150)
	// A markdown transitions into a wide but bounded range. The range should
	// include the SC low and later resistance tests, not just the last 40 bars.
	for i := range cs {
		v := 55.0 - float64(i)*0.20
		if i >= 75 {
			v = 41.0 + float64((i%7)-3)*0.25
		}
		cs[i] = Candle{Date: time.Date(2026, 6, 1+i, 0, 0, 0, 0, time.UTC), Open: v - 0.1, High: v + 1.0, Low: v - 1.0, Close: v, Volume: 1000}
	}
	// The SC candle establishes the range low after a short, sharp markdown.
	for i := 62; i < 82; i++ {
		v := 44.0 - float64(i-62)*0.28
		cs[i].Open, cs[i].High, cs[i].Low, cs[i].Close = v+0.15, v+0.8, v-0.8, v
		cs[i].Volume = 1000
	}
	// Keep the pre-range reference bar elevated enough to represent a clear markdown.
	cs[65] = Candle{Date: cs[65].Date, Open: 44.7, High: 45.4, Low: 44.2, Close: 45.0, Volume: 1000}
	cs[82] = Candle{Date: cs[82].Date, Open: 38.8, High: 39.4, Low: 36.5, Close: 38.6, Volume: 3000}
	cs[95] = Candle{Date: cs[95].Date, Open: 44.0, High: 45.5, Low: 43.8, Close: 45.0, Volume: 1400}
	cs[110] = Candle{Date: cs[110].Date, Open: 44.0, High: 45.5, Low: 43.8, Close: 45.0, Volume: 1200}
	cs[130] = Candle{Date: cs[130].Date, Open: 37.0, High: 38.0, Low: 36.5, Close: 37.6, Volume: 500}
	cs[137] = Candle{Date: cs[137].Date, Open: 37.0, High: 38.0, Low: 36.6, Close: 37.7, Volume: 500}

	a := NewAnalyzer(DefaultConfig())
	rc := a.detectTradingRange(cs, 0.02)
	if !rc.valid {
		t.Fatalf("expected active range candidate, got %+v", rc)
	}
	if math.Abs(rc.support-36.5) > 0.001 || math.Abs(rc.resistance-45.5) > 0.001 {
		t.Fatalf("expected range [36.5, 45.5], got [%.2f, %.2f]", rc.support, rc.resistance)
	}
}

func TestEvidenceEventsUseCanonicalWyckoffOrder(t *testing.T) {
	date := func(day int) time.Time {
		return time.Date(2026, time.July, day, 0, 0, 0, 0, time.UTC)
	}
	input := []Event{
		{Type: "SOS", Date: date(20)},
		{Type: "ST", Date: date(15)},
		{Type: "PS", Date: date(2)},
		{Type: "LPS", Date: date(22)},
		{Type: "SC", Date: date(5)},
		{Type: "SPRING", Date: date(18)},
		{Type: "AR", Date: date(10)},
		{Type: "ST", Date: date(12)},
		{Type: "SOW", Date: date(23)},
	}
	got := orderEventsForEvidence(input)
	wantTypes := []string{"PS", "SC", "AR", "ST", "ST", "SPRING", "SOS", "LPS", "SOW"}
	if len(got) != len(wantTypes) {
		t.Fatalf("got %d evidence events, want %d", len(got), len(wantTypes))
	}
	for i, want := range wantTypes {
		if got[i].Type != want {
			t.Fatalf("evidence event %d = %s, want %s; all types: %+v", i, got[i].Type, want, got)
		}
	}
	if !got[3].Date.Before(got[4].Date) {
		t.Fatalf("same-type events must be chronological: %s before %s", got[3].Date, got[4].Date)
	}
	// Sorting for Evidence must not mutate the chronological Events source slice.
	if input[0].Type != "SOS" || input[0].Date != date(20) {
		t.Fatal("orderEventsForEvidence mutated its input")
	}
}

func TestAnalysisEventsRemainChronological(t *testing.T) {
	got, err := NewAnalyzer(DefaultConfig()).Analyze("TEST", candles(90))
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i < len(got.Events); i++ {
		if got.Events[i].Date.Before(got.Events[i-1].Date) {
			t.Fatalf("Events must remain chronological: %s appears before %s", got.Events[i-1].Date, got.Events[i].Date)
		}
	}
}
