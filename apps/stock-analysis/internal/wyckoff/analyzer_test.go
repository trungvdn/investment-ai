package wyckoff

import (
	"errors"
	"testing"
	"time"
)
func candles(n int)[]Candle{out:=make([]Candle,n);for i:=range out{v:=100+float64(i%5)*0.3;out[i]=Candle{Date:time.Date(2025,1,1+i,0,0,0,0,time.UTC),Open:v-0.2,High:v+0.8,Low:v-0.8,Close:v,Volume:1000}};return out}
func TestRejectsInsufficientCandles(t *testing.T){_,err:=NewAnalyzer(DefaultConfig()).Analyze("TEST",candles(20));if !errors.Is(err,ErrInsufficientData){t.Fatalf("got %v",err)}}
func TestAnalysisIncludesRangeAndEvidence(t *testing.T){got,err:=NewAnalyzer(DefaultConfig()).Analyze("TEST",candles(90));if err!=nil{t.Fatal(err)};if got.DataPoints!=90||got.TradingRange==nil||len(got.Evidence)==0{t.Fatalf("incomplete result: %+v",got)}}
func TestRejectsInvalidCandle(t *testing.T){cs:=candles(90);cs[4].High=1;_,err:=NewAnalyzer(DefaultConfig()).Analyze("TEST",cs);if err==nil{t.Fatal("expected invalid candle error")}}
func TestUnknownPhaseWhenNoEvidence(t *testing.T){cs:=candles(90);got,err:=NewAnalyzer(DefaultConfig()).Analyze("TEST",cs);if err!=nil{t.Fatal(err)};if got.Phase!=PhaseUnknown&&got.Phase!=PhaseB{t.Fatalf("unexpected phase %s",got.Phase)}}

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
	cs := candles(100)
	// Establish a decline, followed by a bounded sideways range.
	for i := 0; i < 35; i++ {
		v := 130 - float64(i)
		cs[i] = Candle{Date: cs[i].Date, Open: v + 0.2, High: v + 1, Low: v - 1, Close: v, Volume: 2000}
	}
	for i := 35; i < len(cs); i++ {
		v := 96 + float64((i%8)-4)*0.25
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
	rc := NewAnalyzer(DefaultConfig()).detectTradingRange(cs, 0.2)
	if rc.valid {
		t.Fatalf("strong trend should not be classified as a stable trading range: %+v", rc)
	}
}
