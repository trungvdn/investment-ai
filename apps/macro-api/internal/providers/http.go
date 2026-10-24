package providers

import (
	"context"
	"fmt"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/macro"
	"io"
	"net/http"
	"time"
)

type Provider interface {
	ID() string
	Codes() []string
	Fetch(context.Context) (macro.Batch, error)
}
type HTTP struct{ Client *http.Client }

func (h HTTP) Get(ctx context.Context, requestURL, publicURL, source string) (macro.RawDocument, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return macro.RawDocument{}, fmt.Errorf("invalid source URL")
	}
	req.Header.Set("User-Agent", "InvestmentAI/1.0 (+https://github.com/trungvdn/investment-ai)")
	client := h.Client
	if client == nil {
		client = &http.Client{Timeout: 35 * time.Second}
	}
	var resp *http.Response
	for attempt := 0; attempt < 3; attempt++ {
		resp, err = client.Do(req)
		retryable := err != nil || resp.StatusCode == 429 || resp.StatusCode >= 500
		if !retryable || attempt == 2 {
			break
		}
		if resp != nil {
			_ = resp.Body.Close()
		}
		timer := time.NewTimer(time.Duration(attempt+1) * time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return macro.RawDocument{}, ctx.Err()
		case <-timer.C:
		}
	}
	if err != nil {
		return macro.RawDocument{}, fmt.Errorf("source %s request failed", source)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return macro.RawDocument{}, fmt.Errorf("source %s returned HTTP %d", source, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 16*1024*1024+1))
	if err != nil {
		return macro.RawDocument{}, fmt.Errorf("read source response failed")
	}
	if len(body) > 16*1024*1024 {
		return macro.RawDocument{}, fmt.Errorf("source payload exceeds 16 MiB")
	}
	return macro.RawDocument{Source: source, URL: publicURL, ContentType: resp.Header.Get("Content-Type"), Body: body, FetchedAt: time.Now().UTC()}, nil
}
