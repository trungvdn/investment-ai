package wyckoff

import (
	"testing"
	"time"
)

func TestSpringRecoveryDelayedAndInvalidation(t *testing.T) {
	cs := make([]Candle, 30)
	for i := range cs {
		cs[i] = Candle{Date: time.Date(2026, 1, 1+i, 0, 0, 0, 0, time.UTC), Open: 105, High: 106, Low: 104, Close: 105, Volume: 100}
	}
	appendBar := func(open, high, low, close, volume float64) {
		cs = append(cs, Candle{Date: cs[len(cs)-1].Date.AddDate(0, 0, 1), Open: open, High: high, Low: low, Close: close, Volume: volume})
	}
	s := StructureState{ID: "range", Status: "RANGE", Support: 100, Resistance: 110}
	appendBar(101, 102, 98, 99, 180)
	s.advanceSpring(cs, 30)
	if len(s.Events) != 0 || s.SpringAttempts[0].Status != "PENDING" {
		t.Fatal("penetration alone cannot confirm Spring")
	}
	appendBar(99, 103, 98, 103, 240)
	s.advanceSpring(cs, 31)
	if len(s.Events) != 1 || s.Events[0].Type != "SPRING" || !s.Events[0].Date.Equal(cs[30].Date) || s.Events[0].ObservedAt == nil || !s.Events[0].ObservedAt.Equal(cs[31].Date) {
		t.Fatal("delayed recovery must retain event date and later availability")
	}
	appendBar(101, 102, 100, 101.5, 50)
	s.advanceSpring(cs, 32)
	if len(s.Events) != 2 || s.Events[1].Type != "TEST" || s.Events[1].StructureID != s.ID {
		t.Fatal("valid test must retain parent Spring range")
	}
	appendBar(99, 100, 97, 99, 80)
	s.advanceSpring(cs, 33)
	if s.SpringAttempts[0].Status != "INVALIDATED" {
		t.Fatal("lower low after recovery must invalidate")
	}
	count := len(s.Events)
	appendBar(101, 102, 100, 101.5, 40)
	s.advanceSpring(cs, 34)
	if len(s.Events) != count {
		t.Fatal("must not emit new tests after invalidation")
	}
}
func TestSpringRecoveryExpiresWithoutReclaim(t *testing.T) {
	cs := make([]Candle, 35)
	for i := range cs {
		cs[i] = Candle{Date: time.Date(2026, 1, 1+i, 0, 0, 0, 0, time.UTC), Open: 99, High: 100, Low: 98, Close: 99, Volume: 100}
	}
	s := StructureState{ID: "range", Status: "RANGE", Support: 100, Resistance: 110}
	for i := 30; i < 35; i++ {
		s.advanceSpring(cs, i)
	}
	if len(s.Events) != 0 || s.SpringAttempts[0].Status != "EXPIRED" {
		t.Fatal("unrecovered breakdown is not Spring")
	}
}
func TestSpringRecoveryReviewedCutoffs(t *testing.T) {
	for _, tc := range []struct{ symbol, start, end, observed string }{{"MSB", "2026-08-26", "2026-08-27", "2026-08-27"}, {"NLG", "2026-10-14", "2026-10-15", "2026-10-15"}} {
		_, cs := readReviewed(t, "../../data/annotations/"+tc.symbol+"/annotation.json")
		cutoff := 0
		for i, c := range cs {
			if c.Date.Format("2006-01-02") == tc.observed {
				cutoff = i + 1
			}
		}
		if cutoff < 60 {
			t.Fatal("missing cutoff")
		}
		for _, n := range []int{cutoff - 1, cutoff} {
			got, err := NewAnalyzer(DefaultConfig()).Analyze(tc.symbol, cs[:n])
			if err != nil {
				t.Fatal(err)
			}
			hit := false
			for _, e := range got.Events {
				date := e.Date.Format("2006-01-02")
				if e.Type == "SPRING" && e.StructureID != "" && date >= tc.start && date <= tc.end {
					hit = true
					if e.ObservedAt == nil || e.ObservedAt.After(got.AsOf) {
						t.Fatal("invalid availability date")
					}
				}
			}
			if hit != (n == cutoff) {
				t.Fatalf("%s cutoff=%s Spring=%t stage=%s range=%+v events=%+v", tc.symbol, cs[n-1].Date.Format("2006-01-02"), hit, got.Context.Stage, got.TradingRange, got.Events)
			}
		}
	}
}

func TestSpringRecoveryRejectedLargeRangeScopes(t *testing.T) {
	for _, tc := range []struct {
		symbol, date string
		support      float64
	}{{"CTD", "2026-09-17", 54.7}, {"HDG", "2026-10-09", 15.2}} {
		_, cs := readReviewed(t, "../../data/annotations/"+tc.symbol+"/annotation.json")
		got, err := NewAnalyzer(DefaultConfig()).Analyze(tc.symbol, cs)
		if err != nil {
			t.Fatal(err)
		}
		for _, s := range got.Structures {
			for _, e := range s.Events {
				if e.Type != "SPRING" || e.Date.Format("2006-01-02") != tc.date {
					continue
				}
				t.Logf("SPRING_SCOPE %s %s id=%s support=%.3f resistance=%.3f identified=%s observed=%v", tc.symbol, tc.date, s.ID, s.Support, s.Resistance, s.IdentifiedAt.Format("2006-01-02"), e.ObservedAt)
				if s.Support >= tc.support-.001 && s.Support <= tc.support+.001 {
					t.Fatalf("rejected Spring of original large range was emitted: %s %s", tc.symbol, s.ID)
				}
			}
		}
	}
}
