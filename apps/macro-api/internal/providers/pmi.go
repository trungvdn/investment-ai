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
// URLs are explicit so a new release or revision cannot silently change periods.
type PMI struct {
	HTTP
	ReleaseURLs []string
}

func (p *PMI) ID() string      { return "spglobal:pmi:vn" }
func (p *PMI) Codes() []string { return []string{"VN_PMI"} }

const PMINotConfigured = "PMI_RELEASE_URLS is not configured"

func (p *PMI) Fetch(ctx context.Context) (macro.Batch, error) {
	b := macro.Batch{}
	if len(p.ReleaseURLs) == 0 {
		return b, fmt.Errorf(PMINotConfigured)
	}
	if len(p.ReleaseURLs) > 36 {
		return b, fmt.Errorf("PMI supports at most 36 release URLs per run")
	}
	seen := map[string]bool{}
	for i, u := range p.ReleaseURLs {
		parsed, err := url.Parse(u)
		if err != nil || parsed.Scheme != "https" || parsed.User != nil || parsed.Port() != "" || (parsed.Hostname() != "www.pmi.spglobal.com" && parsed.Hostname() != "pmi.spglobal.com") || !regexp.MustCompile(`^/Public/Home/PressRelease/[a-fA-F0-9]{32}$`).MatchString(parsed.Path) || parsed.RawQuery != "" || parsed.Fragment != "" {
			return b, fmt.Errorf("PMI requires official S&P Global public release URLs")
		}
		d, err := p.Get(ctx, u, u, "S&P Global")
		if err != nil {
			return b, err
		}
		if i == 0 {
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
		if seen[o.Period] {
			return b, fmt.Errorf("duplicate PMI release month; select one vintage per period")
		}
		seen[o.Period] = true
		o.RawIndex = i
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
	pattern := `(?i)(?:The )?S&P Global Vietnam Manufacturing (?:Purchasing Managers' Index\s*\(PMI\)|PMI|Purchasing Managers' Index)[^.!?]{0,100}?(?:posted|registered|recorded|stood at|rose to|fell to|dropped to|increased to|decreased to|climbed to|edged (?:up|down) to)\s+(\d{1,2}(?:\.\d+)?)\s*(?:points?\s*)?in\s+(` + englishMonth + `)`
	matches := regexp.MustCompile(pattern).FindAllStringSubmatch(text, -1)
	if len(matches) != 1 {
		return empty, fmt.Errorf("PMI headline value/month is missing or ambiguous")
	}
	value, err := strconv.ParseFloat(matches[0][1], 64)
	if err != nil || value <= 0 || value > 100 {
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
