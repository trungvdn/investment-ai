package providers

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/trungvdn/investment-ai/apps/macro-api/internal/macro"
)

const DXYNotConfigured = "DXY_FEED_URL is not configured"
const DXYSourceURL = "https://www.ice.com/fixed-income-data-services/index-solutions/currency-indices"

// DXY reads an operator-provided authorized export. No Broad USD/futures fallback.
type DXY struct {
	HTTP
	URL, Token string
}

func (p *DXY) ID() string      { return "ice:dxy" }
func (p *DXY) Codes() []string { return []string{"US_DXY"} }
func (p *DXY) Fetch(ctx context.Context) (macro.Batch, error) {
	if p.URL == "" {
		return macro.Batch{}, fmt.Errorf(DXYNotConfigured)
	}
	u, err := url.Parse(p.URL)
	if err != nil || u.Scheme != "https" || u.User != nil {
		return macro.Batch{}, fmt.Errorf("DXY feed must use HTTPS without URL credentials")
	}
	h := p.HTTP
	h.Headers = make(http.Header)
	if p.Token != "" {
		h.Headers.Set("Authorization", "Bearer "+p.Token)
	}
	doc, err := h.Get(ctx, p.URL, DXYSourceURL, "ICE")
	b := macro.Batch{Document: doc}
	if err != nil {
		return b, err
	}
	r := csv.NewReader(bytes.NewReader(doc.Body))
	records, err := r.ReadAll()
	if err != nil || len(records) < 2 || len(records) > 10001 {
		return b, fmt.Errorf("invalid DXY CSV")
	}
	if len(records[0]) != 3 || records[0][0] != "date" || records[0][1] != "value" || records[0][2] != "symbol" {
		return b, fmt.Errorf("DXY export requires date,value,symbol columns")
	}
	seen := map[string]bool{}
	for _, row := range records[1:] {
		if len(row) != 3 || row[2] != "DXY" || seen[row[0]] {
			return b, fmt.Errorf("ambiguous DXY date or symbol; futures and Broad USD are not accepted")
		}
		seen[row[0]] = true
		date, e := time.Parse("2006-01-02", row[0])
		if e != nil {
			return b, fmt.Errorf("invalid DXY date")
		}
		value, e := strconv.ParseFloat(row[1], 64)
		if e != nil || value <= 0 {
			return b, fmt.Errorf("invalid DXY value")
		}
		o := macro.Observation{IndicatorCode: "US_DXY", Period: row[0], PeriodEnd: date, Value: value, Unit: "index Mar 1973=100", Source: "ICE", SourceURL: DXYSourceURL, SeriesID: "ICE_DXY", Method: "AUTHORIZED_EXPORT_DXY_SPOT_INDEX; NOT_DX_FUTURES_NOT_BROAD_USD", Priority: 100}
		if e = macro.Validate(o, time.Now()); e != nil {
			return b, e
		}
		b.Observations = append(b.Observations, o)
	}
	return b, nil
}
