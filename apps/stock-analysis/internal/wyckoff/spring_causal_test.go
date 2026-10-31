package wyckoff

import (
	"testing"
	"time"
)

func TestSpringUsesRangeBeforeEvent(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	candles := make([]Candle, 100)
	for i := range candles {
		closePrice := 101.0
		if i%2 == 0 {
			closePrice = 109
		}
		if i >= 80 {
			closePrice = 109 - float64(i-80)*0.4
		}
		candles[i] = Candle{Date: start.AddDate(0, 0, i), Open: closePrice, High: 110, Low: 100, Close: closePrice, Volume: 100}
	}
	spring := Candle{Date: start.AddDate(0, 0, 100), Open: 100, High: 103, Low: 99, Close: 102, Volume: 80}
	analyze := func(last Candle) Analysis {
		input := append(append([]Candle(nil), candles...), last)
		result, err := NewAnalyzer(DefaultConfig()).Analyze("SYNTHETIC", input)
		if err != nil {
			t.Fatal(err)
		}
		return result
	}
	hasSpringAt := func(result Analysis) bool {
		for _, event := range result.Events {
			if event.Type == "SPRING" && event.Date.Equal(spring.Date) {
				return true
			}
		}
		return false
	}
	if !hasSpringAt(analyze(spring)) {
		t.Fatal("expected Spring below previously established support")
	}
	noPenetration := spring
	noPenetration.Low = 100
	if hasSpringAt(analyze(noPenetration)) {
		t.Fatal("support touch is not a Spring")
	}
	noRecovery := spring
	noRecovery.Close = 99.5
	if hasSpringAt(analyze(noRecovery)) {
		t.Fatal("breakdown without recovery is not a Spring")
	}
}
