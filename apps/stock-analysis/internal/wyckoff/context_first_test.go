package wyckoff

import (
 "testing"
 "time"
)
func campaignFixture(mirror bool)[]Candle {
 cs:=make([]Candle,160)
 for i:=range cs {
  v:=140-float64(i)*.65
  if i>=60 {v=94+float64(i-60)*1.1}
  if i>=76 {v=100+float64((i/3)%2)*7}
  if i>=120 {v=116+float64(i-120)*.25+float64((i/4)%2)*3}
  cs[i]=Candle{Date:time.Date(2026,1,1+i,0,0,0,0,time.UTC),Open:v-.5,High:v+1,Low:v-1,Close:v,Volume:500}
 }
 set:=func(i int,o,h,l,c,v float64){cs[i]=Candle{Date:cs[i].Date,Open:o,High:h,Low:l,Close:c,Volume:v}}
 set(60,96,98,88,92,5000)
 set(75,108,112,107,110,1200)
 set(82,94,97,89,95,600);set(83,95,104,94,103,800)
 set(101,92,93,87,91,700);set(102,92,98,91,97,800)
 set(105,93,96,89,95,250);set(106,95,102,94,101,500)
 set(119,107,110,106,109,500)
 set(120,110,115,109,114,2200);set(121,114,117,113,116,1800);set(122,116,119,115,118,1800)
 set(125,118,119,117,118,700);set(126,114,115,112,113,300);set(127,113,117,112,116,700)
 if mirror {for i,c:=range cs {cs[i].Open,cs[i].High,cs[i].Low,cs[i].Close=200-c.Open,200-c.Low,200-c.High,200-c.Close}}
 return cs
}
func TestContextFirstWindowAndSwings(t *testing.T) {
 for _,mirror:=range []bool{false,true} {
  cs:=campaignFixture(mirror);got,err:=NewAnalyzer(DefaultConfig()).Analyze("SYNTH",cs);if err!=nil {t.Fatal(err)}
  t.Logf("CONTEXT_SYNTH mirror=%t structure=%s phase=%s stage=%s pattern=%s",mirror,got.Structure,got.Phase,got.Context.Stage,got.PriceStructure.Pattern)
  for _,e:=range got.Events {t.Logf("CONTEXT_EVENT %s %s",e.Type,e.Date.Format("2006-01-02"))}
  if got.Window.BaseBars!=40||got.Window.UsedBars<40 {t.Fatal("assessment must start from >=40 bars")}
  for _,s:=range got.PriceStructure.Swings {if s.ConfirmedAt.After(got.AsOf) {t.Fatal("future swing confirmation")}}
  if got.Context.Stage==""||got.ActiveStructureID=="" {t.Fatal("missing forming stage or campaign identity")}
 }
 if _,err:=NewAnalyzer(DefaultConfig()).Analyze("FORTY",campaignFixture(false)[:40]);err!=nil {t.Fatal("forty observations must be accepted")}
}
func TestContextFirstSZCCampaign(t *testing.T) {
 _,cs:=readReviewed(t,"../../data/annotations/SZC/annotation.json")
 got,err:=NewAnalyzer(DefaultConfig()).Analyze("SZC",cs);if err!=nil {t.Fatal(err)}
 t.Logf("CONTEXT_SZC %s/%s %s range=%.3f-%.3f used=%d pattern=%s",got.Structure,got.Phase,got.Context.Stage,got.TradingRange.Support,got.TradingRange.Resistance,got.Window.UsedBars,got.PriceStructure.Pattern)
 for _,e:=range got.Events {t.Logf("CONTEXT_SZC_EVENT %s %s",e.Type,e.Date.Format("2006-01-02"))}
 if got.Window.BaseBars<40||got.Window.UsedBars<40 {t.Fatal("missing minimum context")}
 for _,e:=range got.Events {if e.Type=="BC"&&e.Date.Format("2006-01-02")>="2026-10-20" {t.Fatal("one October rally high must not invent a new distribution BC")}}
}
