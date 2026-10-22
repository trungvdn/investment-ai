package wyckoff

import (
	"errors"
	"testing"
	"time"
)

func sampleCandles(n int) []Candle {
	out:=make([]Candle,n)
	for i:=range out {
		close:=100.0+float64(i%5)*0.3
		out[i]=Candle{Date:time.Date(2026,1,1+i,0,0,0,0,time.UTC),Open:close-0.2,High:close+0.8,Low:close-0.8,Close:close,Volume:1000}
	}
	return out
}

func TestAnalyzeRejectsInsufficientCandles(t *testing.T) {
	_,err:=NewAnalyzer(DefaultConfig()).Analyze("TEST",sampleCandles(10))
	if !errors.Is(err,ErrInsufficientData) { t.Fatalf("expected ErrInsufficientData, got %v",err) }
}

func TestAnalyzeReturnsEvidenceAndRange(t *testing.T) {
	got,err:=NewAnalyzer(DefaultConfig()).Analyze("TEST",sampleCandles(40))
	if err!=nil { t.Fatal(err) }
	if got.Symbol!="TEST" || got.DataPoints!=40 { t.Fatalf("unexpected result: %+v",got) }
	if got.TradingRange==nil || got.TradingRange.Support<=0 || got.TradingRange.Resistance<got.TradingRange.Support { t.Fatalf("invalid range: %+v",got.TradingRange) }
	if len(got.Evidence)==0 { t.Fatal("expected evidence") }
}

func TestAnalyzeRejectsInvalidCandle(t *testing.T) {
	c:=sampleCandles(40)
	c[4].High=1
	_,err:=NewAnalyzer(DefaultConfig()).Analyze("TEST",c)
	if err==nil { t.Fatal("expected invalid candle error") }
}
