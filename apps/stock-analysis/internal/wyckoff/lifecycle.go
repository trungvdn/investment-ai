package wyckoff

import (
 "fmt"
 "math"
 "sort"
 "strings"
 "time"
)

// StructureState retains the original bounds after price leaves a range.
// IdentifiedAt is the availability date, distinct from historical StartDate.
// IDs are deterministic for identical input history; revised/truncated feeds may change them.
type StructureState struct {
 ID string `json:"structure_id"`
 ParentID string `json:"parent_structure_id,omitempty"`
 Kind string `json:"kind"`
 StartDate time.Time `json:"start_date"`
 RangeEnd time.Time `json:"range_end"`
 IdentifiedAt time.Time `json:"identified_at"`
 LastObserved time.Time `json:"last_observed"`
 Support float64 `json:"support"`
 Resistance float64 `json:"resistance"`
 Status string `json:"status"`
 BreakoutDate *time.Time `json:"breakout_date,omitempty"`
 Transitions []StructureTransition `json:"transitions"`
 Events []Event `json:"events"`
 Window int `json:"window"`
 preTrend float64
 breakoutIndex int
 upCount,downCount int
}
type StructureTransition struct {
 Status string `json:"status"`
 ObservedAt time.Time `json:"observed_at"`
 Evidence string `json:"evidence"`
}
func (s *StructureState) transition(status string, date time.Time, evidence string) {
 if s.Status==status {return}
 s.Status=status
 s.Transitions=append(s.Transitions,StructureTransition{status,date,evidence})
}
func (s *StructureState) event(kind string, date time.Time, evidence string) {
 for _,e:=range s.Events {if e.Type==kind {return}}
 s.Events=append(s.Events,Event{Type:kind,Date:date,Confidence:0.6,Evidence:evidence,StructureID:s.ID})
}
// lifecycleRange uses only the current prefix. Multiple scales allow an inner
// consolidation to coexist with its larger context. No annotations enter detection.
func lifecycleRange(cs []Candle,end,w int) (float64,float64,float64,bool) {
 start:=end-w+1
 if start<20 {return 0,0,0,false}
 window:=cs[start:end+1]
 lo,hi:=bounds(window);width:=hi-lo
 pct:=width/math.Max((hi+lo)/2,0.00001)
 if pct<0.04||pct>0.35 {return 0,0,0,false}
 inside,lt,ht:=0,0,0
 firstL,lastL,firstH,lastH:=-1,-1,-1,-1
 for j,c:=range window {
  if c.Close>=lo+width*.1&&c.Close<=hi-width*.1 {inside++}
  if c.Low<=lo+width*.12 {lt++;if firstL<0 {firstL=j};lastL=j}
  if c.High>=hi-width*.12 {ht++;if firstH<0 {firstH=j};lastH=j}
 }
 if float64(inside)/float64(w)<.45||lt<2||ht<2||lastL-firstL<3||lastH-firstH<3 {return 0,0,0,false}
 // Both ends of a directional move alone must not manufacture a range.
 drift:=math.Abs(window[w-1].Close-window[0].Close)/width
 if drift>.7 {return 0,0,0,false}
 prior:=maxInt(0,start-40)
 trend:=(cs[start].Close-cs[prior].Close)/cs[prior].Close
 return lo,hi,trend,true
}
func sameLifecycleRange(s StructureState,lo,hi float64,start time.Time) bool {
 width:=s.Resistance-s.Support
 return !start.After(s.RangeEnd)&&math.Abs(s.Support-lo)<=width*.22&&math.Abs(s.Resistance-hi)<=width*.22
}
func trackStructures(symbol string,cs []Candle) []StructureState {
 states:=[]StructureState{}
 for i:=59;i<len(cs);i++ {
  c:=cs[i]
  for j:=range states {
   s:=&states[j];s.LastObserved=c.Date
   width:=s.Resistance-s.Support
   if c.Close>s.Resistance {s.upCount++} else {s.upCount=0}
   if c.Close<s.Support {s.downCount++} else {s.downCount=0}
   st:=stats(c,avgVolume(cs[maxInt(0,i-20):i]))
   if s.BreakoutDate==nil {
    if c.Close>s.Resistance&&st.closeLocation>=.65&&st.relativeVolume>=1.2 {
     d:=c.Date;s.BreakoutDate=&d;s.breakoutIndex=i;if s.Status=="RANGE" {s.RangeEnd=cs[i-1].Date}
     s.transition("BREAKOUT_UP",c.Date,"Close above frozen resistance with expanding volume; follow-through remains required.")
     s.event("SOS",c.Date,"Breakout of this structure's original resistance; candidate strength, not a guarantee of success.")
    } else if s.downCount>=2 {
     if s.Status=="RANGE" {s.RangeEnd=cs[i-2].Date}
     s.transition("BREAKDOWN",c.Date,"Two consecutive closes below original support; no distribution/SOW label inferred from price loss alone.")
    } else if s.Status=="RANGE" {s.RangeEnd=c.Date}
   } else {
    // Do not rewrite the original SOS when a later breakout fails.
    if s.downCount>=2 {
     s.transition("MARKDOWN",c.Date,"Original range support lost after an upside breakout; failed accumulation hypothesis.")
    } else if c.Close<s.Resistance&&cs[i-1].Close<s.Resistance {
     if s.Status!="MARKDOWN" {s.transition("FAILED_BREAKOUT",c.Date,"Two closes back inside the original range after breakout.")}
    } else if s.Status=="BREAKOUT_UP"&&i-s.breakoutIndex>=5&&s.upCount>=3&&c.Close>s.Resistance+width*.35 {
     s.transition("MARKUP",c.Date,"Sustained follow-through above the frozen range; prior trend distinguishes reaccumulation.")
    }
    if (s.Status=="BREAKOUT_UP"||s.Status=="MARKUP")&&i>s.breakoutIndex&&i-s.breakoutIndex<=15&&c.Close<cs[i-1].Close&&c.Low>=s.Resistance-width*.15&&c.Low<=s.Resistance+width*.35&&c.Close>=s.Resistance&&st.relativeVolume<1 {
     s.event("LPS",c.Date,"Actual pullback held this structure's original breakout level on reduced volume; candidate LPS.")
    }
   }
  }
  for _,w:=range []int{12,40,70} {
   lo,hi,trend,ok:=lifecycleRange(cs,i,w);if !ok {continue}
   start:=cs[i-w+1].Date
   duplicate:=false
   for _,s:=range states {if sameLifecycleRange(s,lo,hi,start) {duplicate=true;break}}
   if duplicate {continue}
   kind:="unclear";if trend>.08 {kind="reaccumulation"} else if trend < -0.08 {kind="accumulation"}
   s:=StructureState{ID:fmt.Sprintf("%s_%d_%s_%s",strings.ToUpper(symbol),w,start.Format("20060102"),c.Date.Format("20060102")),Kind:kind,StartDate:start,RangeEnd:c.Date,IdentifiedAt:c.Date,LastObserved:c.Date,Support:lo,Resistance:hi,Window:w,preTrend:trend,Events:[]Event{},Transitions:[]StructureTransition{}}
   s.transition("RANGE",c.Date,"Repeated reactions and overlapping closes established a candidate range on this input prefix.")
   states=append(states,s)
  }
 }
 // Parent attribution is a current-as-of relationship, not a historical claim
 // that a larger range had already been recognized at the child's inception.
 for i:=range states {
  best:=math.MaxFloat64
  for j:=range states {
   a,b:=states[i],states[j]
   if i==j||b.StartDate.After(a.StartDate)||b.IdentifiedAt.After(a.LastObserved) {continue}
   aw,bw:=a.Resistance-a.Support,b.Resistance-b.Support
   if bw>aw*1.5&&b.Support<=a.Support&&b.Resistance>=a.Resistance&&bw<best {states[i].ParentID=b.ID;best=bw}
  }
 }
 return states
}
func applyStructureLifecycle(result *Analysis,cs []Candle) {
 result.Structures=trackStructures(result.Symbol,cs)
 if len(result.Structures)==0 {return}
 // Preserve all historical structures, select a recent one for the headline.
 // A mature markup range outranks a newly inferred provisional range only
 // while price still holds above its original breakout resistance.
 active:=-1
 last:=cs[len(cs)-1]
 for i,s:=range result.Structures {
  if s.Status!="MARKUP"||last.Close<=s.Resistance||s.BreakoutDate==nil||s.Kind!="reaccumulation" {continue}
  if active<0||s.BreakoutDate.After(*result.Structures[active].BreakoutDate) {active=i}
 }
 if active>=0 {
  s:=result.Structures[active]
  result.ActiveStructureID=s.ID
  result.Structure=StructureMarkup;result.Phase=PhaseE
  result.TradingRange=&TradingRange{Support:s.Support,Resistance:s.Resistance,WidthPct:100*(s.Resistance-s.Support)/((s.Resistance+s.Support)/2)}
  result.Confidence=.60
  result.Evidence=append(result.Evidence,"Markup follows a retained reaccumulation range; its original breakout level remains available for pullback/failure analysis.")
 }
 // Lifecycle events carry attribution; legacy unscoped events remain available
 // for compatibility, but are not silently assigned to an arbitrary range.
 if active<0 {
  for i,s:=range result.Structures {
   if last.Close<s.Support||last.Close>s.Resistance {continue}
   if active<0||s.IdentifiedAt.After(result.Structures[active].IdentifiedAt) {active=i}
  }
  if active>=0 {result.ActiveStructureID=result.Structures[active].ID}
 }
 for _,s:=range result.Structures {for _,e:=range s.Events {
  if e.Date.Before(cs[maxInt(0,len(cs)-EventLookback)].Date) {continue}
  replaced:=false
  for j,old:=range result.Events {if old.Type==e.Type&&old.Date.Equal(e.Date)&&old.StructureID=="" {result.Events[j]=e;replaced=true;break}}
  if !replaced {result.Events=append(result.Events,e)}
 }}
 sort.SliceStable(result.Events,func(i,j int)bool{return result.Events[i].Date.Before(result.Events[j].Date)})
}
