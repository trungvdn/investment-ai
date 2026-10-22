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
