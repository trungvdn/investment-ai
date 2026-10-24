package macro

import (
	"fmt"
	"math"
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
	RawID         int64      `json:"rawId"`
}
type RawDocument struct {
	Source, URL, ContentType string
	Body                     []byte
	FetchedAt                time.Time
}
type Batch struct {
	Document     RawDocument
	Observations []Observation
	Warnings     []string
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
	var end time.Time
	var err error
	if ind.Frequency == "MONTHLY" {
		end, err = MonthEnd(o.Period)
	} else {
		end, err = time.Parse("2006-01-02", o.Period)
	}
	if err != nil || !end.Equal(o.PeriodEnd) {
		return fmt.Errorf("invalid period")
	}
	if end.After(now) || o.PublishedAt != nil && o.PublishedAt.After(now) {
		return fmt.Errorf("future observation")
	}
	return nil
}
func Status(ind Indicator, latest *Observation, attempt *Attempt, now time.Time) string {
	if !ind.Enabled {
		return "NOT_IMPLEMENTED"
	}
	if latest == nil {
		if attempt != nil {
			if attempt.Message == "FRED_API_KEY is not configured" {
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
