package providers

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/trungvdn/investment-ai/apps/macro-api/internal/macro"
	htmlparse "golang.org/x/net/html"
)

// PMI consumes public releases, not S&P's licensed historical database.
// Explicit URLs override discovery from the public release index.
type PMI struct {
	HTTP
	ReleaseURLs        []string
	IndexURL           string
	FeedURL, FeedToken string
	ManifestFile       string
	LocalReleases      []PMILocalRelease
}

func (p *PMI) ID() string      { return "spglobal:pmi:vn" }
func (p *PMI) Codes() []string { return []string{"VN_PMI"} }

const PMINotConfigured = "PMI_RELEASE_URLS is not configured"

func (p *PMI) Fetch(ctx context.Context) (macro.Batch, error) {
	if p.FeedToken != "" && p.FeedURL == "" {
		return macro.Batch{}, fmt.Errorf("PMI_FEED_TOKEN requires PMI_FEED_URL; DXY settings are independent")
	}
	if p.FeedURL != "" {
		return p.fetchFeed(ctx)
	}
	if p.ManifestFile != "" {
		return p.fetchManifest(ctx)
	}
	if len(p.LocalReleases) > 0 {
		return p.fetchLocal(ctx)
	}
	b := macro.Batch{}
	urls := p.ReleaseURLs
	offset := 0
	if len(urls) == 0 {
		index := p.IndexURL
		if index == "" {
			index = "https://www.pmi.spglobal.com/Public/Release/PressReleases?language=en"
		}
		u, e := url.Parse(index)
		if e != nil || u.Scheme != "https" || u.User != nil || (u.Host != "www.pmi.spglobal.com" && u.Host != "pmi.spglobal.com") || u.Path != "/Public/Release/PressReleases" {
			return b, fmt.Errorf("invalid PMI official release index URL")
		}
		doc, e := p.Get(ctx, index, index, "S&P Global")
		if e != nil {
			return b, pmiFetchError("release index", e)
		}
		b.Document = doc
		urls, e = DiscoverPMI(doc.Body, index)
		if e != nil {
			return b, e
		}
		offset = 1
	}

	if len(urls) > 36 {
		return b, fmt.Errorf("PMI supports at most 36 release URLs per run")
	}
	seen := map[string]int{}
	for i, u := range urls {
		parsed, err := url.Parse(u)
		if err != nil || parsed.Scheme != "https" || parsed.User != nil || parsed.Port() != "" || (parsed.Hostname() != "www.pmi.spglobal.com" && parsed.Hostname() != "pmi.spglobal.com") || !regexp.MustCompile(`^/Public/Home/PressRelease/[a-fA-F0-9]{32}$`).MatchString(parsed.Path) || parsed.RawQuery != "" || parsed.Fragment != "" {
			return b, fmt.Errorf("PMI requires official S&P Global public release URLs")
		}
		d, err := p.Get(ctx, u, u, "S&P Global")
		if err != nil {
			return b, pmiFetchError("release document", err)
		}
		if i == 0 && offset == 0 {
			b.Document = d
		} else {
			b.Documents = append(b.Documents, d)
		}
		text, err := pmiDocumentText(ctx, d.Body)
		if err != nil {
			return b, err
		}
		o, err := ParsePMIText(text, u, time.Now())
		if err != nil {
			return b, err
		}

		o.RawIndex = i + offset
		if j, exists := seen[o.Period]; exists {
			if offset == 0 {
				return b, fmt.Errorf("duplicate PMI release month; select one vintage per period")
			}
			previous := b.Observations[j]
			if o.PublishedAt.Equal(*previous.PublishedAt) && o.Value != previous.Value {
				return b, fmt.Errorf("conflicting PMI releases for the same embargo time; select an explicit vintage")
			}
			if o.PublishedAt.After(*previous.PublishedAt) {
				b.Observations[j] = o
			}
			continue
		}
		seen[o.Period] = len(b.Observations)
		b.Observations = append(b.Observations, o)
	}
	return b, nil
}
func pmiDocumentText(ctx context.Context, body []byte) (string, error) {
	if !bytes.HasPrefix(body, []byte("%PDF-")) {
		root, err := htmlparse.Parse(bytes.NewReader(body))
		if err != nil {
			return "", fmt.Errorf("invalid PMI document")
		}
		return nodeText(root), nil
	}
	f, err := os.CreateTemp("", "investment-pmi-*.pdf")
	if err != nil {
		return "", fmt.Errorf("cannot prepare PMI PDF")
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(body); err != nil {
		f.Close()
		return "", fmt.Errorf("cannot prepare PMI PDF")
	}
	f.Close()
	run, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	// Only page one contains the headline. Do not parse chart labels or prior months
	// from later pages. No shell interpolation; the document is never executed.
	cmd := exec.CommandContext(run, "pdftotext", "-f", "1", "-l", "1", "-raw", f.Name(), "-")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("PMI PDF extraction failed; install poppler-utils (pdftotext)")
	}
	if len(out) > 1024*1024 {
		return "", fmt.Errorf("PMI extracted text too large")
	}
	return string(out), nil
}

const englishMonth = `January|February|March|April|May|June|July|August|September|October|November|December`

func ParsePMIText(text, sourceURL string, now time.Time) (macro.Observation, error) {
	empty := macro.Observation{}
	text = strings.Join(strings.Fields(strings.NewReplacer("®", "", "™", "", "’", "'", "\u00a0", " ").Replace(text)), " ")
	pubMatch := regexp.MustCompile(`(?i)embargoed until (\d{1,2} (?:` + englishMonth + `) \d{4}) \((\d{4}) ICT`).FindStringSubmatch(text)
	if len(pubMatch) != 3 {
		return empty, fmt.Errorf("PMI release lacks unambiguous embargo date/time")
	}
	pub, err := time.ParseInLocation("2 January 2006 1504", pubMatch[1]+" "+pubMatch[2], time.FixedZone("ICT", 7*3600))
	if err != nil {
		return empty, fmt.Errorf("invalid PMI publication date")
	}
	pattern := `(?i)(?:The )?S&P Global Vietnam Manufacturing (?:Purchasing Managers' Index\s*\(PMI\)|PMI|Purchasing Managers' Index)[^.!?]{0,100}?(?:posted|registered|recorded|stood at|rose to|fell to|dropped to|increased to|decreased to|climbed to|edged (?:up|down) to)\s+(\d{1,3}(?:\.\d+)?)\s*(?:points?\s*)?in\s+(` + englishMonth + `)`
	matches := regexp.MustCompile(pattern).FindAllStringSubmatch(text, -1)
	if len(matches) != 1 {
		return empty, fmt.Errorf("PMI headline value/month is missing or ambiguous")
	}
	value, err := strconv.ParseFloat(matches[0][1], 64)
	if err != nil || value < 0 || value > 100 {
		return empty, fmt.Errorf("invalid PMI headline")
	}
	// Publication normally occurs in the next month; December belongs to the
	// previous year. Require an explicit matching report month/year in the release.
	month, err := time.Parse("January", strings.Title(strings.ToLower(matches[0][2])))
	if err != nil {
		return empty, err
	}
	year := pub.Year()
	if month.Month() > pub.Month() {
		year--
	}
	reportDate := time.Date(year, month.Month(), 1, 0, 0, 0, 0, time.UTC)
	if !strings.Contains(strings.ToLower(text), strings.ToLower(reportDate.Format("January 2006"))) {
		return empty, fmt.Errorf("PMI report month/year not confirmed in release")
	}
	if pub.Before(reportDate) || pub.After(reportDate.AddDate(0, 2, 0)) {
		return empty, fmt.Errorf("PMI report and publication dates disagree")
	}
	end, _ := macro.MonthEnd(reportDate.Format("2006-01"))
	o := macro.Observation{IndicatorCode: "VN_PMI", Period: reportDate.Format("2006-01"), PeriodEnd: end, Value: value, Unit: "index", Source: "S&P Global", SourceURL: sourceURL, SeriesID: "SPGLOBAL_VN_MANUFACTURING_PMI", Method: "SOURCE_HEADLINE_SEASONALLY_ADJUSTED; 50_NO_CHANGE; PUBLIC_RELEASE", Priority: 100, PublishedAt: &pub}
	if err = macro.Validate(o, now); err != nil {
		return empty, err
	}
	return o, nil
}

// Discover only explicitly titled Vietnam manufacturing releases; never guesses IDs.
func DiscoverPMI(body []byte, index string) ([]string, error) {
	root, err := htmlparse.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("invalid PMI release index")
	}
	base, _ := url.Parse(index)
	out := []string{}
	seen := map[string]bool{}
	for _, a := range nodes(root, "a") {
		title := nodeText(a)
		href := ""
		for _, v := range a.Attr {
			if v.Key == "href" {
				href = v.Val
			}
			if v.Key == "title" {
				title += " " + v.Val
			}
		}
		text := fold(title)
		if !(strings.Contains(text, "vietnam") || strings.Contains(text, "viet nam")) || !strings.Contains(text, "manufacturing") {
			continue
		}
		link, e := url.Parse(href)
		if e != nil {
			continue
		}
		u := base.ResolveReference(link)
		if u.Scheme != "https" || u.User != nil || (u.Host != "www.pmi.spglobal.com" && u.Host != "pmi.spglobal.com") || !regexp.MustCompile(`^/Public/Home/PressRelease/[a-fA-F0-9]{32}$`).MatchString(u.Path) {
			continue
		}
		u.RawQuery = ""
		u.Fragment = ""
		value := u.String()
		if seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
		if len(out) == 6 {
			break
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no explicitly titled Vietnam manufacturing releases in official PMI index; configure PMI_RELEASE_URLS if index layout changes")
	}
	return out, nil
}
