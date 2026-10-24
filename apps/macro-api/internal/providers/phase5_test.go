package providers

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

var phase5Now = time.Date(2026, 10, 24, 5, 0, 0, 0, time.UTC)

const pmiRelease = "https://www.pmi.spglobal.com/Public/Home/PressRelease/dba8ca68ce13403591b68cb3e07175d9"
const pmiText = `Embargoed until 1 October 2026 (0730 ICT / 0030 UTC)
S&P Global Vietnam Manufacturing PMI September 2026
The S&P Global Vietnam Manufacturing Purchasing Managers' Index (PMI) dropped to 51.9 in September from 53.3 in August. Output index 55.4. The 50.0 threshold signals no change.`

func TestPMIHeadlineNotPreviousOrSubindex(t *testing.T) {
	o, err := ParsePMIText(pmiText, pmiRelease, phase5Now)
	if err != nil || o.Value != 51.9 || o.Period != "2026-09" || o.PublishedAt == nil || o.PublishedAt.UTC().Format(time.RFC3339) != "2026-10-01T00:30:00Z" {
		t.Fatalf("incorrect headline: %+v %v", o, err)
	}
	for _, bad := range []string{strings.ReplaceAll(pmiText, "Vietnam", "Thailand"), strings.ReplaceAll(pmiText, "September 2026", "August 2026"), strings.ReplaceAll(pmiText, "Embargoed until", "Published"), pmiText + " " + pmiText, strings.ReplaceAll(pmiText, "51.9", "101.2")} {
		if _, err = ParsePMIText(bad, pmiRelease, phase5Now); err == nil {
			t.Fatal("ambiguous/mismatched PMI accepted")
		}
	}
	if _, err = ParsePMIText(pmiText, pmiRelease, time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)); err == nil {
		t.Fatal("embargoed PMI accepted")
	}
	dec := strings.NewReplacer("1 October 2026", "2 January 2026", "September 2026", "December 2025", "in September", "in December").Replace(pmiText)
	o, err = ParsePMIText(dec, pmiRelease, phase5Now)
	if err != nil || o.Period != "2025-12" {
		t.Fatalf("December rollover: %+v %v", o, err)
	}
}
func omoTable(date, body string) string {
	return `<div class="ls01-wrap"><h4>KẾT QUẢ ĐẤU THẦU THỊ TRƯỜNG MỞ</h4><div class="ls01-date">` + date + `</div><table><tr><th>Loại hình giao dịch</th><th>Số thành viên tham gia/trúng thầu</th><th>Khối lượng trúng thầu (Tỷ đồng)</th><th>Lãi suất trúng thầu (%/năm)</th></tr>` + body + `</table></div>`
}

const repoRow = `<tr><td colspan="4">Mua kỳ hạn</td></tr><tr><td>- Kỳ hạn 14 ngày</td><td>1/1</td><td>756,77</td><td>4,5</td></tr>`
const billsRows = `<tr><td colspan="4">Bán tín phiếu NHNN</td></tr><tr><td>- Kỳ hạn 7 ngày</td><td>1/1</td><td>1.000,00</td><td>3,5</td></tr><tr><td>- Kỳ hạn 28 ngày</td><td>2/2</td><td>3.000,00</td><td>4,0</td></tr><tr><td>Tổng cộng</td><td></td><td>4.000,00</td><td></td></tr>`

func TestBillsWeightedRateExcludesRepoAndTreasury(t *testing.T) {
	html := omoTable("Ngày 23 tháng 10 năm 2026", repoRow+billsRows+`<tr><td colspan="4">Tín phiếu Kho bạc</td></tr><tr><td>- Kỳ hạn 7 ngày</td><td>1/1</td><td>9.000,00</td><td>9,0</td></tr>`)
	obs, unavailable, err := ParseSBVBills([]byte(html), SBVOMOURL, phase5Now)
	if err != nil || len(obs) != 1 || len(unavailable) != 0 || math.Abs(obs[0].Value-3.875) > 1e-9 || obs[0].Period != "2026-10-23" || !strings.Contains(obs[0].Method, "TENORS=7d,28d") {
		t.Fatalf("bad bills aggregate %+v %+v %v", obs, unavailable, err)
	}
	for _, bad := range []string{strings.Replace(html, "3,5", "", 1), strings.Replace(html, "28 ngày", "7 ngày", 1), strings.Replace(html, "23 tháng 10", "31 tháng 2", 1), html + html, strings.Replace(html, "23 tháng 10 năm 2026", "24 tháng 10 năm 2027", 1)} {
		if _, _, err = ParseSBVBills([]byte(bad), SBVOMOURL, phase5Now); err == nil {
			t.Fatal("invalid bill result accepted")
		}
	}
}
func TestBillsMissingAuctionIsNotZero(t *testing.T) {
	for _, body := range []string{repoRow, repoRow + strings.NewReplacer("1.000,00", ",00", "3.000,00", "0,00").Replace(billsRows)} {
		obs, reasons, err := ParseSBVBills([]byte(omoTable("Ngày 23 tháng 10 năm 2026", body)), SBVOMOURL, phase5Now)
		if err != nil || len(obs) != 0 || !strings.HasPrefix(reasons["VN_SBV_TBILL"], "NO_PUBLISHED_RATE:") {
			t.Fatalf("fabricated missing rate %+v %+v %v", obs, reasons, err)
		}
	}
	if _, _, err := ParseSBVBills([]byte(`<html>Access denied</html>`), SBVOMOURL, phase5Now); err == nil {
		t.Fatal("access denial treated as no auction")
	}
}

type phase5Transport func(*http.Request) (*http.Response, error)

func (f phase5Transport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func minimalPMIPDF() []byte {
	var content strings.Builder
	content.WriteString("BT /F1 10 Tf 50 750 Td ")
	for i, line := range strings.Split(pmiText, "\n") {
		if i > 0 {
			content.WriteString("0 -20 Td ")
		}
		escaped := strings.NewReplacer("\\", "\\\\", "(", "\\(", ")", "\\)").Replace(line)
		fmt.Fprintf(&content, "(%s) Tj ", escaped)
	}
	content.WriteString("ET")
	objs := []string{`<< /Type /Catalog /Pages 2 0 R >>`, `<< /Type /Pages /Kids [3 0 R] /Count 1 >>`, `<< /Type /Page /Parent 2 0 R /MediaBox [0 0 2000 800] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>`, `<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>`, fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", content.Len(), content.String())}
	var pdf strings.Builder
	pdf.WriteString("%PDF-1.4\n")
	offsets := []int{0}
	for i, obj := range objs {
		offsets = append(offsets, pdf.Len())
		fmt.Fprintf(&pdf, "%d 0 obj\n%s\nendobj\n", i+1, obj)
	}
	xref := pdf.Len()
	fmt.Fprintf(&pdf, "xref\n0 %d\n0000000000 65535 f \n", len(offsets))
	for _, v := range offsets[1:] {
		fmt.Fprintf(&pdf, "%010d 00000 n \n", v)
	}
	fmt.Fprintf(&pdf, "trailer << /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(offsets), xref)
	return []byte(pdf.String())
}
func TestPMIPDFIngestionAndSourceBoundary(t *testing.T) {
	if _, err := exec.LookPath("pdftotext"); err != nil {
		t.Skip("install poppler-utils to test PDF extraction")
	}
	doc := minimalPMIPDF()
	p := &PMI{HTTP: HTTP{Client: &http.Client{Transport: phase5Transport(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/pdf"}}, Body: io.NopCloser(strings.NewReader(string(doc)))}, nil
	})}}, ReleaseURLs: []string{pmiRelease}}
	b, err := p.Fetch(context.Background())
	if err != nil || len(b.Observations) != 1 || len(b.Document.Body) == 0 {
		t.Fatalf("PDF ingestion %+v %v", b, err)
	}
	p.ReleaseURLs = append(p.ReleaseURLs, pmiRelease)
	if _, err = p.Fetch(context.Background()); err == nil {
		t.Fatal("duplicate months accepted")
	}
	p.ReleaseURLs = []string{"https://example.org/release"}
	if _, err = p.Fetch(context.Background()); err == nil {
		t.Fatal("untrusted source accepted")
	}
	p.ReleaseURLs = nil
	if _, err = p.Fetch(context.Background()); err == nil {
		t.Fatal("unrecognized discovery index was accepted")
	}
}
func TestPhase5LiveSmoke(t *testing.T) {
	if os.Getenv("PHASE5_LIVE_SMOKE") != "1" {
		t.Skip("set PHASE5_LIVE_SMOKE=1")
	}
	providers := []Provider{&SBVBills{}}
	if u := os.Getenv("PMI_RELEASE_URLS"); u != "" {
		providers = append(providers, &PMI{ReleaseURLs: strings.Split(u, ",")})
	}
	for _, p := range providers {
		t.Run(p.ID(), func(t *testing.T) {
			b, err := p.Fetch(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("observations=%+v unavailable=%+v", b.Observations, b.Unavailable)
		})
	}
}
