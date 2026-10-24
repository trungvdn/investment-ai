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
	for attempt := 0; attempt < 3; attempt++ {
		resp, requestErr := client.Do(req)
		var body []byte
		var readErr error
		if requestErr == nil {
			if resp.StatusCode == http.StatusOK {
				body, readErr = io.ReadAll(io.LimitReader(resp.Body, 16*1024*1024+1))
			}
			_ = resp.Body.Close()
			if len(body) > 16*1024*1024 {
				return macro.RawDocument{}, fmt.Errorf("source payload exceeds 16 MiB")
			}
			if resp.StatusCode == http.StatusOK && readErr == nil {
				return macro.RawDocument{Source: source, URL: publicURL, ContentType: resp.Header.Get("Content-Type"), Body: body, FetchedAt: time.Now().UTC()}, nil
			}
		}
		retryable := requestErr != nil || readErr != nil || resp.StatusCode == 429 || resp.StatusCode >= 500
		if !retryable || attempt == 2 {
			if requestErr != nil {
				return macro.RawDocument{}, fmt.Errorf("source %s request failed", source)
			}
			if readErr != nil {
				return macro.RawDocument{}, fmt.Errorf("source %s response body could not be read", source)
			}
			return macro.RawDocument{}, fmt.Errorf("source %s returned HTTP %d", source, resp.StatusCode)
		}
		timer := time.NewTimer(time.Duration(attempt+1) * time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return macro.RawDocument{}, ctx.Err()
		case <-timer.C:
		}
	}
	return macro.RawDocument{}, fmt.Errorf("source request failed")
}
