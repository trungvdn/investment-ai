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

const PPISeries = "VNM_PPPI_ISIC4_BTE_BY_PP_IX"

type PPI struct {
	HTTP
	URL string
}

func (p *PPI) ID() string      { return "nso:sdmx:" + PPISeries }
func (p *PPI) Codes() []string { return []string{"VN_PPI_YOY"} }
func (p *PPI) Fetch(ctx context.Context) (macro.Batch, error) {
	u := p.URL
	if u == "" {
		u = NSDPBase + "PPIVNM.xml"
	}
	doc, err := p.Get(ctx, u, u, "NSO")
	if err != nil {
		return macro.Batch{}, err
	}
	out, err := ParsePPI(doc.Body, u)
	return macro.Batch{Document: doc, Observations: out}, err
}
func ParsePPI(body []byte, sourceURL string) ([]macro.Observation, error) {
	var payload struct {
		DataSet struct {
			Series []sdmxSeries `xml:"Series"`
		} `xml:"DataSet"`
	}
	if err := xml.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("invalid PPI SDMX")
	}
	var found *sdmxSeries
	for i := range payload.DataSet.Series {
		s := &payload.DataSet.Series[i]
		if s.Indicator == PPISeries && s.Area == "VN" && s.Counterpart == "_Z" && s.Freq == "Q" {
			if found != nil {
				return nil, fmt.Errorf("ambiguous industrial PPI series")
			}
			found = s
		}
	}
	if found == nil || found.Multiplier != "0" || found.Base != "_Z" {
		return nil, fmt.Errorf("unexpected PPI series or unit/base metadata")
	}
	out := []macro.Observation{}
	seen := map[string]bool{}
	for _, row := range found.Obs {
		end, err := macro.QuarterEnd(row.Period)
		if err != nil || seen[row.Period] {
			return nil, fmt.Errorf("invalid or duplicate PPI quarter")
		}
		seen[row.Period] = true
		index, err := strconv.ParseFloat(row.Value, 64)
		if err != nil || math.IsNaN(index) || math.IsInf(index, 0) || index <= 0 {
			return nil, fmt.Errorf("invalid PPI index")
		}
		// BY_PP series already compares the current quarter with the same quarter
		// of the previous year (100). It is NOT a fixed-base index for another YoY ratio.
		o := macro.Observation{IndicatorCode: "VN_PPI_YOY", Period: row.Period, PeriodEnd: end, Value: index - 100, Unit: "% YoY", Source: "NSO", SourceURL: sourceURL, SeriesID: PPISeries, Method: "SAME_QUARTER_PREVIOUS_YEAR_100_INDEX_MINUS_100; INDUSTRY_B_TO_E", Priority: 100}
		if err := macro.Validate(o, time.Now()); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Period < out[j].Period })
	if len(out) == 0 {
		return nil, fmt.Errorf("empty PPI series")
	}
	return out, nil
}
