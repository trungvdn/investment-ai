package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/trungvdn/investment-ai/apps/stock-analysis/internal/wyckoff"
)

const SchemaVersion = "1.0"

// Report is the stable, presentation-oriented contract for a stock analysis.
// It deliberately separates the human-readable summary from the underlying analysis.
type Report struct {
	SchemaVersion string             `json:"schemaVersion"`
	Symbol        string             `json:"symbol"`
	AsOf          time.Time          `json:"asOf"`
	Headline      string             `json:"headline"`
	Summary       string             `json:"summary"`
	Structure     wyckoff.Structure  `json:"structure"`
	Phase         wyckoff.Phase      `json:"phase"`
	Confidence    float64            `json:"confidence"`
	TradingRange  *RangeSection      `json:"tradingRange,omitempty"`
	Events        []EventSection     `json:"events"`
	Evidence      []string           `json:"evidence"`
	Risks         []string           `json:"risks"`
	Watchlist     []string           `json:"watchlist"`
	DataQuality   DataQualitySection `json:"dataQuality"`
	Disclaimer    string             `json:"disclaimer"`
}

type RangeSection struct {
	Status     string  `json:"status"`
	Support    float64 `json:"support"`
	Resistance float64 `json:"resistance"`
	WidthPct   float64 `json:"widthPct"`
}

type EventSection struct {
	Type       string    `json:"type"`
	Date       time.Time `json:"date"`
	Confidence float64   `json:"confidence"`
	Evidence   string    `json:"evidence"`
}

type DataQualitySection struct {
	Mode       string   `json:"mode"`
	DataPoints int      `json:"dataPoints"`
	Status     string   `json:"status"`
	Notes      []string `json:"notes"`
}

// Build converts an analyzer result into a deterministic, versioned report.
// It does not add new market facts or turn heuristic signals into trade advice.
func Build(a wyckoff.Analysis) Report {
	report := Report{
		SchemaVersion: SchemaVersion,
		Symbol:        strings.ToUpper(strings.TrimSpace(a.Symbol)),
		AsOf:          a.AsOf,
		Structure:     a.Structure,
		Phase:         a.Phase,
		Confidence:    clamp(a.Confidence, 0, 1),
		Events:        make([]EventSection, 0, len(a.Events)),
		Evidence:      append([]string{}, a.Evidence...),
		Risks:         []string{},
		Watchlist:     []string{},
		DataQuality: DataQualitySection{
			Mode:       a.DataMode,
			DataPoints: a.DataPoints,
			Status:     "OK",
			Notes:      []string{},
		},
		Disclaimer: "This report is an automated, heuristic Wyckoff analysis for research purposes only. It is not investment advice or a recommendation to buy or sell.",
	}

	report.Headline, report.Summary = describeStructure(a.Structure, a.Phase)
	for _, event := range a.Events {
		report.Events = append(report.Events, EventSection{
			Type:       event.Type,
			Date:       event.Date,
			Confidence: clamp(event.Confidence, 0, 1),
			Evidence:   event.Evidence,
		})
	}

	if a.TradingRange != nil {
		status := "PROVISIONAL"
		if a.Structure != wyckoff.StructureUnclear && a.Phase != wyckoff.PhaseUnknown {
			status = "IDENTIFIED"
		}
		report.TradingRange = &RangeSection{
			Status:     status,
			Support:    a.TradingRange.Support,
			Resistance: a.TradingRange.Resistance,
			WidthPct:   a.TradingRange.WidthPct,
		}
	} else {
		report.Risks = append(report.Risks, "No trading range was identified; support/resistance levels are unavailable.")
	}

	if a.Structure == wyckoff.StructureUnclear || a.Phase == wyckoff.PhaseUnknown {
		report.Risks = append(report.Risks, "Market structure or Wyckoff phase remains UNKNOWN; avoid treating candidate events as confirmation.")
		report.Watchlist = append(report.Watchlist, "Wait for a validated trading range and a sequence of events that confirms the structure.")
	}
	if a.Confidence < 0.5 {
		report.Risks = append(report.Risks, "Model confidence is low; conclusions may change as additional candles become available.")
	}
	if a.DataMode == "demo" {
		report.DataQuality.Status = "DEMO"
		report.DataQuality.Notes = append(report.DataQuality.Notes, "Demo candles are synthetic and must not be used to make real-market decisions.")
		report.Risks = append(report.Risks, "Input is demo data, not observed market data.")
	}
	if a.DataPoints < 60 {
		report.DataQuality.Status = "INSUFFICIENT"
		report.DataQuality.Notes = append(report.DataQuality.Notes, "Fewer than 60 candles were supplied; the analyzer's minimum-history requirement is not met.")
	}
	if a.DataQuality.Status == "OK" && a.DataPoints > 0 {
		report.DataQuality.Notes = append(report.DataQuality.Notes, "Data-point count meets the analyzer's minimum history requirement; this does not guarantee data completeness.")
	}
	if a.TradingRange != nil && a.TradingRange.WidthPct > 0 {
		report.Watchlist = append(report.Watchlist, fmt.Sprintf("Monitor price reaction around support %.2f and resistance %.2f; confirm with subsequent price and volume.", a.TradingRange.Support, a.TradingRange.Resistance))
	}
	if containsEvent(a.Events, "SPRING") {
		report.Watchlist = append(report.Watchlist, "Check whether a follow-up test holds above the Spring low with reduced supply.")
	}
	if containsEvent(a.Events, "SOS") {
		report.Watchlist = append(report.Watchlist, "Check whether a pullback holds above former range resistance before treating the breakout as sustained.")
	}
	if containsEvent(a.Events, "SOW") {
		report.Watchlist = append(report.Watchlist, "Monitor whether price remains below support and whether selling pressure persists.")
	}
	if len(report.Watchlist) == 0 {
		report.Watchlist = append(report.Watchlist, "Wait for a clear price/volume event before drawing a stronger conclusion.")
	}
	return report
}

func describeStructure(structure wyckoff.Structure, phase wyckoff.Phase) (string, string) {
	switch structure {
	case wyckoff.StructureAccumulation:
		return "Potential accumulation structure", fmt.Sprintf("The analyzer classifies the current structure as ACCUMULATION, phase %s. This is a heuristic interpretation and requires confirmation from subsequent price and volume.", phase)
	case wyckoff.StructureDistribution:
		return "Potential distribution structure", fmt.Sprintf("The analyzer classifies the current structure as DISTRIBUTION, phase %s. This is a heuristic interpretation and requires confirmation from subsequent price and volume.", phase)
	case wyckoff.StructureMarkup:
		return "Potential markup trend", "The analyzer classifies the current structure as MARKUP. Monitor whether price strength persists and prior resistance acts as support."
	case wyckoff.StructureMarkdown:
		return "Potential markdown trend", "The analyzer classifies the current structure as MARKDOWN. Monitor whether weakness persists and broken support continues to act as resistance."
	default:
		return "Market structure is not confirmed", "Available candles do not provide sufficient validated evidence to assign a clear Wyckoff structure and phase."
	}
}

func containsEvent(events []wyckoff.Event, kind string) bool {
	for _, event := range events {
		if event.Type == kind {
			return true
		}
	}
	return false
}

func clamp(value, minimum, maximum float64) float64 {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}
