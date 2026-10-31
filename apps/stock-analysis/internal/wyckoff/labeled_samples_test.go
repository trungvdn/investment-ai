package wyckoff

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// This evaluates sparse positive labels only. Empty cells are unannotated,
// so unmatched predictions cannot be counted as false positives.
func TestLabeledSampleRecall(t *testing.T) {
	files, err := filepath.Glob("../../data/sample/*.csv")
	if err != nil || len(files) == 0 {
		t.Fatalf("sample discovery: %v (%d files)", err, len(files))
	}
	allowed := map[string]bool{"PS": true, "SC": true, "AR": true, "ST": true, "SPRING": true, "TEST": true, "SOS": true, "LPS": true, "BC": true, "UT": true, "UTAD": true, "SOW": true, "LPSY": true}
	total, matched := 0, 0
	for _, path := range files {
		t.Run(filepath.Base(path), func(t *testing.T) {
			file, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()
			rows, err := csv.NewReader(file).ReadAll()
			if err != nil {
				t.Fatal(err)
			}
			if len(rows) < 61 || strings.Join(rows[0], ",") != "date,open,high,low,close,volume,event" {
				t.Fatal("invalid sample schema or insufficient history")
			}
			var candles []Candle
			labels := make(map[int]string)
			for i, row := range rows[1:] {
				if len(row) != 7 {
					t.Fatalf("row %d: expected 7 columns", i+2)
				}
				date, err := time.Parse("2006-01-02", row[0])
				if err != nil {
					t.Fatal(err)
				}
				if i > 0 && !date.After(candles[i-1].Date) {
					t.Fatal("unordered dates")
				}
				var v [5]float64
				for j := range v {
					v[j], err = strconv.ParseFloat(row[j+1], 64)
					if err != nil {
						t.Fatal(err)
					}
				}
				candles = append(candles, Candle{Date: date, Open: v[0], High: v[1], Low: v[2], Close: v[3], Volume: v[4]})
				label := strings.ToUpper(strings.TrimSpace(row[6]))
				if label != "" {
					if !allowed[label] {
						t.Fatalf("unsupported label %q at %s", label, row[0])
					}
					labels[i] = label
				}
			}
			symbol := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(path), "candles_"), "_1D.csv")
			// Exercise OHLCV validation even when annotations are absent.
			if _, err := NewAnalyzer(DefaultConfig()).Analyze(symbol, candles); err != nil {
				t.Fatal(err)
			}
			for idx, label := range labels {
				// Permit up to 15 subsequent observed sessions for retrospective
				// confirmation. This does not measure real-time detection latency.
				hit := false
				for cutoff := idx + 1; cutoff <= idx+16 && cutoff <= len(candles); cutoff++ {
					if cutoff < 60 {
						continue
					}
					result, err := NewAnalyzer(DefaultConfig()).Analyze(symbol, candles[:cutoff])
					if err != nil {
						t.Fatal(err)
					}
					for _, event := range result.Events {
						if event.Date.After(result.AsOf) {
							t.Fatal("event exceeds cutoff")
						}
						if event.Type == label && event.Date.Equal(candles[idx].Date) {
							hit = true
						}
					}
				}
				total++
				if hit {
					matched++
				}
				t.Logf("LABEL symbol=%s date=%s event=%s exactDateMatchedWithin15Sessions=%t", symbol, candles[idx].Date.Format("2006-01-02"), label, hit)
			}
			t.Logf("COVERAGE symbol=%s candles=%d positiveLabels=%d", symbol, len(candles), len(labels))
		})
	}
	if total == 0 {
		t.Log("NOT_EVALUABLE: no machine-readable positive labels; PNG annotations are not parsed")
		return
	}
	t.Logf("SCORE positiveLabels=%d exactDateMatchedWithin15Sessions=%d; precision unavailable; no accuracy gate", total, matched)
}
