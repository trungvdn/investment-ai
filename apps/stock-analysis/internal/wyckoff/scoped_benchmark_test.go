package wyckoff

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

type reviewedZone struct {
	Lower float64 `json:"lower"`
	Upper float64 `json:"upper"`
}
type reviewedRange struct {
	ID         string       `json:"structure_id"`
	Start      string       `json:"start_date"`
	End        string       `json:"end_date"`
	Support    reviewedZone `json:"support_zone"`
	Resistance reviewedZone `json:"resistance_zone"`
}
type scopeMatch struct {
	Target, Status string
	Cost           float64
}

// Fixed geometric/temporal policy shared by every symbol; no event-label hints,
// ticker aliases or manual equivalence IDs enter the matching function.
func matchReviewedScope(s StructureState, targets []reviewedRange, cutoff string) scopeMatch {
	type choice struct {
		id       string
		cost     float64
		reliable bool
	}
	var choices []choice
	date := func(v string) time.Time { d, _ := time.Parse("2006-01-02", v); return d }
	for _, r := range targets {
		end := r.End
		if end == "" {
			end = cutoff
		}
		width := r.Resistance.Upper - r.Support.Lower
		if width <= 0 || r.Support.Lower <= 0 || r.Resistance.Upper <= 0 || r.Start == "" {
			continue
		}
		a, b := s.StartDate, s.IdentifiedAt
		c, d := date(r.Start), date(end)
		if b.Before(a) || d.Before(c) || b.Before(c) || a.After(d) {
			continue
		}
		left, right := a, b
		if c.After(left) {
			left = c
		}
		if d.Before(right) {
			right = d
		}
		overlap := right.Sub(left).Hours()/24 + 1
		shorter := math.Min(b.Sub(a).Hours()/24+1, d.Sub(c).Hours()/24+1)
		if overlap/shorter < .5 {
			continue
		}
		distance := func(v float64, z reviewedZone) float64 {
			if v < z.Lower {
				return z.Lower - v
			}
			if v > z.Upper {
				return v - z.Upper
			}
			return 0
		}
		lo, hi := distance(s.Support, r.Support)/width, distance(s.Resistance, r.Resistance)/width
		if lo > .15 || hi > .15 {
			continue
		}
		choices = append(choices, choice{r.ID, lo + hi, lo <= .05 && hi <= .05})
	}
	if len(choices) == 0 {
		return scopeMatch{Status: "UNMAPPED"}
	}
	sort.Slice(choices, func(i, j int) bool {
		if choices[i].cost == choices[j].cost {
			return choices[i].id < choices[j].id
		}
		return choices[i].cost < choices[j].cost
	})
	if len(choices) > 1 && choices[1].cost-choices[0].cost < .05 {
		return scopeMatch{Status: "AMBIGUOUS", Cost: choices[0].cost}
	}
	if !choices[0].reliable {
		return scopeMatch{Status: "LOW_CONFIDENCE", Cost: choices[0].cost}
	}
	return scopeMatch{Target: choices[0].id, Status: "MAPPED", Cost: choices[0].cost}
}

type scopedPrediction struct {
	Ticker, PredictedID, Type, Date, FirstSeen, Target, Mapping string
	Support, Resistance                                         float64
}
type scopedRow struct {
	Ticker, EventID, Structure, Kind, Label, Start, End, FirstSeen, Disposition string
	Hit                                                                         bool
}
type scopedSummary struct {
	Ticker, Structure                                    string
	Mapped                                               bool
	DefinitionAvailable                                  bool
	Positive, Hits, Rejected, RejectedEmitted, Uncertain int
}

func TestStructureScopedBenchmark(t *testing.T) {
	paths, err := filepath.Glob("../../data/annotations/*/annotation.json")
	if err != nil || len(paths) != 9 {
		t.Fatalf("expected nine approved samples: %v", err)
	}
	var predictions []scopedPrediction
	var rows []scopedRow
	var summaries []scopedSummary
	positive, hits, rejected, rejectedHits, uncertain := 0, 0, 0, 0, 0
	unattributed, unmapped, ambiguous, mapped, lowConfidence := 0, 0, 0, 0, 0
	for _, path := range paths {
		sample, cs := readReviewed(t, path)
		earliest := sample.Cutoff
		summary := map[string]*scopedSummary{}
		for _, r := range sample.Ranges {
			summary[r.ID] = &scopedSummary{Ticker: sample.Ticker, Structure: r.ID, DefinitionAvailable: true}
			if r.Start < earliest {
				earliest = r.Start
			}
		}
		for _, e := range sample.Events {
			if e.Start < earliest {
				earliest = e.Start
			}
		}
		seen := map[string]bool{}
		var local []scopedPrediction
		for n := 60; n <= len(cs); n++ {
			a, err := NewAnalyzer(DefaultConfig()).Analyze(sample.Ticker, cs[:n])
			if err != nil {
				t.Fatal(err)
			}
			byID := map[string]StructureState{}
			mapping := map[string]scopeMatch{}
			for _, s := range a.Structures {
				byID[s.ID] = s
				m := matchReviewedScope(s, sample.Ranges, sample.Cutoff)
				mapping[s.ID] = m
				if m.Status == "MAPPED" {
					summary[m.Target].Mapped = true
				}
			}
			for _, e := range a.Events {
				date := e.Date.Format("2006-01-02")
				if date < earliest {
					continue
				}
				if e.Date.After(a.AsOf) || (e.ObservedAt != nil && e.ObservedAt.After(a.AsOf)) {
					t.Fatal("event/availability exceeds input cutoff")
				}
				key := e.StructureID + "|" + e.Type + "|" + date
				if seen[key] {
					continue
				}
				seen[key] = true
				p := scopedPrediction{Ticker: sample.Ticker, PredictedID: e.StructureID, Type: e.Type, Date: date, FirstSeen: a.AsOf.Format("2006-01-02"), Mapping: "UNATTRIBUTED"}
				if e.StructureID != "" {
					s, ok := byID[e.StructureID]
					if !ok {
						t.Fatalf("dangling event structure: %s", e.StructureID)
					}
					if !s.IdentifiedAt.Before(e.Date) && e.Type!="PS" && e.Type!="SC" && e.Type!="BC" && e.Type!="PSY" && e.Type!="AR" {
						t.Fatalf("structure unavailable: %s %s %s date=%s identified=%s",sample.Ticker,e.Type,e.StructureID,e.Date,s.IdentifiedAt)
					}
					m := mapping[s.ID]
					p.Target, p.Mapping, p.Support, p.Resistance = m.Target, m.Status, s.Support, s.Resistance
				}
				switch p.Mapping {
				case "MAPPED":
					mapped++
				case "AMBIGUOUS":
					ambiguous++
				case "UNMAPPED":
					unmapped++
				case "LOW_CONFIDENCE":
					lowConfidence++
				default:
					unattributed++
				}
				local = append(local, p)
			}
		}
		dedup := map[string]bool{}
		for _, e := range sample.Events {
			if summary[e.Structure] == nil {
				summary[e.Structure] = &scopedSummary{Ticker: sample.Ticker, Structure: e.Structure}
			}
			end := e.End
			if end == "" {
				end = e.Start
			}
			kind, label := "positive", e.Final
			if e.Verdict == "uncertain" || e.Status == "uncertain" || label == "" {
				kind = "uncertain"
			} else if e.Verdict == "incorrect" {
				kind = "rejected"
				label = e.Proposed
			}
			key := kind + "|" + e.Structure + "|" + strings.ToUpper(label) + "|" + e.Start + "|" + end
			if dedup[key] {
				continue
			}
			dedup[key] = true
			row := scopedRow{Ticker: sample.Ticker, EventID: e.ID, Structure: e.Structure, Kind: kind, Label: label, Start: e.Start, End: end, Disposition: "NO_PREDICTION"}
			for _, p := range local {
				if !strings.EqualFold(p.Type, label) || p.Date < e.Start || p.Date > end {
					continue
				}
				if p.Mapping == "MAPPED" && p.Target == e.Structure {
					row.Hit = true
					row.Disposition = "MATCHED_TARGET"
					row.FirstSeen = p.FirstSeen
					break
				}
				if row.Disposition == "NO_PREDICTION" {
					row.Disposition = p.Mapping
				}
				if p.Mapping == "MAPPED" {
					row.Disposition = "OTHER_STRUCTURE"
				}
			}
			if !summary[e.Structure].DefinitionAvailable {
				row.Disposition = "TARGET_UNDEFINED"
				row.Hit = false
			}
			switch kind {
			case "positive":
				positive++
				summary[e.Structure].Positive++
				if row.Hit {
					hits++
					summary[e.Structure].Hits++
				}
			case "rejected":
				rejected++
				summary[e.Structure].Rejected++
				if row.Hit {
					rejectedHits++
					summary[e.Structure].RejectedEmitted++
				}
			case "uncertain":
				uncertain++
				summary[e.Structure].Uncertain++
				row.Disposition = "EXCLUDED_UNCERTAIN"
				row.Hit = false
			}
			rows = append(rows, row)
		}
		ids := []string{}
		for id := range summary {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			summaries = append(summaries, *summary[id])
		}
		predictions = append(predictions, local...)
	}
	out := struct {
		Policy                                                                                                          string
		Positive, Hits, Rejected, RejectedEmitted, Uncertain                                                            int
		MappedPredictions, UnattributedPredictions, UnmappedPredictions, AmbiguousPredictions, LowConfidencePredictions int
		Structures                                                                                                      []scopedSummary
		Events                                                                                                          []scopedRow
		Predictions                                                                                                     []scopedPrediction
	}{"scope-v1: boundary distance <=0.05 reviewed width each (0.05-0.15 held as LOW_CONFIDENCE); formation overlap >=0.5 shorter interval; cost separation >=0.05; no fallback for missing IDs", positive, hits, rejected, rejectedHits, uncertain, mapped, unattributed, unmapped, ambiguous, lowConfidence, summaries, rows, predictions}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("SCOPED_BENCHMARK %s", b)
	if path := os.Getenv("WYCKOFF_SCOPED_OUTPUT"); path != "" {
		if err := os.WriteFile(path, b, 0644); err != nil {
			t.Fatal(err)
		}
	}
	fmt.Printf("SCOPED_SUMMARY positive=%d hits=%d rejected=%d emitted=%d unscoped=%d unmapped=%d ambiguous=%d\n", positive, hits, rejected, rejectedHits, unattributed, unmapped, ambiguous)
}

func TestReviewedScopeMatchingPolicy(t *testing.T) {
	date := func(s string) time.Time { d, _ := time.Parse("2006-01-02", s); return d }
	targets := []reviewedRange{
		{ID: "large", Start: "2026-07-27", Support: reviewedZone{19.15, 20}, Resistance: reviewedZone{25.7, 25.95}},
		{ID: "small", Start: "2026-09-28", Support: reviewedZone{21.3, 21.4}, Resistance: reviewedZone{22.2, 22.5}},
	}
	s := StructureState{StartDate: date("2026-09-28"), IdentifiedAt: date("2026-10-13"), Support: 21.3, Resistance: 22.5}
	if m := matchReviewedScope(s, targets, "2026-10-30"); m.Target != "small" {
		t.Fatalf("small SOS cannot map to large range: %+v", m)
	}
	s.Support, s.Resistance = 19.15, 25.95
	if m := matchReviewedScope(s, targets, "2026-10-30"); m.Target != "large" {
		t.Fatalf("broad context mismatch: %+v", m)
	}
	s.StartDate, s.IdentifiedAt = date("2025-01-01"), date("2025-02-01")
	if m := matchReviewedScope(s, targets, "2026-10-30"); m.Status != "UNMAPPED" {
		t.Fatal("same prices from another cycle must not map")
	}
	s.StartDate, s.IdentifiedAt = date("2026-07-27"), date("2026-10-13")
	s.Support, s.Resistance = 20.5, 25.95
	if m := matchReviewedScope(s, targets, "2026-10-30"); m.Status != "LOW_CONFIDENCE" {
		t.Fatal("loose boundary proximity must not certify identity")
	}
	s.Support, s.Resistance = 19.15, 25.95
	targets = append(targets, reviewedRange{ID: "duplicate", Start: "2026-07-27", Support: reviewedZone{19.15, 20}, Resistance: reviewedZone{25.7, 25.95}})
	s.StartDate, s.IdentifiedAt = date("2026-07-27"), date("2026-10-13")
	if m := matchReviewedScope(s, targets, "2026-10-30"); m.Status != "AMBIGUOUS" {
		t.Fatal("ambiguous mapping must not choose an arbitrary ID")
	}
}
