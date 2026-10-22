package marketdata

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/trungvdn/investment-ai/apps/stock-analysis/internal/wyckoff"
)

// Provider fetches daily Vietnamese equity OHLCV candles.
type Provider interface {
	FetchDaily(ctx context.Context, symbol string, from, to time.Time) ([]wyckoff.Candle, error)
}

// KBSSource uses the publicly reachable KB Securities history endpoint. This
// endpoint is third-party/unofficial; production deployments should consider a
// licensed provider or the official TCBS OpenAPI (which requires credentials).
type KBSSource struct {
	Client  *http.Client
	BaseURL string
}

func NewKBSSource() *KBSSource {
	return &KBSSource{Client: &http.Client{Timeout: 15 * time.Second}, BaseURL: "https://kbbuddywts.kbsec.com.vn/iis-server/investment/stocks"}
}

type response struct {
	Symbol  string `json:"symbol"`
	DataDay []struct {
		T string  `json:"t"`
		O float64 `json:"o"`
		H float64 `json:"h"`
		L float64 `json:"l"`
		C float64 `json:"c"`
		V float64 `json:"v"`
	} `json:"data_day"`
}

func (p *KBSSource) FetchDaily(ctx context.Context, symbol string, from, to time.Time) ([]wyckoff.Candle, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	if from.After(to) {
		return nil, fmt.Errorf("from must not be after to")
	}
	var candles []wyckoff.Candle
	for chunkFrom := from; !chunkFrom.After(to); {
		chunkTo := chunkFrom.AddDate(1, 0, 0).Add(-24 * time.Hour)
		if chunkTo.After(to) {
			chunkTo = to
		}
		chunk, err := p.fetchDailyRange(ctx, symbol, chunkFrom, chunkTo)
		if err != nil {
			return nil, err
		}
		candles = append(candles, chunk...)
		chunkFrom = chunkTo.AddDate(0, 0, 1)
	}
	sort.Slice(candles, func(i, j int) bool { return candles[i].Date.Before(candles[j].Date) })
	return candles, nil
}

func (p *KBSSource) fetchDailyRange(ctx context.Context, symbol string, from, to time.Time) ([]wyckoff.Candle, error) {
	u, err := url.Parse(strings.TrimRight(p.BaseURL, "/") + "/" + url.PathEscape(symbol) + "/data_day")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("sdate", from.Format("02-01-2006"))
	q.Set("edate", to.Format("02-01-2006"))
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,vi;q=0.8")
	req.Header.Set("x-lang", "vi")
	client := p.Client
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch market data: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("market data provider returned HTTP %d for %s through %s: %s", resp.StatusCode, from.Format("2006-01-02"), to.Format("2006-01-02"), strings.TrimSpace(string(body)))
	}
	var payload response
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode market data: %w", err)
	}
	candles := make([]wyckoff.Candle, 0, len(payload.DataDay))
	for _, row := range payload.DataDay {
		var date time.Time
		for _, layout := range []string{"2006-01-02 15:04", "2006-01-02", "2006-01-02T15:04:05Z07:00"} {
			date, err = time.Parse(layout, row.T)
			if err == nil {
				break
			}
		}
		if err != nil {
			return nil, fmt.Errorf("parse candle date %q: %w", row.T, err)
		}
		// KB Securities endpoint reports prices in raw units scaled by 1,000.
		candles = append(candles, wyckoff.Candle{Date: date, Open: row.O / 1000, High: row.H / 1000, Low: row.L / 1000, Close: row.C / 1000, Volume: row.V})
	}
	if len(candles) == 0 {
		return nil, fmt.Errorf("provider returned no candles for %s", symbol)
	}
	return candles, nil
}
