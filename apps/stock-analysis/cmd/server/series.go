package main

import (
	"github.com/trungvdn/investment-ai/apps/stock-analysis/internal/wyckoff"
	"sort"
	"time"
)

// prepareCandles selects inclusive calendar dates, orders/deduplicates observations,
// then aggregates weekly OHLCV before analysis so chart and report share one series.
func prepareCandles(input []wyckoff.Candle, from, to time.Time, timeframe string) []wyckoff.Candle {
	candles := make([]wyckoff.Candle, 0, len(input))
	first, last := from.Format("2006-01-02"), to.Format("2006-01-02")
	for _, c := range input {
		date := c.Date.Format("2006-01-02")
		if date >= first && date <= last {
			candles = append(candles, c)
		}
	}
	sort.SliceStable(candles, func(i, j int) bool { return candles[i].Date.Before(candles[j].Date) })
	unique := make([]wyckoff.Candle, 0, len(candles))
	for _, c := range candles {
		if len(unique) > 0 && unique[len(unique)-1].Date.Format("2006-01-02") == c.Date.Format("2006-01-02") {
			unique[len(unique)-1] = c
		} else {
			unique = append(unique, c)
		}
	}
	if timeframe != "1W" {
		return unique
	}
	weekly := make([]wyckoff.Candle, 0)
	var previousYear, previousWeek int
	for _, c := range unique {
		year, week := c.Date.ISOWeek()
		if len(weekly) == 0 || year != previousYear || week != previousWeek {
			weekly = append(weekly, c)
		} else {
			bar := &weekly[len(weekly)-1]
			if c.High > bar.High {
				bar.High = c.High
			}
			if c.Low < bar.Low {
				bar.Low = c.Low
			}
			bar.Close, bar.Date = c.Close, c.Date
			bar.Volume += c.Volume
		}
		previousYear, previousWeek = year, week
	}
	return weekly
}
