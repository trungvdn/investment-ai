package providers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func pmiExportFixture() []byte {
	return []byte(`{"seriesId":"SPGLOBAL_VN_MANUFACTURING_PMI","country":"VN","frequency":"MONTHLY","observations":[{"period":"2026-09","value":51.9,"publishedAt":"2026-10-01T00:30:00Z","sourceUrl":"` + pmiRelease + `"}]}`)
}
func TestPMI403IdentifiesStageAndKeepsErrorClassification(t *testing.T) {
	for _, explicit := range []bool{false, true} {
		calls := 0
		p := PMI{HTTP: HTTP{Client: &http.Client{Transport: phase5Transport(func(r *http.Request) (*http.Response, error) {
			calls++
			if r.Header.Get("Authorization") != "" {
				t.Fatal("credential sent to public site")
			}
			return &http.Response{StatusCode: 403, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("denied"))}, nil
		})}}}
		stage := "release index"
		if explicit {
			p.ReleaseURLs = []string{pmiRelease}
			stage = "release document"
		}
		_, err := p.Fetch(context.Background())
		var source *SourceError
		if err == nil || !errors.As(err, &source) || ErrorKind(err) != "ACCESS_DENIED" || calls != 1 || !strings.Contains(err.Error(), stage) || !strings.Contains(err.Error(), "PMI_FEED_URL") {
			t.Fatalf("missing safe PMI diagnostic %v calls=%d", err, calls)
		}
	}
}
func TestPMIAuthorizedFeedSkipsBlockedPublicSite(t *testing.T) {
	calls := 0
	p := PMI{FeedURL: "https://export.example.test/pmi?api_key=private", FeedToken: "pmi-secret", HTTP: HTTP{Client: &http.Client{Transport: phase5Transport(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.URL.Host != "export.example.test" {
			t.Fatal("blocked S&P host requested despite feed configuration")
		}
		if r.Header.Get("Authorization") != "Bearer pmi-secret" {
			t.Fatal("PMI feed credential missing")
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(string(pmiExportFixture())))}, nil
	})}}}
	b, err := p.Fetch(context.Background())
	if err != nil || calls != 1 || len(b.Observations) != 1 || b.Observations[0].Value != 51.9 || b.Observations[0].SourceURL != pmiRelease || b.Document.URL != pmiPublicURL {
		t.Fatalf("feed ingestion failed %+v %v", b, err)
	}
	if strings.Contains(b.Document.URL, "private") {
		t.Fatal("private endpoint exposed")
	}
	p.HTTP.Client = &http.Client{Transport: phase5Transport(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 403, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("no"))}, nil
	})}
	_, err = p.Fetch(context.Background())
	if err == nil || strings.Contains(err.Error(), "private") || strings.Contains(err.Error(), "pmi-secret") || ErrorKind(err) != "ACCESS_DENIED" || !strings.Contains(err.Error(), "authorized feed") {
		t.Fatal("unsafe or misclassified export error")
	}
	p.FeedURL = ""
	if _, err = p.Fetch(context.Background()); err == nil || !strings.Contains(err.Error(), "requires PMI_FEED_URL") {
		t.Fatal("orphan token not rejected")
	}
}
func TestPMIFeedRejectsWrongSeriesAndBadData(t *testing.T) {
	good := string(pmiExportFixture())
	bads := []string{strings.Replace(good, `"VN"`, `"US"`, 1), strings.Replace(good, `"MONTHLY"`, `"DAILY"`, 1), strings.Replace(good, "SPGLOBAL_VN_MANUFACTURING_PMI", "SERVICES_PMI", 1), strings.Replace(good, `"value":51.9,`, "", 1), strings.Replace(good, "51.9", "101.1", 1), strings.Replace(good, "2026-09", "2099-09", 1), strings.Replace(good, "2026-10-01T00:30:00Z", "2026-09-01T00:30:00Z", 1), strings.Replace(good, "www.pmi.spglobal.com", "evil.test", 1), good + good}
	var duplicate map[string]any
	json.Unmarshal([]byte(good), &duplicate)
	rows := duplicate["observations"].([]any)
	duplicate["observations"] = append(rows, rows[0])
	body, _ := json.Marshal(duplicate)
	bads = append(bads, string(body))
	for _, text := range bads {
		if _, err := ParsePMIFeed([]byte(text), phase5Now); err == nil {
			t.Fatal("invalid PMI export accepted")
		}
	}
	if out, err := ParsePMIFeed([]byte(strings.Replace(good, "51.9", "0", 1)), phase5Now); err != nil || len(out) != 1 || out[0].Value != 0 {
		t.Fatal("actual zero cannot be distinguished from a missing value")
	}
}
func TestPMILocalPDFNeedsNoNetworkAndRetainsExactProvenance(t *testing.T) {
	if _, err := exec.LookPath("pdftotext"); err != nil {
		t.Skip("install poppler-utils")
	}
	path := filepath.Join(t.TempDir(), "official.pdf")
	if err := os.WriteFile(path, minimalPMIPDF(), 0600); err != nil {
		t.Fatal(err)
	}
	p := PMI{LocalReleases: []PMILocalRelease{{Path: path, SourceURL: pmiRelease}}, HTTP: HTTP{Client: &http.Client{Transport: phase5Transport(func(*http.Request) (*http.Response, error) {
		t.Fatal("local mode must not fetch the blocked site")
		return nil, nil
	})}}}
	b, err := p.Fetch(context.Background())
	if err != nil || len(b.Observations) != 1 || b.Observations[0].RawIndex != 0 || b.Document.URL != pmiRelease || !strings.HasPrefix(string(b.Document.Body), "%PDF-") {
		t.Fatalf("local PDF ingestion failed %+v %v", b, err)
	}
	p.LocalReleases = append(p.LocalReleases, p.LocalReleases[0])
	if _, err = p.Fetch(context.Background()); err == nil {
		t.Fatal("duplicate local vintage accepted")
	}
	p.LocalReleases = PMILocalFromEnv("not-json")
	if _, err = p.Fetch(context.Background()); err == nil {
		t.Fatal("invalid local configuration silently fell through")
	}
	p.LocalReleases = []PMILocalRelease{{Path: t.TempDir(), SourceURL: pmiRelease}}
	if _, err = p.Fetch(context.Background()); err == nil {
		t.Fatal("directory accepted as release")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	p.LocalReleases = []PMILocalRelease{{Path: path, SourceURL: pmiRelease}}
	if _, err = p.Fetch(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal("local import ignored cancellation")
	}
}

func TestPythonCollectorManifestImportsAndPropagatesHealth(t *testing.T) {
	if _, err := exec.LookPath("pdftotext"); err != nil {
		t.Skip("install poppler-utils")
	}
	dir := t.TempDir()
	pdf := filepath.Join(dir, "release.pdf")
	os.WriteFile(pdf, minimalPMIPDF(), 0600)
	manifest := filepath.Join(dir, "manifest.json")
	now := time.Now().UTC()
	data := map[string]any{"version": 1, "collector": "vn_pmi", "targetPeriod": "2026-09", "checkedAt": now.Add(-time.Minute).Format(time.RFC3339), "nextDue": now.Add(time.Hour).Format(time.RFC3339), "status": "READY", "releases": []PMILocalRelease{{Path: pdf, SourceURL: pmiRelease}}}
	write := func() {
		b, _ := json.Marshal(data)
		if e := os.WriteFile(manifest, b, 0600); e != nil {
			t.Fatal(e)
		}
	}
	write()
	p := PMI{ManifestFile: manifest, HTTP: HTTP{Client: &http.Client{Transport: phase5Transport(func(*http.Request) (*http.Response, error) {
		t.Fatal("manifest must not request public source")
		return nil, nil
	})}}}
	b, err := p.Fetch(context.Background())
	if err != nil || len(b.Observations) != 1 || b.Observations[0].RawIndex != 1 || len(b.Documents) != 1 {
		t.Fatalf("Python bridge failed %+v %v", b, err)
	}
	data["status"] = "ERROR"
	data["errorKind"] = "ACCESS_DENIED"
	write()
	b, err = p.Fetch(context.Background())
	if ErrorKind(err) != "ACCESS_DENIED" || len(b.Observations) != 0 || len(b.Document.Body) == 0 {
		t.Fatal("cached PDF hid scraper access denial")
	}
	data["status"] = "PENDING"
	write()
	b, err = p.Fetch(context.Background())
	if err != nil || !strings.HasPrefix(b.Unavailable["VN_PMI"], "AWAITING_RELEASE:") {
		t.Fatal("publication delay incorrectly treated as a value")
	}
	data["status"] = "READY"
	data["nextDue"] = now.Add(-25 * time.Hour).Format(time.RFC3339)
	data["checkedAt"] = now.Add(-30 * time.Hour).Format(time.RFC3339)
	write()
	if _, err = p.Fetch(context.Background()); err == nil {
		t.Fatal("stopped collector appeared healthy")
	}
}
