package wyckoff

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestSampleBaseline records reproducible observations, not accuracy: folder
// labels describe the user's assessment, not per-event ground truth.
func TestSampleBaseline(t *testing.T) {
	files, err := filepath.Glob("../../data/sample/*/*.csv")
	if err != nil || len(files) == 0 {
		t.Fatalf("sample discovery: files=%d err=%v", len(files), err)
	}
	for _, path := range files {
		t.Run(filepath.Base(path), func(t *testing.T) {
			f, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			rows, err := csv.NewReader(f).ReadAll()
			if err != nil {
				t.Fatal(err)
			}
			if len(rows) < 61 {
				t.Fatal("insufficient sample history")
			}
			if strings.Join(rows[0], ",") != "date,open,high,low,close,volume,event" {
				t.Fatal("unexpected CSV schema")
			}
			candles := make([]Candle, 0, len(rows)-1)
			for i, row := range rows[1:] {
				if len(row) != 7 {
					t.Fatalf("row %d: expected 7 columns", i+2)
				}
				date, err := time.Parse("2006-01-02", row[0])
				if err != nil {
					t.Fatal(err)
				}
				if len(candles) > 0 && !date.After(candles[len(candles)-1].Date) {
					t.Fatalf("row %d: duplicate or unordered date", i+2)
				}
				var values [5]float64
				for j := range values {
					values[j], err = strconv.ParseFloat(row[j+1], 64)
					if err != nil {
						t.Fatalf("row %d column %d: %v", i+2, j+2, err)
					}
				}
				candles = append(candles, Candle{Date: date, Open: values[0], High: values[1], Low: values[2], Close: values[3], Volume: values[4]})
			}
			symbol := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(path), "candles_"), "_1D.csv")
			// Prefix replay prevents the analyzer from receiving future candles.
			for _, n := range []int{60, len(candles)} {
				result, err := NewAnalyzer(DefaultConfig()).Analyze(symbol, candles[:n])
				if err != nil {
					t.Fatal(err)
				}
				if !result.AsOf.Equal(candles[n-1].Date) {
					t.Fatal("asOf differs from input cutoff")
				}
				for _, event := range result.Events {
					if event.Date.After(result.AsOf) {
						t.Fatal("event after cutoff")
					}
				}
				record := struct {
					Group        string   `json:"userAssessmentGroup"`
					Observations int      `json:"observations"`
					Result       Analysis `json:"result"`
				}{filepath.Base(filepath.Dir(path)), n, result}
				data, err := json.Marshal(record)
				if err != nil {
					t.Fatal(err)
				}
				fmt.Printf("BASELINE %s\n", data)
			}
		})
	}
}
