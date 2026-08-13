package macrodata

import (
	"context"
	"testing"
	"time"

	dom "github.com/trungvdn/investment-ai/apps/macro-api/internal/domain/macrodata"
)

func TestInMemoryRepoSaveAndQuery(t *testing.T) {
	repo := NewInMemoryRepository()
	ctx := context.Background()
	d1 := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	o := dom.MacroObservation{IndicatorCode: "VN_M2_GROWTH", Value: 1.0, ObservationDate: d1, Unit: "%", Source: "MOCK", CollectedAt: d1.Add(time.Hour)}
	if err := repo.Save(ctx, o); err != nil {
		t.Fatalf("save err %v", err)
	}
	latest, _ := repo.FindLatest(ctx, "VN_M2_GROWTH")
	if latest == nil || latest.Value != 1.0 {
		t.Fatalf("latest mismatch")
	}

	from := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	list, _ := repo.FindByIndicator(ctx, "VN_M2_GROWTH", &from, &to)
	if len(list) != 1 {
		t.Fatalf("expected 1")
	}
}
