package wyckoff

import "testing"

// Evidence checkpoints verify dated observations without promoting a candidate
// to a confirmed market event or claiming these are earliest confirmation dates.
func TestReviewedConfirmationCheckpoints(t *testing.T) {
	for _, symbol := range []string{"MSB", "HDG", "NLG"} {
		s, cs := readReviewed(t, "../../data/annotations/"+symbol+"/annotation.json")
		bars := map[string]Candle{}
		for _, c := range cs {
			bars[c.Date.Format("2006-01-02")] = c
		}
		bar := func(date string) Candle {
			c, ok := bars[date]
			if !ok || date > s.Cutoff {
				t.Fatalf("missing/future evidence %s %s", symbol, date)
			}
			return c
		}
		switch symbol {
		case "MSB":
			if bar("2026-08-28").Close <= 12.808 {
				t.Fatal("Spring recovery evidence missing")
			}
			if bar("2026-09-24").Close <= 13.726 || bar("2026-10-05").Low <= 13.726 || bar("2026-10-08").Close <= bar("2026-10-05").Close {
				t.Fatal("markup follow-through / held breakout evidence missing")
			}
		case "HDG":
			if bar("2026-09-04").Close <= 17.2 {
				t.Fatal("original breakout evidence missing")
			}
			if bar("2026-10-02").Close >= 15.2 || bar("2026-10-05").Close >= 15.2 {
				t.Fatal("failed accumulation continuation evidence missing")
			}
		case "NLG":
			if bar("2026-10-15").Low >= 21.3 || bar("2026-10-15").Low <= 19.15 {
				t.Fatal("small-range penetration must not be broad-range Spring")
			}
			if bar("2026-10-23").Close <= 22.5 || bar("2026-10-23").High >= 25.95 {
				t.Fatal("small-range SOS must not imply broad-range breakout")
			}
		}
	}
}
