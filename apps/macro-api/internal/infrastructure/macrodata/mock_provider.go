package macrodata

import (
	"context"
	"time"
)

type RawMacroObservation struct {
	Indicator string
	Value     float64
	Date      string
	Unit      string
}

type MockMacroDataProvider struct{}

func NewMockProvider() *MockMacroDataProvider { return &MockMacroDataProvider{} }

func (p *MockMacroDataProvider) GetObservations(ctx context.Context) ([]RawMacroObservation, error) {
	// deterministic mock data for 6 months
	base := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	var out []RawMacroObservation
	indicators := []string{"VN_M2_GROWTH", "VN_CREDIT_GROWTH", "VN_INTERBANK_ON", "VN_SBV_TBILL", "VN_USD_VND"}
	for i := 0; i < 6; i++ {
		d := base.AddDate(0, i, 0).Format("2006-01-02")
		for _, ind := range indicators {
			v := 1.0 + float64(i)
			if ind == "VN_USD_VND" {
				v = 23400 + float64(i)*10
			}
			out = append(out, RawMacroObservation{Indicator: ind, Value: v, Date: d, Unit: "%"})
		}
	}
	return out, nil
}
