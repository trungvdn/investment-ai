package macro

import (
	"fmt"
	"math"
	"strings"
	"time"
)

type Indicator struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Group       string `json:"group"`
	Country     string `json:"country"`
	Unit        string `json:"unit"`
	Frequency   string `json:"frequency"`
	Basis       string `json:"comparisonBasis"`
	Enabled     bool   `json:"enabled"`
	MaxAgeDays  int    `json:"maxAgeDays"`
	Description string `json:"description"`
}
type Observation struct {
	UpperValue    *float64   `json:"upperValue,omitempty"`
	IndicatorCode string     `json:"indicatorCode"`
	Period        string     `json:"period"`
	PeriodEnd     time.Time  `json:"periodEnd"`
	Value         float64    `json:"value"`
	Unit          string     `json:"unit"`
	Source        string     `json:"source"`
	SourceURL     string     `json:"sourceUrl"`
	SeriesID      string     `json:"sourceSeriesId"`
	Method        string     `json:"method"`
	Priority      int        `json:"-"`
	PublishedAt   *time.Time `json:"publishedAt"`
	FirstSeenAt   time.Time  `json:"firstSeenAt"`
	CollectedAt   time.Time  `json:"collectedAt"`
	Revision      int        `json:"revision"`
	RawIndex      int        `json:"-"`
	RawID         int64      `json:"rawId"`
}
type RawDocument struct {
	Source, URL, ContentType string
	Body                     []byte
	FetchedAt                time.Time
}
type Batch struct {
	Document     RawDocument
	Documents    []RawDocument
	Observations []Observation
	Warnings     []string
	Unavailable  map[string]string
}
type Attempt struct {
	Code      string    `json:"indicatorCode"`
	CheckedAt time.Time `json:"checkedAt"`
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
}
type Card struct {
	Indicator
	Status      string        `json:"status"`
	Latest      *Observation  `json:"latest"`
	History     []Observation `json:"history"`
	LastAttempt *Attempt      `json:"lastAttempt"`
}

func MonthEnd(period string) (time.Time, error) {
	t, err := time.Parse("2006-01", period)
	if err != nil {
		return time.Time{}, err
	}
	return t.AddDate(0, 1, -1), nil
}
func QuarterEnd(period string) (time.Time, error) {
	if len(period) != 7 || period[4:6] != "-Q" || period[6] < '1' || period[6] > '4' {
		return time.Time{}, fmt.Errorf("invalid quarter")
	}
	year, err := time.Parse("2006", period[:4])
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(year.Year(), time.Month(int(period[6]-'0')*3)+1, 0, 0, 0, 0, 0, time.UTC), nil
}
func Validate(o Observation, now time.Time) error {
	ind, ok := Find(o.IndicatorCode)
	if !ok || !ind.Enabled {
		return fmt.Errorf("unknown or disabled indicator %s", o.IndicatorCode)
	}
	if math.IsNaN(o.Value) || math.IsInf(o.Value, 0) {
		return fmt.Errorf("non-finite observation")
	}
	if o.Unit != ind.Unit || o.PeriodEnd.IsZero() || o.Source == "" || o.SourceURL == "" || o.SeriesID == "" || o.Method == "" {
		return fmt.Errorf("missing or inconsistent observation metadata")
	}
	if (o.IndicatorCode == "VN_PMI" && (o.Value < 0 || o.Value > 100)) || ((o.IndicatorCode == "US_DXY" || o.IndicatorCode == "US_BROAD_USD" || o.IndicatorCode == "VN_USD_VND" || o.IndicatorCode == "VN_M2_LEVEL") && o.Value <= 0) || ((ind.Basis == "YoY" || ind.Basis == "YTD") && o.Value < -100) {
		return fmt.Errorf("observation outside definition bounds")
	}
	if ind.Basis == "RANGE" {
		if o.UpperValue == nil || math.IsNaN(*o.UpperValue) || math.IsInf(*o.UpperValue, 0) || o.Value < 0 || *o.UpperValue > 100 || o.Value > *o.UpperValue {
			return fmt.Errorf("invalid deposit rate range")
		}
	} else if o.UpperValue != nil {
		return fmt.Errorf("upper bound on non-range indicator")
	}
	var end time.Time
	var err error
	if ind.Frequency == "MONTHLY" {
		end, err = MonthEnd(o.Period)
	} else if ind.Frequency == "QUARTERLY" {
		end, err = QuarterEnd(o.Period)
	} else {
		end, err = time.Parse("2006-01-02", o.Period)
	}
	if err != nil || !end.Equal(o.PeriodEnd) {
		return fmt.Errorf("invalid period")
	}
	location := time.UTC
	if ind.Country == "VN" {
		location = time.FixedZone("ICT", 7*3600)
	}
	localNow := now.In(location)
	today := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, time.UTC)
	if end.After(today) || o.PublishedAt != nil && o.PublishedAt.After(now) {
		return fmt.Errorf("future observation")
	}
	return nil
}
func Status(ind Indicator, latest *Observation, attempt *Attempt, now time.Time) string {
	if !ind.Enabled {
		return "NOT_IMPLEMENTED"
	}
	if attempt != nil && !attempt.Success && strings.HasPrefix(attempt.Message, "AWAITING_RELEASE:") {
		if latest != nil && now.Sub(latest.PeriodEnd) > time.Duration(ind.MaxAgeDays)*24*time.Hour {
			return "STALE"
		}
		return "AWAITING_RELEASE"
	}
	if attempt != nil && !attempt.Success && strings.HasPrefix(attempt.Message, "NO_PUBLISHED_RATE:") {
		return "NO_PUBLISHED_RATE"
	}
	if attempt != nil && !attempt.Success && strings.HasPrefix(attempt.Message, "NOT_COMPARABLE:") {
		return "NOT_COMPARABLE"
	}
	if latest == nil {
		if attempt != nil {
			if attempt.Message == "FRED_API_KEY is not configured" || attempt.Message == "PMI_RELEASE_URLS is not configured" || attempt.Message == "DXY_FEED_URL is not configured" {
				return "NOT_CONFIGURED"
			}
			if !attempt.Success {
				return "ERROR"
			}
		}
		return "MISSING"
	}
	if now.Sub(latest.PeriodEnd) > time.Duration(ind.MaxAgeDays)*24*time.Hour {
		return "STALE"
	}
	if attempt != nil && !attempt.Success {
		return "SOURCE_ERROR"
	}
	return "AVAILABLE"
}
