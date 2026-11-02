package wyckoff

import "testing"

func TestPresentationReviewedCharts(t *testing.T) {
 for _,symbol:=range []string{"MSB","BCM"} {
  _,cs:=readReviewed(t,"../../data/annotations/"+symbol+"/annotation.json")
  got,err:=NewAnalyzer(DefaultConfig()).Analyze(symbol,cs);if err!=nil {t.Fatal(err)}
  seen:=map[string]bool{};counts:=map[string]int{}
  for _,e:=range got.PresentationEvents {
   key:=e.Type+e.Date.String();if seen[key] {t.Fatalf("duplicate %s %s",symbol,key)};seen[key]=true;counts[e.Type]++
   t.Logf("PRESENTATION %s %s %s scope=%s",symbol,e.Type,e.Date.Format("2006-01-02"),e.StructureID)
   if e.StructureID!=""&&e.StructureID!=got.PresentationStructureID {t.Fatal("event from another structure leaked onto primary chart")}
   if e.Type=="SPRING" {for _,c:=range cs {if c.Date.Equal(e.Date)&&c.Low>=got.TradingRange.Support {t.Fatal("Spring did not penetrate displayed support")}}}
  }
  if counts["SOS"]>1||counts["LPS"]>1 {t.Fatalf("multiple SOS/LPS on one range: %v",counts)}
  if symbol=="MSB" {
   expected:=map[string]string{"SPRING":"2026-08-26","SOS":"2026-09-23","LPS":"2026-10-05"}
   for kind,date:=range expected {found:=false;for _,e:=range got.PresentationEvents {if e.Type==kind&&e.Date.Format("2006-01-02")==date {found=true}};if !found {t.Fatalf("MSB expected %s on %s",kind,date)}}
   if counts["TEST"]!=0 {t.Fatal("MSB uncertain internal tests must not be promoted on the chart")}
  }
  if symbol=="BCM" {
   if counts["SPRING"]!=0 {t.Fatal("BCM must not show a Spring of a smaller range")}
   found:=false;for _,e:=range got.PresentationEvents {if e.Type=="SC"&&e.Date.Format("2006-01-02")=="2026-07-27" {found=true}}
   if !found {t.Fatal("BCM SC should be strongest July selloff, not a later minor down day")}
  }
 }
}
