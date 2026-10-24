package providers

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestRetryClassificationAndRetryAfter(t *testing.T) {
	for _, status := range []int{403, 429, 503} {
		calls := 0
		waits := []time.Duration{}
		h := HTTP{Client: &http.Client{Transport: phase5Transport(func(r *http.Request) (*http.Response, error) {
			calls++
			header := http.Header{}
			if status == 429 {
				header.Set("Retry-After", "2")
			}
			return &http.Response{StatusCode: status, Header: header, Body: io.NopCloser(strings.NewReader("blocked"))}, nil
		})}, Wait: func(_ context.Context, d time.Duration) error { waits = append(waits, d); return nil }}
		_, err := h.Get(context.Background(), "https://source.test/?api_key=secret", "https://public.test/", "SBV")
		if err == nil || strings.Contains(err.Error(), "secret") {
			t.Fatal("unsafe/missing error")
		}
		if status == 403 && (calls != 1 || len(waits) != 0 || ErrorKind(err) != "ACCESS_DENIED") {
			t.Fatal("403 was retried or misclassified")
		}
		if status == 429 && (calls != 3 || len(waits) != 2 || waits[0] != 2*time.Second || ErrorKind(err) != "RATE_LIMIT") {
			t.Fatal("Retry-After not respected")
		}
		if status == 503 && (calls != 3 || len(waits) != 2 || waits[0] < time.Second || waits[1] < 2*time.Second) {
			t.Fatal("exponential retry missing")
		}
	}
	calls := 0
	h := HTTP{Client: &http.Client{Transport: phase5Transport(func(r *http.Request) (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: 429, Header: http.Header{"Retry-After": []string{"7200"}}, Body: io.NopCloser(strings.NewReader(""))}, nil
	})}}
	_, err := h.Get(context.Background(), "https://source.test/", "https://public.test/", "ICE")
	if calls != 1 || RetryAt(err) == nil || RetryAt(err).Before(time.Now().Add(119*time.Minute)) {
		t.Fatal("long Retry-After was shortened or slept inline")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if waitRetry(ctx, time.Hour) == nil {
		t.Fatal("retry wait ignored cancellation")
	}
}
func TestDXYDoesNotSubstituteBroadUSDOrFutures(t *testing.T) {
	text := "date,value,symbol\n2026-10-23,101.2,DXY\n"
	p := DXY{URL: "https://export.example.test/data", Token: "private-token", HTTP: HTTP{Client: &http.Client{Transport: phase5Transport(func(r *http.Request) (*http.Response, error) {
		if r.Header.Get("Authorization") != "Bearer private-token" {
			t.Fatal("feed credential missing")
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(text))}, nil
	})}}}
	b, err := p.Fetch(context.Background())
	if err != nil || len(b.Observations) != 1 || b.Observations[0].IndicatorCode != "US_DXY" || b.Document.URL != DXYSourceURL {
		t.Fatalf("DXY export failed %+v %v", b, err)
	}
	for _, bad := range []string{strings.ReplaceAll(text, "DXY", "DX"), strings.ReplaceAll(text, "DXY", "DTWEXBGS"), text + "2026-10-23,101.3,DXY\n", strings.ReplaceAll(text, "101.2", "NaN"), strings.ReplaceAll(text, "2026-10-23", "2099-01-01")} {
		text = bad
		if _, err = p.Fetch(context.Background()); err == nil {
			t.Fatal("invalid/substitute DXY accepted")
		}
	}
	p.URL = ""
	if _, err = p.Fetch(context.Background()); ErrorKind(err) != "NOT_CONFIGURED" {
		t.Fatal("unconfigured DXY not explicit")
	}
}
func TestPMIDiscoveryAndRawDocumentIndex(t *testing.T) {
	index := "https://www.pmi.spglobal.com/Public/Release/PressReleases?language=en"
	html := `<a href="/Public/Home/PressRelease/dba8ca68ce13403591b68cb3e07175d9">S&amp;P Global Vietnam Manufacturing PMI</a><a href="https://evil.test/Public/Home/PressRelease/dba8ca68ce13403591b68cb3e07175d9">Vietnam Manufacturing</a><a href="/Public/Home/PressRelease/11111111111111111111111111111111">Thailand Manufacturing PMI</a>`
	urls, err := DiscoverPMI([]byte(html), index)
	if err != nil || len(urls) != 1 || urls[0] != pmiRelease {
		t.Fatalf("unsafe discovery %+v %v", urls, err)
	}
	p := PMI{HTTP: HTTP{Client: &http.Client{Transport: phase5Transport(func(r *http.Request) (*http.Response, error) {
		text := pmiText
		if strings.Contains(r.URL.Path, "PressReleases") {
			text = html
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(text))}, nil
	})}}}
	b, err := p.Fetch(context.Background())
	if err != nil || len(b.Documents) != 1 || len(b.Observations) != 1 || b.Observations[0].RawIndex != 1 {
		t.Fatalf("index provenance wrong %+v %v", b, err)
	}
}
