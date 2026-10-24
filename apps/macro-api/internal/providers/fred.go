package providers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/macro"
	"net/url"
	"strconv"
	"time"
)

var ErrNotConfigured = errors.New("FRED_API_KEY is not configured")

type FRED struct {
	HTTP
	Code, Series, APIKey, BaseURL string
	Start                         time.Time
}

func (f *FRED) ID() string      { return "fred:" + f.Series }
func (f *FRED) Codes() []string { return []string{f.Code} }
func (f *FRED) Fetch(ctx context.Context) (macro.Batch, error) {
	if f.APIKey == "" {
		return macro.Batch{}, ErrNotConfigured
	}
	base := f.BaseURL
	if base == "" {
		base = "https://api.stlouisfed.org/fred/series/observations"
	}
	u, err := url.Parse(base)
	if err != nil {
		return macro.Batch{}, fmt.Errorf("invalid FRED URL")
	}
	start := f.Start
	if start.IsZero() {
		start = time.Now().AddDate(-5, 0, 0)
	}
	q := u.Query()
	q.Set("api_key", f.APIKey)
	q.Set("series_id", f.Series)
	q.Set("file_type", "json")
	q.Set("observation_start", start.Format("2006-01-02"))
	q.Set("sort_order", "asc")
	q.Set("limit", "100000")
	u.RawQuery = q.Encode()
	doc, err := f.Get(ctx, u.String(), "https://fred.stlouisfed.org/series/"+f.Series, "FRED")
	if err != nil {
		return macro.Batch{}, err
	}
	obs, err := ParseFRED(doc.Body, f.Code, f.Series)
	return macro.Batch{Document: doc, Observations: obs}, err
}
func ParseFRED(body []byte, code, series string) ([]macro.Observation, error) {
	if map[string]string{"US_EFFR": "DFF", "US_10Y": "DGS10", "US_BROAD_USD": "DTWEXBGS"}[code] != series {
		return nil, fmt.Errorf("FRED indicator/series mismatch")
	}
	var data struct {
		Count        int `json:"count"`
		Observations []struct {
			Date  string `json:"date"`
			Value string `json:"value"`
		} `json:"observations"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("invalid FRED JSON")
	}
	if data.Count > len(data.Observations) {
		return nil, fmt.Errorf("FRED response is truncated; pagination required")
	}
	ind, ok := macro.Find(code)
	if !ok {
		return nil, fmt.Errorf("unknown FRED code")
	}
	out := []macro.Observation{}
	for _, row := range data.Observations {
		if row.Value == "." || row.Value == "" {
			continue
		}
		v, err := strconv.ParseFloat(row.Value, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid FRED numeric observation")
		}
		date, err := time.Parse("2006-01-02", row.Date)
		if err != nil {
			return nil, fmt.Errorf("invalid FRED date")
		}
		o := macro.Observation{IndicatorCode: code, Period: row.Date, PeriodEnd: date, Value: v, Unit: ind.Unit, Source: "FRED", SourceURL: "https://fred.stlouisfed.org/series/" + series, SeriesID: series, Method: "SOURCE_LEVEL", Priority: 100}
		if err := macro.Validate(o, time.Now()); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("FRED returned no valid observations")
	}
	return out, nil
}
