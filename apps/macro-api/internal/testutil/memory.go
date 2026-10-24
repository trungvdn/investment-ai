// Package testutil provides fixtures for tests only. Production requires PostgreSQL.
package testutil

import (
	"context"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/macro"
	"sort"
	"time"
)

type Memory struct {
	Rows   []macro.Observation
	Raw    []macro.RawDocument
	Checks map[string]macro.Attempt
}

func (m *Memory) SaveDocument(_ context.Context, d macro.RawDocument) (int64, error) {
	m.Raw = append(m.Raw, d)
	return int64(len(m.Raw)), nil
}
func (m *Memory) Save(_ context.Context, o macro.Observation) error {
	m.Rows = append(m.Rows, o)
	return nil
}
func (m *Memory) RecordAttempt(_ context.Context, _ string, a macro.Attempt) error {
	if m.Checks == nil {
		m.Checks = map[string]macro.Attempt{}
	}
	m.Checks[a.Code] = a
	return nil
}
func (m *Memory) History(_ context.Context, code string, from, to *time.Time, _ bool) ([]macro.Observation, error) {
	out := []macro.Observation{}
	for _, o := range m.Rows {
		if o.IndicatorCode == code && (from == nil || !o.PeriodEnd.Before(*from)) && (to == nil || !o.PeriodEnd.After(*to)) {
			out = append(out, o)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Period < out[j].Period })
	return out, nil
}
func (m *Memory) Attempts(context.Context) (map[string]macro.Attempt, error) { return m.Checks, nil }
