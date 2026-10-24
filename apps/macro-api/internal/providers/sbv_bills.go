package providers

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/trungvdn/investment-ai/apps/macro-api/internal/macro"
	htmlparse "golang.org/x/net/html"
)

const SBVOMOURL = "https://sbv.gov.vn/vi/nghiệp-vụ-thị-trường-mở"

type SBVBills struct {
	HTTP
	URL string
}

func (p *SBVBills) ID() string      { return "sbv:bills" }
func (p *SBVBills) Codes() []string { return []string{"VN_SBV_TBILL"} }
func (p *SBVBills) Fetch(ctx context.Context) (macro.Batch, error) {
	u := p.URL
	if u == "" {
		u = SBVOMOURL
	}
	d, err := p.Get(ctx, u, u, "SBV")
	b := macro.Batch{Document: d}
	if err != nil {
		return b, err
	}
	b.Observations, b.Unavailable, err = ParseSBVBills(d.Body, u, time.Now())
	return b, err
}
func hasClass(n *htmlparse.Node, want string) bool {
	for _, a := range n.Attr {
		if a.Key == "class" {
			for _, c := range strings.Fields(a.Val) {
				if c == want {
					return true
				}
			}
		}
	}
	return false
}
func ParseSBVBills(body []byte, sourceURL string, now time.Time) ([]macro.Observation, map[string]string, error) {
	root, err := htmlparse.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil, nil, fmt.Errorf("invalid SBV OMO HTML")
	}
	var scope *htmlparse.Node
	for _, n := range nodes(root, "div") {
		if hasClass(n, "ls01-wrap") {
			if scope != nil {
				return nil, nil, fmt.Errorf("ambiguous OMO results")
			}
			scope = n
		}
	}
	if scope == nil || !strings.Contains(fold(nodeText(scope)), "ket qua dau thau thi truong mo") {
		return nil, nil, fmt.Errorf("missing official OMO results block")
	}
	var dates []string
	for _, n := range nodes(scope, "div") {
		if hasClass(n, "ls01-date") {
			dates = append(dates, nodeText(n))
		}
	}
	if len(dates) != 1 {
		return nil, nil, fmt.Errorf("missing or ambiguous OMO result date")
	}
	m := regexp.MustCompile(`(?i)ngay (\d{1,2}) thang (\d{1,2}) nam (\d{4})`).FindStringSubmatch(fold(dates[0]))
	if len(m) != 4 {
		return nil, nil, fmt.Errorf("invalid OMO result date")
	}
	day, _ := strconv.Atoi(m[1])
	month, _ := strconv.Atoi(m[2])
	year, _ := strconv.Atoi(m[3])
	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	if date.Day() != day || int(date.Month()) != month {
		return nil, nil, fmt.Errorf("invalid OMO calendar date")
	}
	local := now.In(time.FixedZone("ICT", 7*3600))
	today := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC)
	if date.After(today) {
		return nil, nil, fmt.Errorf("future OMO result date")
	}
	tables := nodes(scope, "table")
	if len(tables) != 1 {
		return nil, nil, fmt.Errorf("ambiguous OMO results table")
	}
	table := tables[0]
	header := fold(nodeText(table))
	if !strings.Contains(header, "khoi luong trung thau") || !strings.Contains(header, "ty dong") || !strings.Contains(header, "lai suat trung thau") || !strings.Contains(header, "%/nam") {
		return nil, nil, fmt.Errorf("unexpected OMO columns or units")
	}
	active := false
	volume, weighted := 0.0, 0.0
	tenors := []string{}
	seen := map[string]bool{}
	for _, r := range rows(table) {
		if len(r) == 1 {
			label := fold(r[0])
			active = (strings.Contains(label, "ban tin phieu nhnn") || strings.Contains(label, "ban tin phieu ngan hang nha nuoc")) && !strings.Contains(label, "kho bac")
			continue
		}
		if !active || len(r) != 4 || strings.Contains(fold(r[0]), "tong cong") {
			continue
		}
		tenor := regexp.MustCompile(`(?i)ky han (\d+) ngay`).FindStringSubmatch(fold(r[0]))
		if len(tenor) != 2 {
			return nil, nil, fmt.Errorf("unrecognized SBV bill tenor")
		}
		if seen[tenor[1]] {
			return nil, nil, fmt.Errorf("duplicate SBV bill tenor")
		}
		seen[tenor[1]] = true
		amount := strings.TrimSpace(r[2])
		if strings.HasPrefix(amount, ",") {
			amount = "0" + amount
		}
		v, e := vietnamNumber(amount)
		if e != nil || v < 0 {
			return nil, nil, fmt.Errorf("invalid SBV bill awarded volume")
		}
		if v == 0 {
			continue
		} // No award has no yield, even if source prints zero.
		rate, e := vietnamNumber(r[3])
		if e != nil || rate < 0 || rate > 100 {
			return nil, nil, fmt.Errorf("missing or invalid SBV bill awarded rate")
		}
		volume += v
		weighted += v * rate
		tenors = append(tenors, tenor[1]+"d")
	}
	if volume == 0 {
		return nil, map[string]string{"VN_SBV_TBILL": "NO_PUBLISHED_RATE: Kết quả NHNN ngày " + date.Format("02/01/2006") + " không có dòng tín phiếu NHNN trúng thầu có lãi suất; không suy ra lãi suất 0 hoặc khối lượng lưu hành bằng 0."}, nil
	}
	o := macro.Observation{IndicatorCode: "VN_SBV_TBILL", Period: date.Format("2006-01-02"), PeriodEnd: date, Value: weighted / volume, Unit: "%", Source: "SBV", SourceURL: sourceURL, SeriesID: "SBV_BILLS_AWARDED_VOLUME_WEIGHTED_RATE", Method: fmt.Sprintf("AWARDED_VOLUME_WEIGHTED_RATE; TENORS=%s; AWARDED_BILLION_VND=%g; NOT_REPO_NOT_TREASURY_NOT_OUTSTANDING", strings.Join(tenors, ","), volume), Priority: 100}
	if err = macro.Validate(o, now); err != nil {
		return nil, nil, err
	}
	return []macro.Observation{o}, nil, nil
}
