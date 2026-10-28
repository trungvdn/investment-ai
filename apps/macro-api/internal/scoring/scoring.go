// Package scoring calculates an explainable, experimental macro support index.
// It does not predict returns. Missing observations never receive a neutral score.
package scoring

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/trungvdn/investment-ai/apps/macro-api/internal/macro"
)

//go:embed model.json
var configuration []byte

type Rule struct {
	Code        string  `json:"code"`
	Group       string  `json:"group"`
	Weight      float64 `json:"weight"`
	Critical    bool    `json:"critical"`
	Mode        string  `json:"mode"`
	Months      int     `json:"months"`
	Low         float64 `json:"low"`
	High        float64 `json:"high"`
	Slope       float64 `json:"slope"`
	LevelWeight float64 `json:"levelWeight"`
}
type Model struct {
	Version     string  `json:"version"`
	MinCoverage float64 `json:"minCoverage"`
	Rules       []Rule  `json:"rules"`
	Hash        string  `json:"hash"`
}

func DefaultModel() Model {
	var m Model
	if err := json.Unmarshal(configuration, &m); err != nil {
		panic(err)
	}
	sum := sha256.Sum256(configuration)
	m.Hash = hex.EncodeToString(sum[:])
	if err := m.Validate(); err != nil {
		panic(err)
	}
	return m
}
func (m Model) Validate() error {
	seen := map[string]bool{}
	total := 0.0
	for _, r := range m.Rules {
		if _, ok := macro.Find(r.Code); !ok || seen[r.Code] || r.Weight <= 0 || r.Months <= 0 || r.Slope <= 0 || r.LevelWeight < 0 || r.LevelWeight > 1 {
			return fmt.Errorf("invalid scoring rule %s", r.Code)
		}
		switch r.Mode {
		case "range", "pmi", "band", "credit", "iip", "rate":
		default:
			return fmt.Errorf("unknown scoring mode")
		}
		if r.Mode == "band" && r.Low > r.High {
			return fmt.Errorf("invalid band")
		}
		seen[r.Code] = true
		total += r.Weight
	}
	if math.Abs(total-100) > 0.001 || m.Version == "" || m.MinCoverage <= 0 || m.MinCoverage > 100 {
		return fmt.Errorf("invalid scoring model")
	}
	return nil
}

type Component struct {
	Code            string              `json:"code"`
	Name            string              `json:"name"`
	Group           string              `json:"group"`
	Weight          float64             `json:"weight"`
	Critical        bool                `json:"critical"`
	Score           *float64            `json:"score"`
	Contribution    float64             `json:"contribution"`
	EffectiveWeight float64             `json:"effectiveWeight"`
	Reason          string              `json:"reason"`
	Inputs          []macro.Observation `json:"inputs"`
	Warnings        []string            `json:"warnings"`
}
type Group struct {
	Name     string   `json:"name"`
	Weight   float64  `json:"weight"`
	Coverage float64  `json:"coverage"`
	Score    *float64 `json:"score"`
}
type Snapshot struct {
	AsOf         time.Time   `json:"asOf"`
	ModelVersion string      `json:"modelVersion"`
	ModelHash    string      `json:"modelHash"`
	InputHash    string      `json:"inputHash"`
	Score        *float64    `json:"score"`
	Status       string      `json:"status"`
	Coverage     float64     `json:"coverage"`
	Components   []Component `json:"components"`
	Groups       []Group     `json:"groups"`
}
type Point struct {
	Kind string `json:"kind"`
	Snapshot
}
type Response struct {
	Model     Model     `json:"model"`
	Frequency string    `json:"frequency"`
	Current   *Snapshot `json:"current"`
	History   []Point   `json:"history"`
}

func clamp(v float64) float64               { return math.Max(0, math.Min(100, v)) }
func ptr(v float64) *float64                { return &v }
func distance(v, low, high float64) float64 { return math.Max(low-v, math.Max(0, v-high)) }
func month(t time.Time, offset int) string {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, offset, 0).Format("2006-01")
}
func mean(v []float64) float64 {
	s := 0.0
	for _, x := range v {
		s += x
	}
	return s / float64(len(v))
}

// Calculate receives canonical, one-per-period observations. Every input vintage is
// frozen in the result; subsequent source revisions cannot rewrite old snapshots.
func Calculate(m Model, data map[string][]macro.Observation, now time.Time) Snapshot {
	out := Snapshot{AsOf: now.UTC(), ModelVersion: m.Version, ModelHash: m.Hash, Status: "INSUFFICIENT", Components: []Component{}, Groups: []Group{}}
	critical := true
	groupValues := map[string][]float64{}
	for _, r := range m.Rules {
		ind, _ := macro.Find(r.Code)
		c := Component{Code: r.Code, Name: ind.Name, Group: r.Group, Weight: r.Weight, Critical: r.Critical, Reason: "MISSING", Inputs: []macro.Observation{}, Warnings: []string{}}
		all := append([]macro.Observation(nil), data[r.Code]...)
		sort.Slice(all, func(i, j int) bool { return all[i].PeriodEnd.Before(all[j].PeriodEnd) })
		usable := []macro.Observation{}
		invalid := false
		for _, o := range all {
			if o.IndicatorCode != r.Code || o.PeriodEnd.After(now) || o.FirstSeenAt.After(now) || (o.PublishedAt != nil && o.PublishedAt.After(now)) {
				continue
			}
			if err := macro.Validate(o, now); err != nil || o.RawID <= 0 || o.Revision <= 0 || o.FirstSeenAt.IsZero() {
				invalid = true
				continue
			}
			usable = append(usable, o)
		}
		if invalid {
			c.Warnings = append(c.Warnings, "INVALID_INPUT_EXCLUDED")
		}
		if len(usable) > 0 {
			latest := usable[len(usable)-1]
			c.Inputs = []macro.Observation{latest}
			if now.Sub(latest.PeriodEnd) > time.Duration(ind.MaxAgeDays)*24*time.Hour {
				c.Reason = "STALE"
			} else {
				inputs, score, err := component(r, usable)
				c.Inputs = inputs
				if err != nil {
					c.Reason = err.Error()
				} else {
					c.Score = ptr(score)
					c.Reason = "ELIGIBLE"
					out.Coverage += r.Weight
				}
			}
		}
		for _, o := range c.Inputs {
			if strings.Contains(o.Method, "NOT_INDEPENDENTLY_VERIFIED") {
				c.Warnings = append(c.Warnings, "USER_SUPPLIED_UNVERIFIED")
				break
			}
		}
		if c.Score == nil && r.Critical {
			critical = false
		}
		g := groupValues[r.Group]
		if g == nil {
			g = []float64{0, 0, 0}
		}
		g[0] += r.Weight
		if c.Score != nil {
			g[1] += r.Weight
			g[2] += r.Weight * *c.Score
		}
		groupValues[r.Group] = g
		out.Components = append(out.Components, c)
	}
	if out.Coverage > 0 {
		total := 0.0
		for i := range out.Components {
			c := &out.Components[i]
			if c.Score != nil {
				c.EffectiveWeight = c.Weight / out.Coverage * 100
				c.Contribution = c.EffectiveWeight * *c.Score / 100
				total += c.Contribution
			}
		}
		out.Score = ptr(total)
		out.Status = "PROVISIONAL"
		if out.Coverage >= m.MinCoverage && critical {
			out.Status = "READY"
		}
	}
	for _, name := range []string{"Liquidity", "Production", "Inflation"} {
		g := groupValues[name]
		if len(g) == 0 {
			continue
		}
		result := Group{Name: name, Weight: g[0], Coverage: g[1] / g[0] * 100}
		if g[1] > 0 {
			result.Score = ptr(g[2] / g[1])
		}
		out.Groups = append(out.Groups, result)
	}
	// CollectedAt is a refresh timestamp, not a new information vintage. Exclude it
	// and Priority from the fingerprint while keeping source, method and revision.
	fingerprint := out.Components
	fingerprint = append([]Component(nil), fingerprint...)
	for i := range fingerprint {
		fingerprint[i].Inputs = append([]macro.Observation(nil), fingerprint[i].Inputs...)
		for j := range fingerprint[i].Inputs {
			fingerprint[i].Inputs[j].CollectedAt = time.Time{}
		}
	}
	raw, _ := json.Marshal(struct {
		Model      string
		Components []Component
	}{m.Hash, fingerprint})
	hash := sha256.Sum256(raw)
	out.InputHash = hex.EncodeToString(hash[:])
	return out
}

var methodMetadata = regexp.MustCompile(`; (ARTICLE_ID|DENOMINATOR_ARTICLE_ID|PARSER)=[^;]*`)

func methodology(method string) string { return methodMetadata.ReplaceAllString(method, "") }

func component(r Rule, all []macro.Observation) ([]macro.Observation, float64, error) {
	latest := all[len(all)-1]
	inputs := []macro.Observation{latest}
	fail := func(reason string) ([]macro.Observation, float64, error) { return inputs, 0, fmt.Errorf("%s", reason) }
	byMonth := map[string]macro.Observation{}
	for _, o := range all {
		byMonth[month(o.PeriodEnd, 0)] = o
	}
	get := func(offset int) (macro.Observation, bool) {
		o, ok := byMonth[month(latest.PeriodEnd, offset)]
		return o, ok
	}
	// Different series cannot safely be spliced into a trend. For M2 the provider
	// must also use the same methodology, including the October 2025 break.
	compatible := func(o macro.Observation) bool {
		if o.SeriesID != latest.SeriesID {
			return false
		}
		if r.Code == "VN_M2_GROWTH" && methodology(o.Method) != methodology(latest.Method) {
			return false
		}
		return true
	}
	if r.Mode == "credit" || r.Mode == "iip" {
		values := []float64{}
		for offset := 0; offset < 6; offset++ {
			o, ok := get(-offset)
			if !ok {
				return fail("INSUFFICIENT_HISTORY")
			}
			if !compatible(o) {
				return fail("INCOMPARABLE_HISTORY")
			}
			if offset > 0 {
				inputs = append(inputs, o)
			}
			value := o.Value
			if r.Mode == "credit" {
				if o.PeriodEnd.Month() != time.January {
					previous, ok := byMonth[month(o.PeriodEnd, -1)]
					if !ok {
						return fail("INSUFFICIENT_HISTORY")
					}
					if !compatible(previous) {
						return fail("INCOMPARABLE_HISTORY")
					}
					if previous.Value <= -100 || o.Value <= -100 {
						return fail("INVALID_CREDIT_BASE")
					}
					value = 100 * ((1+o.Value/100)/(1+previous.Value/100) - 1)
					inputs = append(inputs, previous)
				}
			}
			values = append(values, value)
		}
		a, b := mean(values[:3]), mean(values[3:])
		momentum := clamp(50 + (a-b)*r.Slope)
		if r.Mode == "credit" {
			return inputs, momentum, nil
		}
		return inputs, r.LevelWeight*clamp(50+a*r.Slope) + (1-r.LevelWeight)*momentum, nil
	}
	var previous macro.Observation
	if r.Mode == "rate" {
		// Anchor on calendar-month start to avoid AddDate's end-of-month rollover.
		anchorMonth := time.Date(latest.PeriodEnd.Year(), latest.PeriodEnd.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -r.Months, 0)
		lastDay := anchorMonth.AddDate(0, 1, -1).Day()
		day := latest.PeriodEnd.Day()
		if day > lastDay {
			day = lastDay
		}
		anchor := time.Date(anchorMonth.Year(), anchorMonth.Month(), day, 0, 0, 0, 0, time.UTC)
		found := false
		for _, o := range all {
			if !o.PeriodEnd.After(anchor) && anchor.Sub(o.PeriodEnd) <= 7*24*time.Hour {
				previous = o
				found = true
			}
		}
		if !found {
			return fail("INSUFFICIENT_HISTORY")
		}
	} else {
		var ok bool
		previous, ok = get(-r.Months)
		if !ok {
			return fail("INSUFFICIENT_HISTORY")
		}
	}
	inputs = append(inputs, previous)
	if !compatible(previous) {
		return fail("INCOMPARABLE_HISTORY")
	}
	delta := latest.Value - previous.Value
	switch r.Mode {
	case "range":
		if latest.UpperValue == nil || previous.UpperValue == nil {
			return fail("RANGE_REQUIRED")
		}
		// Average two scores, never synthesize a midpoint interest rate.
		return inputs, (clamp(50-delta*r.Slope) + clamp(50-(*latest.UpperValue-*previous.UpperValue)*r.Slope)) / 2, nil
	case "rate":
		return inputs, clamp(50 - delta*r.Slope), nil
	case "pmi":
		return inputs, r.LevelWeight*clamp(50+(latest.Value-50)*r.Slope) + (1-r.LevelWeight)*clamp(50+delta*r.Slope), nil
	case "band":
		currentDistance := distance(latest.Value, r.Low, r.High)
		oldDistance := distance(previous.Value, r.Low, r.High)
		trend := clamp(50 + (oldDistance-currentDistance)*r.Slope)
		return inputs, r.LevelWeight*clamp(100-currentDistance*r.Slope) + (1-r.LevelWeight)*trend, nil
	}
	return fail("UNSUPPORTED_RULE")
}
