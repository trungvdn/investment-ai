package macrodata

import (
	"time"

	"github.com/trungvdn/investment-ai/apps/macro-api/internal/domain/macrodata"
)

type Normalizer struct{}

func NewNormalizer() *Normalizer { return &Normalizer{} }

func (n *Normalizer) Normalize(raw RawMacroObservation) (macrodata.MacroObservation, error) {
	// map units and parse date
	date, err := time.Parse("2006-01-02", raw.Date)
	if err != nil {
		return macrodata.MacroObservation{}, err
	}
	unit := raw.Unit
	if raw.Indicator == "VN_USD_VND" {
		unit = "VND"
	}
	obs := macrodata.MacroObservation{
		IndicatorCode:   raw.Indicator,
		Value:           raw.Value,
		ObservationDate: date,
		Unit:            unit,
		Source:          "MOCK",
		CollectedAt:     time.Now().UTC(),
		QualityStatus:   macrodata.QualityValid,
	}
	return obs, nil
}
