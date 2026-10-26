package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/trungvdn/investment-ai/apps/macro-api/internal/macro"
)

const pmiPublicURL = "https://www.pmi.spglobal.com/"

type PMILocalRelease struct {
	Path      string `json:"path"`
	SourceURL string `json:"sourceUrl"`
}

// An invalid environment value becomes an explicit configuration error at Fetch,
// never a silent fallthrough to the blocked public site.
func PMILocalFromEnv(raw string) []PMILocalRelease {
	if raw == "" {
		return nil
	}
	var out []PMILocalRelease
	if json.Unmarshal([]byte(raw), &out) != nil || len(out) == 0 {
		return []PMILocalRelease{{}}
	}
	return out
}
func officialPMIRelease(u string) bool {
	parsed, err := url.Parse(u)
	return err == nil && parsed.Scheme == "https" && parsed.User == nil && parsed.Port() == "" && (parsed.Hostname() == "www.pmi.spglobal.com" || parsed.Hostname() == "pmi.spglobal.com") && regexp.MustCompile(`^/Public/Home/PressRelease/[a-fA-F0-9]{32}$`).MatchString(parsed.Path) && parsed.RawQuery == "" && parsed.Fragment == ""
}
func pmiFetchError(stage string, err error) error {
	var source *SourceError
	if errors.As(err, &source) && source.Status == http.StatusForbidden {
		return fmt.Errorf("PMI %s: %w; S&P denied this request before parsing. Use a permitted PMI_FEED_URL/PMI_FEED_TOKEN export or PMI_RELEASE_FILES with official PDFs; DXY settings do not apply", stage, err)
	}
	return fmt.Errorf("PMI %s: %w", stage, err)
}
func appendPMIDocument(b *macro.Batch, d macro.RawDocument) int {
	if len(b.Document.Body) == 0 {
		b.Document = d
		return 0
	}
	b.Documents = append(b.Documents, d)
	return len(b.Documents)
}
func (p *PMI) fetchLocal(ctx context.Context) (macro.Batch, error) {
	b := macro.Batch{}
	if len(p.LocalReleases) > 36 {
		return b, fmt.Errorf("PMI_RELEASE_FILES supports at most 36 releases")
	}
	seen := map[string]bool{}
	for _, file := range p.LocalReleases {
		if !filepath.IsAbs(file.Path) || !officialPMIRelease(file.SourceURL) {
			return b, fmt.Errorf("PMI_RELEASE_FILES requires absolute file paths and official S&P release URLs")
		}
		if err := ctx.Err(); err != nil {
			return b, err
		}
		info, err := os.Stat(file.Path)
		if err != nil || !info.Mode().IsRegular() || info.Size() > 16*1024*1024 {
			return b, fmt.Errorf("PMI local release must be a readable regular file no larger than 16 MiB")
		}
		f, err := os.Open(file.Path)
		if err != nil {
			return b, fmt.Errorf("PMI local release could not be opened; verify mounted file permissions")
		}
		info, err = f.Stat()
		if err != nil || !info.Mode().IsRegular() {
			f.Close()
			return b, fmt.Errorf("PMI local release must be a regular file")
		}
		body, err := io.ReadAll(io.LimitReader(f, 16*1024*1024+1))
		f.Close()
		if err != nil || len(body) > 16*1024*1024 || !bytes.HasPrefix(body, []byte("%PDF-")) {
			return b, fmt.Errorf("PMI local release must be a PDF no larger than 16 MiB")
		}
		index := appendPMIDocument(&b, macro.RawDocument{Source: "S&P Global", URL: file.SourceURL, ContentType: "application/pdf", Body: body, FetchedAt: time.Now().UTC()})
		text, err := pmiDocumentText(ctx, body)
		if err != nil {
			return b, err
		}
		o, err := ParsePMIText(text, file.SourceURL, time.Now())
		if err != nil {
			return b, err
		}
		if seen[o.Period] {
			return b, fmt.Errorf("duplicate local PMI release month; select one vintage per period")
		}
		seen[o.Period] = true
		o.RawIndex = index
		o.Method += "; TRANSPORT_LOCAL_OFFICIAL_PDF"
		b.Observations = append(b.Observations, o)
	}
	return b, nil
}
func (p *PMI) fetchFeed(ctx context.Context) (macro.Batch, error) {
	b := macro.Batch{}
	u, err := url.Parse(p.FeedURL)
	if err != nil || u.Scheme != "https" || u.User != nil || u.Host == "" || u.Fragment != "" {
		return b, fmt.Errorf("PMI_FEED_URL must use HTTPS without URL credentials or fragment")
	}
	h := p.HTTP
	h.Headers = h.Headers.Clone()
	if h.Headers == nil {
		h.Headers = make(http.Header)
	}
	if p.FeedToken != "" {
		h.Headers.Set("Authorization", "Bearer "+p.FeedToken)
	}
	d, err := h.Get(ctx, p.FeedURL, pmiPublicURL, "PMI authorized feed")
	b.Document = d
	if err != nil {
		return b, fmt.Errorf("PMI authorized feed: %w", err)
	}
	b.Observations, err = ParsePMIFeed(d.Body, time.Now())
	return b, err
}

// This is an explicit export contract, not a generic scrape or a replacement index.
func ParsePMIFeed(body []byte, now time.Time) ([]macro.Observation, error) {
	var payload struct {
		SeriesID     string `json:"seriesId"`
		Country      string `json:"country"`
		Frequency    string `json:"frequency"`
		Observations []struct {
			Period      string   `json:"period"`
			Value       *float64 `json:"value"`
			PublishedAt string   `json:"publishedAt"`
			SourceURL   string   `json:"sourceUrl"`
		} `json:"observations"`
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return nil, fmt.Errorf("invalid PMI export JSON; check documented schema")
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return nil, fmt.Errorf("PMI export must contain exactly one JSON object")
	}
	if payload.SeriesID != "SPGLOBAL_VN_MANUFACTURING_PMI" || payload.Country != "VN" || payload.Frequency != "MONTHLY" || len(payload.Observations) == 0 || len(payload.Observations) > 1200 {
		return nil, fmt.Errorf("PMI export must be monthly Vietnam manufacturing headline data")
	}
	out := []macro.Observation{}
	seen := map[string]bool{}
	for _, row := range payload.Observations {
		end, err := macro.MonthEnd(row.Period)
		if err != nil || seen[row.Period] || row.Value == nil || !officialPMIRelease(row.SourceURL) {
			return nil, fmt.Errorf("PMI export has invalid/duplicate period, missing value or non-official release provenance")
		}
		seen[row.Period] = true
		published, err := time.Parse(time.RFC3339, row.PublishedAt)
		if err != nil || published.Before(end) || published.After(end.AddDate(0, 1, 2)) {
			return nil, fmt.Errorf("PMI export publication time is missing or inconsistent with report period")
		}
		o := macro.Observation{IndicatorCode: "VN_PMI", Period: row.Period, PeriodEnd: end, Value: *row.Value, Unit: "index", Source: "S&P Global", SourceURL: row.SourceURL, SeriesID: payload.SeriesID, Method: "AUTHORIZED_EXPORT_HEADLINE_SEASONALLY_ADJUSTED; 50_NO_CHANGE; TRANSPORT_AUTHORIZED_FEED", Priority: 100, PublishedAt: &published}
		if err = macro.Validate(o, now); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, nil
}
