package wyckoff

import (
 "testing"
 "time"
)
func lifecycleCandles() []Candle {
 cs:=make([]Candle,100)
 for i:=range cs {
  v:=70+float64(i)*.5
  if i>=40 {v=105;if i%4<2 {v=101}}
  cs[i]=Candle{Date:time.Date(2026,1,1+i,0,0,0,0,time.UTC),Open:v,High:v+1,Low:v-1,Close:v,Volume:100}
 }
 return cs
}
func TestLifecycleFrozenBoundsAndFailure(t *testing.T) {
 cs:=lifecycleCandles()
 before:=trackStructures("TEST",cs)
 if len(before)==0 {t.Fatal("expected range")}
 original:=before[0]
 appendBar:=func(close,low,volume float64){date:=cs[len(cs)-1].Date.AddDate(0,0,1);cs=append(cs,Candle{Date:date,Open:close-.2,High:close+.2,Low:low,Close:close,Volume:volume})}
 appendBar(108,106,200)
 for i:=0;i<6;i++ {appendBar(109,108,120)}
 after:=trackStructures("TEST",cs)
 var retained StructureState
 for _,s:=range after {if s.ID==original.ID {retained=s}}
 if retained.Support!=original.Support||retained.Resistance!=original.Resistance||retained.Status!="MARKUP" {t.Fatalf("range overwritten / markup missing: %+v",retained)}
 appendBar(106.5,106,60)
 lpsFound:=false
 for _,s:=range trackStructures("TEST",cs) {if s.ID==original.ID {for _,e:=range s.Events {if e.Type=="LPS"&&e.StructureID==s.ID {lpsFound=true}}}}
 if !lpsFound {t.Fatal("pullback must reference retained breakout resistance")}
 appendBar(104,103,120);appendBar(103,102,120)
 failed:=trackStructures("TEST",cs)
 for _,s:=range failed {if s.ID==original.ID {if s.Status!="FAILED_BREAKOUT"||len(s.Events)==0||s.Events[0].Type!="SOS" {t.Fatalf("failure must preserve original SOS: %+v",s)}}}
 // Prefix reconstruction must not include the later breakout or failure.
 replay:=trackStructures("TEST",cs[:100])
 for _,s:=range replay {for _,tr:=range s.Transitions {if tr.ObservedAt.After(cs[99].Date) {t.Fatal("future transition")}};if s.BreakoutDate!=nil {t.Fatal("prefix saw future breakout")}}
}
func TestLifecycleReviewedMSBAndNLG(t *testing.T) {
 for _,symbol:=range []string{"MSB","NLG"} {
  _,cs:=readReviewed(t,"../../data/annotations/"+symbol+"/annotation.json")
  result,err:=NewAnalyzer(DefaultConfig()).Analyze(symbol,cs);if err!=nil {t.Fatal(err)}
  for _,s:=range result.Structures {t.Logf("LIFECYCLE %s id=%s parent=%s window=%d kind=%s start=%s identified=%s support=%.3f resistance=%.3f status=%s",symbol,s.ID,s.ParentID,s.Window,s.Kind,s.StartDate.Format("2006-01-02"),s.IdentifiedAt.Format("2006-01-02"),s.Support,s.Resistance,s.Status)}
  if symbol=="MSB"&& (result.Structure!=StructureMarkup||result.Phase!=PhaseE||result.ActiveStructureID=="") {t.Fatalf("MSB should retain markup after reaccumulation, got %s/%s",result.Structure,result.Phase)}
  if symbol=="NLG" {
   nested:=false
   for _,s:=range result.Structures {if s.ParentID!=""&&s.Support>=20.8&&s.Support<=21.6&&s.Resistance>=22.2&&s.Resistance<=23 {nested=true}}
   if !nested {t.Fatal("NLG short range must retain an enclosing context")}
  }
 }
}
