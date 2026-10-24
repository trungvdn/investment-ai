package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/macro"
	htmlparse "golang.org/x/net/html"
	"math"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const SBVArticlesURL = "https://sbv.gov.vn/o/article/v1.0/articles"
const M2MethodStart = "2025-10"

type SBVStatistics struct {
	HTTP
	Kind, Endpoint string
}

func (s *SBVStatistics) ID() string { return "sbv:statistics:" + s.Kind }
func (s *SBVStatistics) Codes() []string {
	if s.Kind == "m2" {
		return []string{"VN_M2_LEVEL", "VN_M2_YTD", "VN_M2_GROWTH"}
	}
	return []string{"VN_CREDIT_GROWTH"}
}
func (s *SBVStatistics) Fetch(ctx context.Context) (macro.Batch, error) {
	endpoint := s.Endpoint
	if endpoint == "" {
		endpoint = SBVArticlesURL
	}
	structure := "10038520"
	if s.Kind == "credit" {
		structure = "10034332"
	}
	batch := macro.Batch{}
	all := []sbvArticle{}
	expected := 0
	for page := 1; page <= 20; page++ {
		u, err := url.Parse(endpoint)
		if err != nil {
			return batch, fmt.Errorf("invalid SBV articles endpoint")
		}
		q := u.Query()
		q.Set("scopeKey", "20117")
		q.Set("contentStructureId", structure)
		q.Set("pageSize", "200")
		q.Set("page", strconv.Itoa(page))
		q.Set("filter", "status eq 0")
		u.RawQuery = q.Encode()
		doc, err := s.Get(ctx, u.String(), u.String(), "SBV")
		if err != nil {
			return batch, err
		}
		if page == 1 {
			batch.Document = doc
		} else {
			batch.Documents = append(batch.Documents, doc)
		}
		var payload struct {
			Articles   []sbvArticle `json:"articles"`
			Page       int          `json:"page"`
			LastPage   int          `json:"lastPage"`
			TotalCount int          `json:"totalCount"`
		}
		decoder := json.NewDecoder(bytes.NewReader(doc.Body))
		decoder.UseNumber()
		if err = decoder.Decode(&payload); err != nil {
			return batch, fmt.Errorf("invalid SBV article JSON")
		}
		if payload.Page != page || payload.LastPage < page || payload.LastPage > 20 || len(payload.Articles) == 0 {
			return batch, fmt.Errorf("invalid or truncated SBV pagination")
		}
		if page == 1 {
			expected = payload.TotalCount
		} else if payload.TotalCount != expected {
			return batch, fmt.Errorf("SBV dataset changed during pagination; retry later")
		}
		for _, a := range payload.Articles {
			a.RawIndex = page - 1
			all = append(all, a)
		}
		if page == payload.LastPage {
			if expected != len(all) {
				return batch, fmt.Errorf("incomplete SBV history")
			}
			break
		}
	}
	out, reasons, warnings, err := parseStatistics(all, s.Kind)
	batch.Observations = out
	batch.Unavailable = reasons
	batch.Warnings = warnings
	return batch, err
}

type sbvArticle struct {
	ID        json.Number    `json:"id"`
	ArticleID string         `json:"articleId"`
	Status    int            `json:"status"`
	Structure int            `json:"ddmStructureId"`
	Title     string         `json:"title"`
	Fields    map[string]any `json:"fields"`
	Published string         `json:"datePublished"`
	Modified  string         `json:"dateModified"`
	Version   float64        `json:"version"`
	RawIndex  int
}

func fieldText(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case json.Number:
		return string(x)
	case map[string]any:
		for _, key := range []string{"content", "value", "vi_VN"} {
			if value, ok := x[key]; ok {
				return fieldText(value)
			}
		}
	}
	return ""
}
func fieldNumber(v any) (float64, error) {
	if n, ok := v.(json.Number); ok {
		value, err := n.Float64()
		if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
			return 0, fmt.Errorf("invalid numeric field")
		}
		return value, nil
	}
	return vietnamNumber(fieldText(v))
}
func sourceTime(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil
	}
	return &t
}

type monthlySBV struct {
	Article    sbvArticle
	Period     string
	Level, YTD float64
	HasYTD     bool
	Updated    time.Time
}

func parseStatistics(articles []sbvArticle, kind string) ([]macro.Observation, map[string]string, []string, error) {
	return parseStatisticsAt(articles, kind, time.Now())
}

func parseStatisticsAt(articles []sbvArticle, kind string, now time.Time) ([]macro.Observation, map[string]string, []string, error) {
	if kind != "m2" && kind != "credit" {
		return nil, nil, nil, fmt.Errorf("unknown SBV statistic kind")
	}
	wanted, dateKey, levelKey, growthKey := 10038520, "Date00760329", "soDuTongPhuongTienThanhToan", "tocDoTangGiamTongPhuongTienThanhToan"
	publicURL := SBVM2URL
	if kind == "credit" {
		wanted, dateKey, levelKey, growthKey = 10034332, "Date62813077", "soDuTongCong", "tocDoTangGiamTongCong"
		publicURL = SBVCreditURL
	}
	selected := map[string]monthlySBV{}
	warnings := []string{}
	for _, a := range articles {
		if a.Status != 0 || a.Structure != wanted {
			continue
		}
		dateString := fieldText(a.Fields[dateKey])
		if dateString == "" {
			continue
		}
		date, err := time.Parse("2006-01-02", dateString)
		if err != nil {
			warnings = append(warnings, "Skipped invalid source period date")
			continue
		}
		period, err := reportPeriod(a.Title)
		if err != nil || period != date.Format("2006-01") {
			warnings = append(warnings, "Skipped mismatched period/title")
			continue
		}
		end, _ := macro.MonthEnd(period)
		if !date.Equal(end) {
			warnings = append(warnings, "Skipped legacy date that is not an explicit month end")
			continue
		}
		if kind == "m2" && period < M2MethodStart {
			warnings = append(warnings, "Excluded pre-10/2025 M2 methodology")
			continue
		}
		if kind == "m2" {
			noteText := fieldText(a.Fields["ghiChu"])
			root, _ := htmlparse.Parse(strings.NewReader(noteText))
			note := fold(nodeText(root))
			known := strings.Contains(note, "10/2025") || strings.Contains(note, "10 nam 2025") || (strings.Contains(note, "phan to lai") && strings.Contains(note, "giay to co gia") && strings.Contains(note, "tctd khac") && strings.Contains(note, "tieu chuan quoc te"))
			if strings.Contains(note, "thay doi") && !known {
				return nil, nil, warnings, fmt.Errorf("unrecognized M2 methodology change in %s; review source before importing", period)
			}
		}
		level, err := fieldNumber(a.Fields[levelKey])
		if err != nil || level <= 0 {
			return nil, nil, warnings, fmt.Errorf("invalid SBV total level for %s", period)
		}
		hasYTD := fieldText(a.Fields[growthKey]) != ""
		growth := 0.0
		if hasYTD {
			growth, err = fieldNumber(a.Fields[growthKey])
			if err != nil {
				return nil, nil, warnings, fmt.Errorf("invalid SBV YTD for %s", period)
			}
		} else if kind == "credit" {
			return nil, nil, warnings, fmt.Errorf("missing total credit YTD for %s", period)
		}
		updated := time.Time{}
		if t := sourceTime(a.Modified); t != nil {
			updated = *t
		} else if t := sourceTime(a.Published); t != nil {
			updated = *t
		}
		row := monthlySBV{Article: a, Period: period, Level: level, YTD: growth, HasYTD: hasYTD, Updated: updated}
		if old, ok := selected[period]; ok {
			if updated.Equal(old.Updated) && a.Version == old.Article.Version && (level != old.Level || growth != old.YTD || hasYTD != old.HasYTD) {
				return nil, nil, warnings, fmt.Errorf("conflicting same-vintage SBV totals for %s", period)
			}
			if updated.Before(old.Updated) || (updated.Equal(old.Updated) && a.Version <= old.Article.Version) {
				continue
			}
		}
		selected[period] = row
	}
	periods := []string{}
	for p := range selected {
		periods = append(periods, p)
	}
	sort.Strings(periods)
	out := []macro.Observation{}
	reasons := map[string]string{}
	add := func(row monthlySBV, code, series, method string, value float64) error {
		ind, _ := macro.Find(code)
		end, _ := macro.MonthEnd(row.Period)
		o := macro.Observation{IndicatorCode: code, Period: row.Period, PeriodEnd: end, Value: value, Unit: ind.Unit, Source: "SBV", SourceURL: publicURL, SeriesID: series, Method: method + "; ARTICLE_ID=" + row.Article.ArticleID, Priority: 100, PublishedAt: sourceTime(row.Article.Published), RawIndex: row.Article.RawIndex}
		if err := macro.Validate(o, now); err != nil {
			return err
		}
		out = append(out, o)
		return nil
	}
	for _, period := range periods {
		row := selected[period]
		if kind == "credit" {
			if err := add(row, "VN_CREDIT_GROWTH", "SBV_TOTAL_CREDIT_YTD", "SOURCE_TOTAL_CREDIT_YTD_NOT_YOY", row.YTD); err != nil {
				return nil, nil, warnings, err
			}
			continue
		}
		for _, spec := range []struct {
			Code, Series, Method string
			Value                float64
		}{{"VN_M2_LEVEL", "SBV_M2_LEVEL_FROM_2025_10", "SOURCE_M2_LEVEL; METHODOLOGY_FROM_2025_10", row.Level}, {"VN_M2_YTD", "SBV_M2_YTD_FROM_2025_10", "SOURCE_M2_YTD; METHODOLOGY_FROM_2025_10", row.YTD}} {
			if spec.Code == "VN_M2_YTD" && !row.HasYTD {
				continue
			}
			if err := add(row, spec.Code, spec.Series, spec.Method, spec.Value); err != nil {
				return nil, nil, warnings, err
			}
		}
		date, _ := time.Parse("2006-01", period)
		previousPeriod := date.AddDate(-1, 0, 0).Format("2006-01")
		if previous, ok := selected[previousPeriod]; ok && previousPeriod >= M2MethodStart {
			if err := add(row, "VN_M2_GROWTH", "SBV_M2_YOY_FROM_2025_10", "YOY_FROM_COMPARABLE_M2_LEVELS; METHODOLOGY_FROM_2025_10; DENOMINATOR_ARTICLE_ID="+previous.Article.ArticleID, (row.Level/previous.Level-1)*100); err != nil {
				return nil, nil, warnings, err
			}
		}
	}
	if len(periods) == 0 {
		return nil, nil, warnings, fmt.Errorf("no verified month-end SBV totals")
	}
	if kind == "m2" {
		latest := periods[len(periods)-1]
		if !selected[latest].HasYTD {
			reasons["VN_M2_YTD"] = "NOT_COMPARABLE: NHNN không công bố mức YTD có ý nghĩa so sánh cho kỳ này do thay đổi phương pháp."
		}
		hasLatestYoY := false
		for _, o := range out {
			if o.IndicatorCode == "VN_M2_GROWTH" && o.Period == latest {
				hasLatestYoY = true
			}
		}
		if !hasLatestYoY {
			reasons["VN_M2_GROWTH"] = "NOT_COMPARABLE: Chưa có cùng tháng năm trước theo phương pháp M2 từ 10/2025; không tính YoY qua mốc thay đổi phương pháp."
		}
	}
	return out, reasons, warnings, nil
}
