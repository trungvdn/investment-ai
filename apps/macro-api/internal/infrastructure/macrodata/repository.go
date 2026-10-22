package macrodata

import (
	"context"
	"sort"
	"sync"
	"time"

	dom "github.com/trungvdn/investment-ai/apps/macro-api/internal/domain/macrodata"
)

type InMemoryMacroObservationRepository struct {
	mu           sync.RWMutex
	observations map[string][]dom.MacroObservation // keyed by indicator
}

func NewInMemoryRepository() *InMemoryMacroObservationRepository {
	return &InMemoryMacroObservationRepository{
		observations: make(map[string][]dom.MacroObservation),
	}
}

func (r *InMemoryMacroObservationRepository) Save(ctx context.Context, o dom.MacroObservation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	list := r.observations[o.IndicatorCode]
	// dedupe by observation date
	for _, ex := range list {
		if ex.ObservationDate.Equal(o.ObservationDate) {
			return nil
		}
	}
	r.observations[o.IndicatorCode] = append(list, o)
	sort.Slice(r.observations[o.IndicatorCode], func(i, j int) bool {
		return r.observations[o.IndicatorCode][i].ObservationDate.Before(r.observations[o.IndicatorCode][j].ObservationDate)
	})
	return nil
}

func (r *InMemoryMacroObservationRepository) FindLatest(ctx context.Context, indicatorCode string) (*dom.MacroObservation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := r.observations[indicatorCode]
	if len(list) == 0 {
		return nil, nil
	}
	latest := list[len(list)-1]
	return &latest, nil
}

func (r *InMemoryMacroObservationRepository) FindByIndicator(ctx context.Context, indicatorCode string, from, to *time.Time) ([]dom.MacroObservation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := r.observations[indicatorCode]
	if from == nil && to == nil {
		return append([]dom.MacroObservation(nil), list...), nil
	}
	var res []dom.MacroObservation
	for _, o := range list {
		if from != nil && o.ObservationDate.Before(*from) {
			continue
		}
		if to != nil && o.ObservationDate.After(*to) {
			continue
		}
		res = append(res, o)
	}
	return res, nil
}

func (r *InMemoryMacroObservationRepository) FindByIndicators(ctx context.Context, indicatorCodes []string, from, to *time.Time) (map[string][]dom.MacroObservation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string][]dom.MacroObservation)
	for _, code := range indicatorCodes {
		list := r.observations[code]
		if from == nil && to == nil {
			out[code] = append([]dom.MacroObservation(nil), list...)
			continue
		}
		var res []dom.MacroObservation
		for _, o := range list {
			if from != nil && o.ObservationDate.Before(*from) {
				continue
			}
			if to != nil && o.ObservationDate.After(*to) {
				continue
			}
			res = append(res, o)
		}
		out[code] = res
	}
	return out, nil
}
