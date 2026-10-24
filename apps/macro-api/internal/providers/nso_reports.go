package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/macro"
	htmlparse "golang.org/x/net/html"
	"golang.org/x/text/unicode/norm"
	"html"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type Reports struct {
	HTTP
	Kind, URL string
}

func (r *Reports) ID() string { return "nso:reports:" + r.Kind }
func (r *Reports) Codes() []string {
	if r.Kind == "prices" {
		return []string{"VN_CORE_CPI_YOY"}
	}
	return []string{"VN_CPI_YOY", "VN_RETAIL_YOY", "VN_EXPORT_YOY"}
}
func (r *Reports) Fetch(ctx context.Context) (macro.Batch, error) {
	u := r.URL
	if u == "" {
		phrase := "\"Báo cáo tình hình kinh tế\""
		if r.Kind == "prices" {
			phrase = "\"Thông cáo báo chí về tình hình giá\""
		}
		q := url.Values{"search": {phrase}, "per_page": {"100"}, "_fields": {"id,date,link,title,content"}}
		u = "https://www.nso.gov.vn/wp-json/wp/v2/posts?" + q.Encode()
	}
	doc, err := r.Get(ctx, u, u, "NSO")
	if err != nil {
		return macro.Batch{}, err
	}
	obs, warnings, err := ParseReports(doc.Body, r.Kind)
	return macro.Batch{Document: doc, Observations: obs, Warnings: warnings}, err
}

type post struct {
	Date  string `json:"date"`
	Link  string `json:"link"`
	Title struct {
		Rendered string `json:"rendered"`
	} `json:"title"`
	Content struct {
		Rendered string `json:"rendered"`
	} `json:"content"`
}

func fold(s string) string {
	s = strings.ToLower(norm.NFD.String(html.UnescapeString(s)))
	s = strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Mn, r) {
			return -1
		}
		if r == 'đ' {
			return 'd'
		}
		return r
	}, s)
	return strings.Join(strings.Fields(s), " ")
}
func paragraphs(input string) []string {
	root, err := htmlparse.Parse(strings.NewReader(input))
	if err != nil {
		return nil
	}
	out := []string{}
	var text func(*htmlparse.Node) string
	text = func(n *htmlparse.Node) string {
		if n.Type == htmlparse.TextNode {
			return n.Data
		}
		if n.Type == htmlparse.ElementNode && (n.Data == "script" || n.Data == "style") {
			return ""
		}
		var b strings.Builder
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			b.WriteString(text(c))
			b.WriteByte(' ')
		}
		return b.String()
	}
	var walk func(*htmlparse.Node)
	walk = func(n *htmlparse.Node) {
		if n.Type == htmlparse.ElementNode && n.Data == "p" {
			out = append(out, fold(text(n)))
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(root)
	return out
}

var monthNames = map[string]int{"mot": 1, "gieng": 1, "hai": 2, "ba": 3, "tu": 4, "bon": 4, "nam": 5, "sau": 6, "bay": 7, "tam": 8, "chin": 9, "muoi": 10, "muoi mot": 11, "muoi hai": 12}

func monthNumber(s string) int {
	if v, err := strconv.Atoi(s); err == nil && v >= 1 && v <= 12 {
		return v
	}
	return monthNames[s]
}

const monthPattern = `(?:muoi hai|muoi mot|muoi|gieng|mot|hai|ba|bon|tu|nam|sau|bay|tam|chin|1[012]|[1-9])`

func reportPeriod(title string) (string, error) {
	title = fold(title)
	yearMatch := regexp.MustCompile(`nam (20\d{2})`).FindStringSubmatch(title)
	if len(yearMatch) != 2 {
		return "", fmt.Errorf("report year not explicit")
	}
	year, _ := strconv.Atoi(yearMatch[1])
	month := 0
	direct := regexp.MustCompile(`thang (` + monthPattern + `)(?:\b|/)`).FindStringSubmatch(strings.ReplaceAll(title, yearMatch[0], "year"))
	if len(direct) == 2 {
		month = monthNumber(direct[1])
	}
	if month == 0 {
		cum := regexp.MustCompile(`(` + monthPattern + `) thang (?:dau )?nam`).FindStringSubmatch(title)
		if len(cum) == 2 {
			month = monthNumber(cum[1])
		}
	}
	if month == 0 && strings.Contains(title, "quy iv") {
		month = 12
	}
	if month == 0 {
		return "", fmt.Errorf("report month not explicit")
	}
	return fmt.Sprintf("%04d-%02d", year, month), nil
}
func monthlyRate(paragraph, subject, period string) (float64, bool) {
	start := strings.Index(paragraph, subject)
	if start < 0 {
		return 0, false
	}
	p := paragraph[start:]
	// End before cumulative/quarterly passages; never borrow their YoY values.
	for _, stop := range []string{"binh quan", "tinh chung", "trong quy", "cpi quy", "quy i/", "quy ii/", "quy iii/", "quy iv/"} {
		if i := strings.Index(p, stop); i >= 0 {
			p = p[:i]
		}
	}
	firstMonth := regexp.MustCompile(`thang (` + monthPattern + `)(?:\b|/)`).FindStringSubmatch(regexp.MustCompile(`thang nam 20\d{2}`).ReplaceAllString(p, "month-year"))
	if len(firstMonth) != 2 {
		return 0, false
	}
	n, _ := strconv.Atoi(period[5:])
	if monthNumber(firstMonth[1]) != n {
		return 0, false
	}
	// Reject YTD clauses even when a later sentence contains the desired words.
	if strings.Contains(p, "thang dau nam") {
		return 0, false
	}
	matches := regexp.MustCompile(`(tang|giam)\s+([0-9]+(?:[.,][0-9]+)?)\s*%\s+so voi cung ky(?: nam truoc)?`).FindAllStringSubmatch(p, -1)
	if len(matches) != 1 {
		return 0, false
	}
	v, err := strconv.ParseFloat(strings.ReplaceAll(matches[0][2], ",", "."), 64)
	if err != nil {
		return 0, false
	}
	if matches[0][1] == "giam" {
		v = -v
	}
	return v, true
}
func ParseReports(body []byte, kind string) ([]macro.Observation, []string, error) {
	var posts []post
	if err := json.Unmarshal(body, &posts); err != nil {
		return nil, nil, fmt.Errorf("invalid NSO WordPress JSON")
	}
	sort.Slice(posts, func(i, j int) bool { return posts[i].Date < posts[j].Date })
	results := map[string]macro.Observation{}
	warnings := []string{}
	for _, p := range posts {
		title := fold(p.Title.Rendered)
		prefix := "bao cao tinh hinh kinh te"
		if kind == "prices" {
			prefix = "thong cao bao chi ve tinh hinh gia"
		}
		if !strings.HasPrefix(title, prefix) {
			continue
		}
		// Exclude regional reports and constrain provenance to the official NSO host.
		if strings.Contains(title, "ha noi") || strings.Contains(title, "thanh pho") || strings.Contains(title, "cua tinh ") || strings.Contains(title, "xa hoi tinh ") {
			continue
		}
		link, err := url.Parse(p.Link)
		if err != nil || link.Scheme != "https" || (link.Hostname() != "www.nso.gov.vn" && link.Hostname() != "nso.gov.vn") {
			continue
		}
		period, err := reportPeriod(title)
		if err != nil {
			warnings = append(warnings, "Skipped report without explicit period")
			continue
		}
		end, _ := macro.MonthEnd(period)
		published, err := time.ParseInLocation("2006-01-02T15:04:05", p.Date, time.FixedZone("ICT", 7*3600))
		if err != nil {
			warnings = append(warnings, "Skipped report with invalid publication time")
			continue
		}
		targets := map[string]string{"VN_CPI_YOY": "cpi", "VN_RETAIL_YOY": "tong muc ban le hang hoa va doanh thu dich vu tieu dung", "VN_EXPORT_YOY": "kim ngach xuat khau hang hoa"}
		if kind == "prices" {
			targets = map[string]string{"VN_CORE_CPI_YOY": "lam phat co ban"}
		}
		for code, subject := range targets {
			values := []float64{}
			for _, para := range paragraphs(p.Content.Rendered) {
				if value, ok := monthlyRate(para, subject, period); ok {
					values = append(values, value)
				}
			}
			if len(values) == 0 {
				continue
			}
			v := values[0]
			ambiguous := false
			for _, other := range values[1:] {
				if other != v {
					ambiguous = true
				}
			}
			if ambiguous {
				warnings = append(warnings, "Conflicting monthly values in NSO report for "+code)
				continue
			}
			ind, _ := macro.Find(code)
			o := macro.Observation{IndicatorCode: code, Period: period, PeriodEnd: end, Value: v, Unit: ind.Unit, Source: "NSO", SourceURL: p.Link, SeriesID: "NSO_REPORT_" + kind + "_" + code, Method: "EXPLICIT_MONTHLY_YOY_FROM_OFFICIAL_REPORT", Priority: 110, PublishedAt: &published}
			if err := macro.Validate(o, time.Now()); err != nil {
				warnings = append(warnings, "Invalid observation in report for "+code)
				continue
			}
			results[code+period] = o
		}
	}
	out := []macro.Observation{}
	for _, o := range results {
		out = append(out, o)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Period == out[j].Period {
			return out[i].IndicatorCode < out[j].IndicatorCode
		}
		return out[i].Period < out[j].Period
	})
	if len(out) == 0 {
		return nil, warnings, fmt.Errorf("no unambiguous monthly YoY observations in NSO reports")
	}
	return out, warnings, nil
}
