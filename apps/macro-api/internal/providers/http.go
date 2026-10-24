package providers

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/macro"
	"io"
	"math/rand/v2"
	"net"
	"net/http"
	"strconv"
	"time"
)

type Provider interface {
	ID() string
	Codes() []string
	Fetch(context.Context) (macro.Batch, error)
}
type HTTP struct {
	Client  *http.Client
	Headers http.Header
	Wait    func(context.Context, time.Duration) error
}

type SourceError struct {
	Source, Kind string
	Status       int
	RetryAt      time.Time
}

func (e *SourceError) Error() string {
	if e.Status != 0 {
		return fmt.Sprintf("source %s returned HTTP %d", e.Source, e.Status)
	}
	return fmt.Sprintf("source %s %s", e.Source, e.Kind)
}
func ErrorKind(err error) string {
	var e *SourceError
	if errors.As(err, &e) {
		return e.Kind
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return "TIMEOUT"
	}
	if err != nil && (err.Error() == PMINotConfigured || err.Error() == DXYNotConfigured || errors.Is(err, ErrNotConfigured)) {
		return "NOT_CONFIGURED"
	}
	return "DATA_QUALITY"
}
func retryDelay(header string, now time.Time, attempt int) time.Duration {
	d := time.Second*time.Duration(1<<attempt) + time.Duration(rand.IntN(500))*time.Millisecond
	if n, e := strconv.Atoi(header); e == nil && n >= 0 {
		if uint64(n) > uint64((1<<63-1)/int64(time.Second)) {
			return time.Duration(1<<63 - 1)
		}
		d = time.Duration(n) * time.Second
	} else if t, e := http.ParseTime(header); e == nil && t.After(now) {
		d = t.Sub(now)
	}

	return d
}
func waitRetry(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (h HTTP) Get(ctx context.Context, requestURL, publicURL, source string) (macro.RawDocument, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return macro.RawDocument{}, fmt.Errorf("invalid source URL")
	}
	req.Header = h.Headers.Clone()
	if req.Header == nil {
		req.Header = make(http.Header)
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
				kind := "TRANSPORT"
				var dns *net.DNSError
				var cert *tls.CertificateVerificationError
				var timeout net.Error
				if errors.As(requestErr, &cert) {
					kind = "TLS"
				} else if errors.As(requestErr, &dns) {
					kind = "DNS"
				} else if errors.As(requestErr, &timeout) && timeout.Timeout() {
					kind = "TIMEOUT"
				}
				return macro.RawDocument{}, &SourceError{Source: source, Kind: kind}
			}
			if readErr != nil {
				return macro.RawDocument{}, &SourceError{Source: source, Kind: "BODY_READ"}
			}
			kind := "HTTP"
			if resp.StatusCode == 403 {
				kind = "ACCESS_DENIED"
			}
			if resp.StatusCode == 429 {
				kind = "RATE_LIMIT"
			}
			if resp.StatusCode >= 500 {
				kind = "UPSTREAM"
			}
			e := &SourceError{Source: source, Kind: kind, Status: resp.StatusCode}
			if retryable && resp.Header.Get("Retry-After") != "" {
				e.RetryAt = time.Now().Add(retryDelay(resp.Header.Get("Retry-After"), time.Now(), attempt))
			}
			return macro.RawDocument{}, e
		}
		header := ""
		if resp != nil {
			header = resp.Header.Get("Retry-After")
		}
		wait := h.Wait
		if wait == nil {
			wait = waitRetry
		}
		delay := retryDelay(header, time.Now(), attempt)
		if delay > 60*time.Second {
			return macro.RawDocument{}, &SourceError{Source: source, Kind: "RATE_LIMIT", Status: resp.StatusCode, RetryAt: time.Now().Add(delay)}
		}
		if err := wait(ctx, delay); err != nil {
			return macro.RawDocument{}, err
		}
	}
	return macro.RawDocument{}, fmt.Errorf("source request failed")
}

func RetryAt(err error) *time.Time {
	var e *SourceError
	if errors.As(err, &e) && !e.RetryAt.IsZero() {
		return &e.RetryAt
	}
	return nil
}
