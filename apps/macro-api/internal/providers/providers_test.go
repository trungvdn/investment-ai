package providers

import (
	"context"
	"encoding/json"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/macro"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestFREDMissingDaysAndMetadata(t *testing.T) {
	data := `{"count":3,"observations":[{"date":"2026-01-01","value":"."},{"date":"2026-01-02","value":"3.50"},{"date":"2026-01-03","value":"3.60"}]}`
	obs, err := ParseFRED([]byte(data), "US_EFFR", "DFF")
	if err != nil {
		t.Fatal(err)
	}
	if len(obs) != 2 || obs[0].Value != 3.5 || obs[0].PublishedAt != nil || obs[0].SeriesID != "DFF" {
		t.Fatalf("invalid FRED observations: %+v", obs)
	}
	for _, bad := range []string{`{"count":2,"observations":[]}`, `{"observations":[{"date":"2026-01-01","value":"NaN"}]}`, `{"observations":[{"date":"bad","value":"1"}]}`, `<html>Error</html>`} {
		if _, err := ParseFRED([]byte(bad), "US_EFFR", "DFF"); err == nil {
			t.Fatal("invalid FRED payload accepted")
		}
	}
}
func TestFREDRequestsAndSecretRedaction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("api_key") != "test-secret" || r.URL.Query().Get("series_id") != "DGS10" || r.URL.Query().Get("file_type") != "json" {
			t.Error("missing FRED query parameters")
		}
		http.Error(w, "bad key test-secret", 400)
	}))
	defer server.Close()
	f := FRED{Code: "US_10Y", Series: "DGS10", APIKey: "test-secret", BaseURL: server.URL}
	_, err := f.Fetch(context.Background())
	if err == nil || strings.Contains(err.Error(), "test-secret") {
		t.Fatal("error must be safe without API key")
	}
	f.APIKey = ""
	if _, err = f.Fetch(context.Background()); err != ErrNotConfigured {
		t.Fatal("missing key must be explicit")
	}
}
func xmlSeries(indicator, base string, rows string) string {
	return `<Series REF_AREA="VN" INDICATOR="` + indicator + `" COUNTERPART_AREA="_Z" FREQ="M" BASE_PER="` + base + `" UNIT_MULT="0">` + rows + `</Series>`
}
func TestSDMXExactSeriesAndYoY(t *testing.T) {
	wanted := xmlSeries("PCPI_IX", "2019", `<Obs TIME_PERIOD="2025-01" OBS_VALUE="100"/><Obs TIME_PERIOD="2026-01" OBS_VALUE="105"/>`)
	wrong := xmlSeries("PCPI_CP_01_IX", "2019", `<Obs TIME_PERIOD="2025-01" OBS_VALUE="200"/><Obs TIME_PERIOD="2026-01" OBS_VALUE="500"/>`)
	input := []byte(`<message:StructureSpecificData xmlns:message="urn:sdmx"><message:DataSet>` + wrong + wanted + `</message:DataSet></message:StructureSpecificData>`)
	obs, err := ParseSDMX(input, "VN_CPI_YOY", "PCPI_IX", "https://nsdp.nso.gov.vn/cpi.xml")
	if err != nil {
		t.Fatal(err)
	}
	if len(obs) != 1 || math.Abs(obs[0].Value-5) > .0001 || obs[0].Period != "2026-01" {
		t.Fatalf("wrong aggregate or denominator %+v", obs)
	}
	for _, series := range []string{wanted + wanted, strings.ReplaceAll(wanted, `OBS_VALUE="105"`, `OBS_VALUE="chênh"`), strings.ReplaceAll(wanted, `UNIT_MULT="0"`, `UNIT_MULT="3"`), wrong} {
		if _, err := ParseSDMX([]byte(`<Root><DataSet>`+series+`</DataSet></Root>`), "VN_CPI_YOY", "PCPI_IX", "https://nsdp.nso.gov.vn/cpi.xml"); err == nil {
			t.Fatal("ambiguous or invalid data accepted")
		}
	}
}
func reportJSON(title, body string) []byte {
	p := post{Date: "2026-10-03T08:00:06", Link: "https://www.nso.gov.vn/report/"}
	p.Title.Rendered = title
	p.Content.Rendered = body
	data, _ := json.Marshal([]post{p})
	return data
}
func TestMonthlyReportsNeverUseCumulativeValues(t *testing.T) {
	body := `<p>Chỉ số giá tiêu dùng (CPI) tháng Chín tăng 0,62% so với tháng trước; tăng 5,08% so với cùng kỳ năm trước. CPI bình quân quý III/2026 tăng 4,8%. Bình quân chín tháng CPI tăng 4,52% so với cùng kỳ năm trước.</p><p>Tổng mức bán lẻ hàng hóa và doanh thu dịch vụ tiêu dùng theo giá hiện hành tháng Chín năm 2026 ước đạt 690,6 nghìn tỷ đồng, tăng 1,7% so với tháng trước và tăng 14,1% so với cùng kỳ năm trước. Tính chung chín tháng tăng 13,4% so với cùng kỳ năm trước.</p>`
	obs, _, err := ParseReports(reportJSON("Báo cáo tình hình kinh tế – xã hội quý III và 9 tháng năm 2026", body), "economic")
	if err != nil {
		t.Fatal(err)
	}
	if len(obs) != 2 || obs[0].Value != 5.08 || obs[1].Value != 14.1 || obs[0].Period != "2026-09" || obs[0].PublishedAt == nil {
		t.Fatalf("wrong monthly extraction %+v", obs)
	}
	bad := `<p>Bình quân chín tháng năm 2026, CPI tăng 4,52% so với cùng kỳ năm trước; lạm phát cơ bản tăng 4,26%.</p>`
	if _, _, err := ParseReports(reportJSON("Báo cáo tình hình kinh tế – xã hội quý III và 9 tháng năm 2026", bad), "economic"); err == nil {
		t.Fatal("cumulative CPI must not be extracted")
	}
	core := `<p>Lạm phát cơ bản [2] tháng 9/2026 tăng 0,11% so với tháng trước và tăng 4,45% so với cùng kỳ năm trước. Bình quân 9 tháng năm 2026, lạm phát cơ bản tăng 4,26% so với cùng kỳ năm trước.</p>`
	obs, _, err = ParseReports(reportJSON("Thông cáo báo chí về tình hình giá tháng Chín, quý III và 9 tháng năm 2026", core), "prices")
	if err != nil || len(obs) != 1 || obs[0].Value != 4.45 {
		t.Fatalf("wrong core inflation %+v %v", obs, err)
	}
	conflicting := strings.ReplaceAll(core, "4,45%", "5,45%")
	if _, _, err := ParseReports(reportJSON("Thông cáo báo chí về tình hình giá tháng Chín, quý III và 9 tháng năm 2026", core+conflicting), "prices"); err == nil {
		t.Fatal("conflicting report values accepted")
	}
}
func TestNSOLiveSmoke(t *testing.T) {
	if os.Getenv("NSO_LIVE_SMOKE") != "1" {
		t.Skip("set NSO_LIVE_SMOKE=1 for official-source smoke test")
	}
	sources := []Provider{&SDMX{Code: "VN_CPI_YOY", File: "CPIVNM", Series: "PCPI_IX"}, &SDMX{Code: "VN_IIP_YOY", File: "IIPVNM", Series: "AIP_ISIC4_IX"}, &SDMX{Code: "VN_EXPORT_YOY", File: "METVNM", Series: "TXG_FOB_USD"}, &Reports{Kind: "economic"}, &Reports{Kind: "prices"}}
	for _, source := range sources {
		t.Run(source.ID(), func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			batch, err := source.Fetch(ctx)
			if err != nil {
				t.Fatal(err)
			}
			latest := map[string]macro.Observation{}
			for _, o := range batch.Observations {
				if old, ok := latest[o.IndicatorCode]; !ok || o.Period > old.Period {
					latest[o.IndicatorCode] = o
				}
			}
			for code, o := range latest {
				t.Logf("%s: %s = %.4f (%s), observations=%d", code, o.Period, o.Value, o.SeriesID, len(batch.Observations))
			}
		})
	}
}
