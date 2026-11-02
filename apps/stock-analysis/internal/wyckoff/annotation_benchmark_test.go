package wyckoff

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

type reviewedEvent struct {
	ID        string `json:"event_id"`
	Structure string `json:"structure_id"`
	Proposed  string `json:"proposed_label"`
	Final     string `json:"final_label"`
	Start     string `json:"start_date"`
	End       string `json:"end_date"`
	Verdict   string `json:"label_verdict"`
	Status    string `json:"event_status"`
	Confirmed string `json:"confirmed_as_of"`
}
type reviewedSample struct {
	Ticker string `json:"ticker"`
	Cutoff string `json:"analysis_as_of"`
	Data   struct {
		Path string `json:"ohlcv_path"`
	} `json:"data"`
	Review struct {
		Status   string `json:"status"`
		Eligible bool   `json:"eligible_for_ai"`
	} `json:"review"`
	Assessment struct {
		Structure string `json:"structure"`
		Phase     string `json:"phase"`
	} `json:"assessment"`
	Ranges []reviewedRange `json:"trading_ranges"`
	Events []reviewedEvent `json:"events"`
}

func readReviewed(t *testing.T, path string) (reviewedSample, []Candle) {
	t.Helper()
	var s reviewedSample
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &s); err != nil {
		t.Fatal(err)
	}
	if s.Review.Status != "approved" || !s.Review.Eligible {
		t.Fatalf("unapproved sample %s", path)
	}
	if filepath.Base(s.Data.Path) != s.Data.Path {
		t.Fatal("OHLCV must be colocated")
	}
	f, err := os.Open(filepath.Join(filepath.Dir(path), s.Data.Path))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	var cs []Candle
	for _, r := range rows[1:] {
		if len(r) != 7 {
			t.Fatal("invalid OHLCV row")
		}
		date, err := time.Parse("2006-01-02", r[0])
		if err != nil {
			t.Fatal(err)
		}
		if len(cs) > 0 && !date.After(cs[len(cs)-1].Date) {
			t.Fatal("unordered OHLCV")
		}
		var v [5]float64
		for j := range v {
			v[j], err = strconv.ParseFloat(r[j+1], 64)
			if err != nil {
				t.Fatal(err)
			}
		}
		if r[0] <= s.Cutoff {
			cs = append(cs, Candle{Date: date, Open: v[0], High: v[1], Low: v[2], Close: v[3], Volume: v[4]})
		}
	}
	return s, cs
}
func eventHit(es []Event, label, start, end string) bool {
	for _, e := range es {
		d := e.Date.Format("2006-01-02")
		if e.Type == strings.ToUpper(label) && d >= start && d <= end {
			return true
		}
	}
	return false
}

// An approved candidate is a reviewed hypothesis, not a confirmed market event.
// Empty/unreviewed intervals are never treated as negative labels.
func TestReviewedAnnotationBenchmark(t *testing.T) {
	paths, err := filepath.Glob("../../data/annotations/*/annotation.json")
	if err != nil || len(paths) != 9 {
		t.Fatalf("expected nine samples: %v %d", err, len(paths))
	}
	type row struct {
		Ticker, ID, Structure, Kind, Label, Start, End, FirstSeen string
		Hit, AtCutoff                                             bool
	}
	var out []row
	positive, hits, rejected, falseHits, uncertain, unsupported, ambiguous := 0, 0, 0, 0, 0, 0, 0
	for _, path := range paths {
		s, cs := readReviewed(t, path)
		if len(cs) < 60 {
			t.Fatal("insufficient history")
		}
		results := make([]Analysis, len(cs))
		for n := 60; n <= len(cs); n++ {
			r, err := NewAnalyzer(DefaultConfig()).Analyze(s.Ticker, cs[:n])
			if err != nil {
				t.Fatal(err)
			}
			results[n-1] = r
			for _, e := range r.Events {
				if e.Date.After(cs[n-1].Date) {
					t.Fatal("future event")
				}
			}
		}
		seen := map[string]bool{}
		for _, e := range s.Events {
			if e.End == "" {
				e.End = e.Start
			}
			if e.Start > e.End || e.End > s.Cutoff {
				t.Fatalf("invalid event dates %s/%s", s.Ticker, e.ID)
			}
			kind, label := "positive", e.Final
			if e.Verdict == "uncertain" || e.Status == "uncertain" || e.Final == "" {
				kind = "uncertain"
				uncertain++
			} else if e.Verdict == "incorrect" {
				kind = "rejected"
				label = e.Proposed
			}
			if kind != "uncertain" && !strings.Contains("|PS|SC|AR|ST|SPRING|TEST|SOS|LPS|PSY|BC|UT|UTAD|SOW|LPSY|", "|"+strings.ToUpper(label)+"|") {
				kind = "unsupported"
				unsupported++
			}
			// Raw analyzer events have no structure_id: a same-date SOS in a small
			// range cannot be scored as a false SOS in the enclosing range.
			if kind == "rejected" {
				for _, other := range s.Events {
					if other.Verdict == "correct" && strings.EqualFold(other.Final, label) && other.Structure != e.Structure && other.Start <= e.End && other.End >= e.Start {
						kind = "structure_ambiguous"
						ambiguous++
						break
					}
				}
			}
			key := kind + "|" + e.Structure + "|" + label + "|" + e.Start + "|" + e.End
			if seen[key] {
				continue
			}
			seen[key] = true
			r := row{Ticker: s.Ticker, ID: e.ID, Structure: e.Structure, Kind: kind, Label: label, Start: e.Start, End: e.End}
			for i := 59; i < len(cs); i++ {
				if cs[i].Date.Format("2006-01-02") < e.End {
					continue
				}
				if eventHit(results[i].Events, label, e.Start, e.End) {
					r.Hit = true
					r.FirstSeen = cs[i].Date.Format("2006-01-02")
					break
				}
			}
			r.AtCutoff = eventHit(results[len(cs)-1].Events, label, e.Start, e.End)
			if kind == "positive" {
				positive++
				if r.Hit {
					hits++
				}
			}
			if kind == "rejected" {
				rejected++
				if r.Hit {
					falseHits++
				}
			}
			out = append(out, r)
		}
		r := results[len(cs)-1]
		t.Logf("ASSESSMENT %s expected=%s/%s actual=%s/%s ranges=%d", s.Ticker, s.Assessment.Structure, s.Assessment.Phase, r.Structure, r.Phase, len(s.Ranges))
	}
	report := struct {
		Positive, Hits, Rejected, RejectedEmitted, Uncertain, Unsupported, StructureAmbiguous int
		Events                                                                                []row
	}{positive, hits, rejected, falseHits, uncertain, unsupported, ambiguous, out}
	b, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("BENCHMARK %s", b)
	if path := os.Getenv("WYCKOFF_BENCHMARK_OUTPUT"); path != "" {
		if err := os.WriteFile(path, b, 0644); err != nil {
			t.Fatal(err)
		}
	}
}

// These checks protect the human-reviewed meaning; analyzer disagreements are
// benchmark results, never silently rewritten labels or passing accuracy gates.
func TestReviewedCriticalCases(t *testing.T) {
	for _, symbol := range []string{"MSB", "HDG", "NLG"} {
		s, _ := readReviewed(t, "../../data/annotations/"+symbol+"/annotation.json")
		events := map[string]reviewedEvent{}
		for _, e := range s.Events {
			events[e.ID] = e
		}
		switch symbol {
		case "MSB":
			if s.Assessment.Structure != "markup" || s.Assessment.Phase != "E" || events["E007"].Final != "no_event" {
				t.Fatal("MSB markup/rejected LPS contract changed")
			}
		case "HDG":
			if events["E005"].Final != "markdown_continuation" || events["E005"].Confirmed == "" || events["E004"].Final != "SOS" {
				t.Fatal("failed accumulation must preserve original SOS and subsequent markdown")
			}
		case "NLG":
			if len(s.Ranges) != 2 || events["E005"].Structure == events["E008"].Structure || events["E005"].Final != "SOS" || events["E008"].Final != "no_event" {
				t.Fatal("nested ranges must not share breakout verdict")
			}
		}
	}
}
