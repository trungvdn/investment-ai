package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestSBVFXScopeAndEffectiveDate(t *testing.T) {
	html := `<div class="bi01-wrap">Tỷ giá trung tâm áp dụng cho ngày 23/10/2026 như sau:<table><tr><th>Tỷ giá trung tâm</th><th>Tỷ giá</th></tr><tr><td>1 Đô la Mỹ =</td><td>25.627 VND</td></tr><tr><td>Ngày ban hành</td><td>22/10/2026</td></tr></table></div><div class="bi01-wrap">Tỷ giá mua bán ngày 22/10/2026<table><tr><td>USD</td><td>24.396,00</td><td>26.858,00</td></tr></table></div><script>fake='<table><tr><td>1 Đô la Mỹ =</td><td>1</td></tr></table>'</script>`
	obs, err := ParseSBVPage([]byte(html), "fx", SBVFXURL)
	if err != nil {
		t.Fatal(err)
	}
	if len(obs) != 1 || obs[0].Value != 25627 || obs[0].Period != "2026-10-23" || obs[0].PublishedAt == nil || obs[0].PublishedAt.Day() != 22 {
		t.Fatalf("wrong central rate/date %+v", obs)
	}
	if _, err := ParseSBVPage([]byte(strings.Replace(html, "1 Đô la Mỹ =", "USD", 1)), "fx", SBVFXURL); err == nil {
		t.Fatal("buy/sell rate used as central rate")
	}
	if _, err := ParseSBVPage([]byte(html+html), "fx", SBVFXURL); err == nil {
		t.Fatal("ambiguous central rate accepted")
	}
}
func TestSBVOvernightNotPolicyRateOrOtherTenor(t *testing.T) {
	html := `<div class="bi01-wrap">Lãi suất điều hành ngày 01/01/2026<table><tr><td>Lãi suất tái cấp vốn</td><td>4,5</td></tr></table></div><div class="bi01-wrap">Lãi suất thị trường liên ngân hàng. Ngày áp dụng: 22/10/2026<table><tr><th>Thời hạn</th><th>Lãi suất BQ liên Ngân hàng (% năm)</th><th>Doanh số</th></tr><tr><td>1 Tuần</td><td>4,35</td><td>100</td></tr><tr><td>Qua đêm</td><td><span>4,31</span><span class="bi01-footnote"></span></td><td>835.912,0</td></tr></table></div>`
	out, err := ParseSBVPage([]byte(html), "interbank", SBVInterbankURL)
	if err != nil || len(out) != 1 || out[0].Value != 4.31 || out[0].Period != "2026-10-22" {
		t.Fatalf("wrong O/N rate %+v %v", out, err)
	}
	if _, err = ParseSBVPage([]byte(strings.ReplaceAll(html, "Ngày áp dụng: 22/10/2026", "")), "interbank", SBVInterbankURL); err == nil {
		t.Fatal("missing effective date accepted")
	}
}
func moneyArticle(period, level, growth string) sbvArticle {
	end, _ := time.Parse("2006-01", period)
	date := end.AddDate(0, 1, -1)
	return sbvArticle{ArticleID: period, Status: 0, Structure: 10038520, Title: "Tháng " + end.Format("1") + " năm " + end.Format("2006"), Modified: "2026-10-01T12:00:00+07:00", Fields: map[string]any{"Date00760329": date.Format("2006-01-02"), "soDuTongPhuongTienThanhToan": level, "tocDoTangGiamTongPhuongTienThanhToan": growth, "ghiChu": "Số liệu bắt đầu từ kỳ báo cáo tháng 10/2025 có sự thay đổi"}}
}
func TestM2MethodBreakAndPublishedYTD(t *testing.T) {
	a := moneyArticle("2025-08", "19.685.504", "6,10")
	b := moneyArticle("2026-08", "20.774.696", "6,84")
	out, reasons, _, err := parseStatistics([]sbvArticle{a, b}, "m2")
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 2 || out[0].IndicatorCode != "VN_M2_LEVEL" || out[0].Value != 20774696 || out[1].Value != 6.84 || !strings.HasPrefix(reasons["VN_M2_GROWTH"], "NOT_COMPARABLE:") {
		t.Fatalf("method break or YTD mishandled: %+v %+v", out, reasons)
	}
	// Same-method October pair computes YoY once a full comparable year exists.
	a = moneyArticle("2025-10", "18.000.000", "5,00")
	b = moneyArticle("2026-10", "19.800.000", "6,00")
	out, _, _, err = parseStatisticsAt([]sbvArticle{a, b}, "m2", time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, o := range out {
		if o.IndicatorCode == "VN_M2_GROWTH" && o.Period == "2026-10" {
			found = true
			if math.Abs(o.Value-10) > .000001 {
				t.Fatal("same-method YoY incorrect")
			}
		}
	}
	if !found {
		t.Fatal("valid same-method YoY was omitted")
	}
	if _, _, _, err = parseStatisticsAt([]sbvArticle{a, b}, "m2", time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC)); err == nil {
		t.Fatal("future month accepted")
	}
	for _, bad := range []string{"", "N/A", "20.77", "NaN"} {
		b = moneyArticle("2026-08", bad, "6,84")
		if _, _, _, err = parseStatistics([]sbvArticle{b}, "m2"); err == nil {
			t.Fatal("invalid level became zero")
		}
	}
}
func TestM2ConflictsAndCreditTotals(t *testing.T) {
	a := moneyArticle("2026-08", "20.774.696", "6,84")
	b := moneyArticle("2026-08", "20.000.000", "6,84")
	if _, _, _, err := parseStatistics([]sbvArticle{a, b}, "m2"); err == nil {
		t.Fatal("conflicting vintage accepted")
	}
	credit := sbvArticle{ArticleID: "credit-aug", Status: 0, Structure: 10034332, Title: "Tháng 8 năm 2026", Fields: map[string]any{"Date62813077": "2026-08-31", "soDuTongCong": "20.489.434", "tocDoTangGiamTongCong": "10,19", "tocDoTangGiamNongNghiep": "9,77"}}
	out, _, _, err := parseStatistics([]sbvArticle{credit}, "credit")
	if err != nil || len(out) != 1 || out[0].Value != 10.19 || out[0].Period != "2026-08" || out[0].Unit != "% YTD" {
		t.Fatalf("sector credit or YoY substituted: %+v %v", out, err)
	}
	credit.Fields["Date62813077"] = "2026-09-01"
	if _, _, _, err := parseStatistics([]sbvArticle{credit}, "credit"); err == nil {
		t.Fatal("title/date mismatch accepted")
	}
}
func TestSBVPaginationRetainsEveryRawPage(t *testing.T) {
	one := moneyArticle("2026-07", "20.455.644", "5,20")
	two := moneyArticle("2026-08", "20.774.696", "6,84")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("contentStructureId") != "10038520" {
			t.Error("wrong statistics structure")
		}
		page := 1
		item := one
		if r.URL.Query().Get("page") == "2" {
			page = 2
			item = two
		}
		json.NewEncoder(w).Encode(map[string]any{"articles": []sbvArticle{item}, "page": page, "lastPage": 2, "totalCount": 2})
	}))
	defer server.Close()
	provider := SBVStatistics{Kind: "m2", Endpoint: server.URL}
	batch, err := provider.Fetch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(batch.Documents) != 1 || len(batch.Document.Body) == 0 || len(batch.Observations) != 4 || batch.Observations[2].RawIndex != 1 {
		t.Fatalf("pagination/provenance failed %+v", batch)
	}
}
func TestQuarterlyPPIIsAlreadySameQuarterPreviousYearIndex(t *testing.T) {
	series := `<Series REF_AREA="VN" COUNTERPART_AREA="_Z" FREQ="Q" BASE_PER="_Z" UNIT_MULT="0" INDICATOR="` + PPISeries + `"><Obs TIME_PERIOD="2025-Q3" OBS_VALUE="103"/><Obs TIME_PERIOD="2026-Q3" OBS_VALUE="104.75115176098"/></Series>`
	body := []byte(`<Root><DataSet><Series INDICATOR="PPPI_ISIC4_A_BY_PP_IX"><Obs TIME_PERIOD="2026-Q3" OBS_VALUE="chênh"/></Series>` + series + `</DataSet></Root>`)
	out, err := ParsePPI(body, NSDPBase+"PPIVNM.xml")
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 2 || math.Abs(out[1].Value-4.75115176098) > .000001 || out[1].PeriodEnd.Format("2006-01-02") != "2026-09-30" {
		t.Fatalf("PPI incorrectly double-transformed: %+v", out)
	}
	if _, err = ParsePPI(bytes.ReplaceAll(body, []byte(`OBS_VALUE="104.75115176098"`), []byte(`OBS_VALUE="chênh"`)), "source"); err == nil {
		t.Fatal("invalid aggregate PPI accepted")
	}
}
func TestSBVPPILiveSmoke(t *testing.T) {
	if os.Getenv("SBV_PPI_LIVE_SMOKE") != "1" {
		t.Skip("set SBV_PPI_LIVE_SMOKE=1 for official-source check")
	}
	for _, provider := range []Provider{&SBVPage{Kind: "fx"}, &SBVPage{Kind: "interbank"}, &SBVStatistics{Kind: "m2"}, &SBVStatistics{Kind: "credit"}, &PPI{}} {
		t.Run(provider.ID(), func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			b, err := provider.Fetch(ctx)
			if err != nil {
				t.Fatal(err)
			}
			latest := map[string]int{}
			for j, o := range b.Observations {
				if i, ok := latest[o.IndicatorCode]; !ok || o.Period > b.Observations[i].Period {
					latest[o.IndicatorCode] = j
				}
			}
			for code, i := range latest {
				o := b.Observations[i]
				t.Logf("%s %s = %.6f %s", code, o.Period, o.Value, o.Unit)
			}
			for code, reason := range b.Unavailable {
				t.Log(code, reason)
			}
		})
	}
}

func TestM2BlankHistoricalYTDIsNotZeroAndLatestVersionWins(t *testing.T) {
	old := moneyArticle("2025-12", "19.444.476", "")
	old.Version = 1.1
	old.Fields["ghiChu"] = "Phân tổ lại một số khu vực thể chế; loại giấy tờ có giá do các TCTD khác mua; cập nhật tiêu chuẩn quốc tế; không có ý nghĩa so sánh do phương pháp thống kê thay đổi."
	current := old
	current.Version = 1.2
	out, reasons, _, err := parseStatistics([]sbvArticle{old, current}, "m2")
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].IndicatorCode != "VN_M2_LEVEL" || !strings.HasPrefix(reasons["VN_M2_YTD"], "NOT_COMPARABLE:") {
		t.Fatalf("blank YTD fabricated or version selection failed %+v %+v", out, reasons)
	}
}

func TestHTTPRetriesTruncatedBodyWithoutExposingRequest(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt := calls.Add(1)
		if attempt == 1 {
			w.Header().Set("Content-Length", "100")
			_, _ = w.Write([]byte("{"))
			return
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	doc, err := (HTTP{}).Get(ctx, server.URL+"?private=test-secret", "https://sbv.gov.vn/public-source", "SBV")
	if err != nil || calls.Load() != 2 || string(doc.Body) != `{"ok":true}` || strings.Contains(doc.URL, "test-secret") {
		t.Fatalf("truncated-body retry or safe provenance failed: %d %v", calls.Load(), err)
	}
}
