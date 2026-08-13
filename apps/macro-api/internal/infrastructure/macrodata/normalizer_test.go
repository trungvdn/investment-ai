package macrodata

import (
	"testing"
)

func TestNormalizer(t *testing.T) {
	n := NewNormalizer()
	raw := RawMacroObservation{Indicator: "VN_USD_VND", Value: 23400, Date: "2026-06-01", Unit: ""}
	obs, err := n.Normalize(raw)
	if err != nil {
		t.Fatalf("normalize err %v", err)
	}
	if obs.Unit != "VND" {
		t.Fatalf("unit mismatch got %s", obs.Unit)
	}
}
