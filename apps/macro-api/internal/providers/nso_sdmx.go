package providers

import (
	"context"
	"encoding/xml"
	"fmt"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/macro"
	"math"
	"sort"
	"strconv"
	"time"
)

const NSDPBase = "https://nsdp.nso.gov.vn/GSO-chung/SDMXFiles/GSO/"

type SDMX struct {
	HTTP
	Code, File, Series, URL string
}

func (s *SDMX) ID() string      { return "nso:sdmx:" + s.Series }
func (s *SDMX) Codes() []string { return []string{s.Code} }
func (s *SDMX) Fetch(ctx context.Context) (macro.Batch, error) {
	u := s.URL
	if u == "" {
		u = NSDPBase + s.File + ".xml"
	}
	doc, err := s.Get(ctx, u, u, "NSO")
	if err != nil {
		return macro.Batch{}, err
	}
	obs, err := ParseSDMX(doc.Body, s.Code, s.Series, u)
	return macro.Batch{Document: doc, Observations: obs}, err
}

type sdmxSeries struct {
	Area        string `xml:"REF_AREA,attr"`
	Indicator   string `xml:"INDICATOR,attr"`
	Freq        string `xml:"FREQ,attr"`
	Base        string `xml:"BASE_PER,attr"`
	Multiplier  string `xml:"UNIT_MULT,attr"`
	Counterpart string `xml:"COUNTERPART_AREA,attr"`
	Obs         []struct {
		Period string `xml:"TIME_PERIOD,attr"`
		Value  string `xml:"OBS_VALUE,attr"`
	} `xml:"Obs"`
}

func ParseSDMX(body []byte, code, seriesID, sourceURL string) ([]macro.Observation, error) {
	var payload struct {
		DataSet struct {
			Series []sdmxSeries `xml:"Series"`
		} `xml:"DataSet"`
	}
	if err := xml.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("invalid NSO SDMX XML")
	}
	var selected *sdmxSeries
	for i := range payload.DataSet.Series {
		s := &payload.DataSet.Series[i]
		if s.Area == "VN" && s.Indicator == seriesID && s.Freq == "M" && s.Counterpart == "_Z" {
			if selected != nil {
				return nil, fmt.Errorf("ambiguous NSO series")
			}
			selected = s
		}
	}
	if selected == nil {
		return nil, fmt.Errorf("NSO series %s not found", seriesID)
	}
	if (code == "VN_EXPORT_YOY" && selected.Multiplier != "6") || (code != "VN_EXPORT_YOY" && (selected.Multiplier != "0" || selected.Base == "")) {
		return nil, fmt.Errorf("unexpected NSO unit multiplier")
	}
	// One selected series supplies both sides of YoY; never bridge different bases.
	values := map[string]float64{}
	for _, row := range selected.Obs {
		if _, err := macro.MonthEnd(row.Period); err != nil {
			return nil, fmt.Errorf("invalid NSO monthly period")
		}
		v, err := strconv.ParseFloat(row.Value, 64)
		if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, fmt.Errorf("invalid NSO numeric observation")
		}
		if _, exists := values[row.Period]; exists {
			return nil, fmt.Errorf("duplicate NSO monthly observation")
		}
		values[row.Period] = v
	}
	ind, ok := macro.Find(code)
	if !ok {
		return nil, fmt.Errorf("unknown NSO indicator")
	}
	out := []macro.Observation{}
	for period, v := range values {
		t, _ := time.Parse("2006-01", period)
		previous, ok := values[t.AddDate(-1, 0, 0).Format("2006-01")]
		if !ok || previous <= 0 {
			continue
		}
		end, _ := macro.MonthEnd(period)
		o := macro.Observation{IndicatorCode: code, Period: period, PeriodEnd: end, Value: (v/previous - 1) * 100, Unit: ind.Unit, Source: "NSO", SourceURL: sourceURL, SeriesID: seriesID, Method: "YOY_FROM_SAME_MONTH_INDEX_OR_LEVEL; base=" + selected.Base + "; unit_mult=" + selected.Multiplier, Priority: 50}
		if err := macro.Validate(o, time.Now()); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Period < out[j].Period })
	if len(out) == 0 {
		return nil, fmt.Errorf("NSO series lacks comparable prior-year observations")
	}
	return out, nil
}
