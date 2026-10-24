package providers

import (
	"context"
	"fmt"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/macro"
	htmlparse "golang.org/x/net/html"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const SBVFXURL = "https://sbv.gov.vn/vi/tỷ-giá"
const SBVInterbankURL = "https://sbv.gov.vn/vi/lãi-suất1"
const SBVM2URL = "https://sbv.gov.vn/vi/tổng-phương-tiện-thanh-toán-và-tiền-gửi-của-khách-hàng-tại-tctd"
const SBVCreditURL = "https://sbv.gov.vn/vi/du-no-tin-dung-doi-voi-nen-kt-dttktt"

type SBVPage struct {
	HTTP
	Kind, URL string
}

func (s *SBVPage) ID() string { return "sbv:" + s.Kind }
func (s *SBVPage) Codes() []string {
	if s.Kind == "fx" {
		return []string{"VN_USD_VND"}
	}
	return []string{"VN_INTERBANK_ON"}
}
func (s *SBVPage) Fetch(ctx context.Context) (macro.Batch, error) {
	u := s.URL
	if u == "" {
		if s.Kind == "fx" {
			u = SBVFXURL
		} else {
			u = SBVInterbankURL
		}
	}
	doc, err := s.Get(ctx, u, u, "SBV")
	if err != nil {
		return macro.Batch{}, err
	}
	obs, err := ParseSBVPage(doc.Body, s.Kind, u)
	return macro.Batch{Document: doc, Observations: obs}, err
}
func nodeText(n *htmlparse.Node) string {
	if n.Type == htmlparse.TextNode {
		return n.Data
	}
	if n.Type == htmlparse.ElementNode && (n.Data == "script" || n.Data == "style") {
		return ""
	}
	var b strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		b.WriteString(nodeText(c))
		b.WriteByte(' ')
	}
	return strings.Join(strings.Fields(b.String()), " ")
}
func nodes(root *htmlparse.Node, tag string) []*htmlparse.Node {
	out := []*htmlparse.Node{}
	var visit func(*htmlparse.Node)
	visit = func(n *htmlparse.Node) {
		if n.Type == htmlparse.ElementNode && (n.Data == "script" || n.Data == "style") {
			return
		}
		if n.Type == htmlparse.ElementNode && n.Data == tag {
			out = append(out, n)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			visit(c)
		}
	}
	visit(root)
	return out
}
func rows(table *htmlparse.Node) [][]string {
	out := [][]string{}
	for _, tr := range nodes(table, "tr") {
		cells := []string{}
		for c := tr.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == htmlparse.ElementNode && (c.Data == "td" || c.Data == "th") {
				cells = append(cells, nodeText(c))
			}
		}
		if len(cells) > 0 {
			out = append(out, cells)
		}
	}
	return out
}
func scopeText(table *htmlparse.Node) string {
	for n := table.Parent; n != nil; n = n.Parent {
		for _, a := range n.Attr {
			if a.Key == "class" {
				for _, class := range strings.Fields(a.Val) {
					if class == "bi01-wrap" {
						return fold(nodeText(n))
					}
				}
			}
		}
	}
	return ""
}
func vietnamNumber(value string) (float64, error) {
	s := strings.TrimSpace(value)
	if !regexp.MustCompile(`^[+-]?(?:\d{1,3}(?:\.\d{3})+|\d+)(?:,\d+)?$`).MatchString(s) {
		return 0, fmt.Errorf("invalid Vietnamese numeric value")
	}
	v, err := strconv.ParseFloat(strings.ReplaceAll(strings.ReplaceAll(s, ".", ""), ",", "."), 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, fmt.Errorf("nonfinite numeric value")
	}
	return v, nil
}
func dayFromMatch(context, pattern string) (time.Time, error) {
	matches := regexp.MustCompile(pattern).FindAllStringSubmatch(context, -1)
	if len(matches) != 1 || len(matches[0]) < 2 {
		return time.Time{}, fmt.Errorf("missing or ambiguous SBV effective date")
	}
	return time.Parse("02/01/2006", matches[0][1])
}
func ParseSBVPage(body []byte, kind, sourceURL string) ([]macro.Observation, error) {
	root, err := htmlparse.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("invalid SBV HTML")
	}
	out := []macro.Observation{}
	for _, table := range nodes(root, "table") {
		scoped := scopeText(table)
		if scoped == "" {
			continue
		}
		var period time.Time
		var value float64
		var pub *time.Time
		matched := false
		series, code, method := "SBV_CENTRAL_USD_VND", "VN_USD_VND", "SOURCE_CENTRAL_RATE; PERIOD=EFFECTIVE_DATE; PUBLICATION_DATE_ONLY"
		if kind == "fx" {
			for _, r := range rows(table) {
				if len(r) == 2 && fold(r[0]) == "1 do la my =" {
					if matched {
						return nil, fmt.Errorf("duplicate central FX row")
					}
					value, err = vietnamNumber(strings.TrimSpace(strings.TrimSuffix(fold(r[1]), "vnd")))
					if err != nil {
						return nil, err
					}
					matched = true
				}
				if len(r) == 2 && fold(r[0]) == "ngay ban hanh" {
					p, e := time.ParseInLocation("02/01/2006", strings.TrimSpace(r[1]), time.FixedZone("ICT", 7*3600))
					if e != nil {
						return nil, fmt.Errorf("invalid issue date")
					}
					pub = &p
				}
			}
			if !matched {
				continue
			}
			period, err = dayFromMatch(scoped, `ap dung cho ngay (\d{2}/\d{2}/\d{4})`)
		} else {
			if !strings.Contains(fold(nodeText(table)), "lai suat bq lien ngan hang (% nam)") {
				continue
			}
			for _, r := range rows(table) {
				if len(r) >= 2 && fold(r[0]) == "qua dem" {
					if matched {
						return nil, fmt.Errorf("duplicate O/N row")
					}
					value, err = vietnamNumber(r[1])
					if err != nil {
						return nil, err
					}
					matched = true
				}
			}
			if !matched {
				continue
			}
			period, err = dayFromMatch(scoped, `ngay ap dung: (\d{2}/\d{2}/\d{4})`)
			series, code, method = "SBV_VND_INTERBANK_OVERNIGHT", "VN_INTERBANK_ON", "SOURCE_VND_OVERNIGHT_WEIGHTED_AVERAGE_RATE; PERIOD=TABLE_EFFECTIVE_DATE"
		}
		if err != nil {
			return nil, err
		}
		if value <= 0 && kind == "fx" {
			return nil, fmt.Errorf("nonpositive central FX")
		}
		ind, _ := macro.Find(code)
		o := macro.Observation{IndicatorCode: code, Period: period.Format("2006-01-02"), PeriodEnd: period, Value: value, Unit: ind.Unit, Source: "SBV", SourceURL: sourceURL, SeriesID: series, Method: method, Priority: 100, PublishedAt: pub}
		if err := macro.Validate(o, time.Now()); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	if len(out) != 1 {
		return nil, fmt.Errorf("expected exactly one SBV %s observation, got %d", kind, len(out))
	}
	return out, nil
}
