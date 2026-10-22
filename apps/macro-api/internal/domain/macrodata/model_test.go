package macrodata

import (
	"testing"
	"time"
)

func TestValidateObservation(t *testing.T) {
	now := time.Now().UTC()
	o := MacroObservation{IndicatorCode: "VN_M2_GROWTH", Value: 1.2, ObservationDate: now, Unit: "%", Source: "MOCK", CollectedAt: now.Add(time.Hour)}
	if err := ValidateObservation(o); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}

	o2 := MacroObservation{IndicatorCode: "VN_INTERBANK_ON", Value: -1, ObservationDate: now, Unit: "%", Source: "MOCK", CollectedAt: now.Add(time.Hour)}
	if err := ValidateObservation(o2); err == nil {
		t.Fatalf("expected invalid negative value")
	}

	o3 := MacroObservation{IndicatorCode: "", Value: 1.0, ObservationDate: now, Unit: "%", Source: "MOCK", CollectedAt: now.Add(time.Hour)}
	if err := ValidateObservation(o3); err == nil {
		t.Fatalf("expected indicator not found")
	}
}
