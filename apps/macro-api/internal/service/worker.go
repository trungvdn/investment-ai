package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/macro"
	"strings"
	"time"
)

type OperationalStore interface {
	ClaimSource(context.Context, string, time.Time) (string, error)
	FinishSource(context.Context, string, string, time.Time, time.Time, string, string) error
	SourceStates(context.Context) ([]macro.SourceState, error)
}
type Schedule map[string]time.Duration

func ParseSchedule(raw string, ids []string) (Schedule, error) {
	out := Schedule{}
	if raw == "" {
		return out, nil
	}
	var values map[string]string
	if err := json.Unmarshal([]byte(raw), &values); err != nil || values == nil {
		return nil, fmt.Errorf("INGEST_SCHEDULE must be a JSON object of provider IDs to durations")
	}
	for id, value := range values {
		known := false
		for _, v := range ids {
			if v == id {
				known = true
			}
		}
		if !known {
			return nil, fmt.Errorf("unknown provider in INGEST_SCHEDULE: %s", id)
		}
		d, err := time.ParseDuration(value)
		if err != nil || d < 5*time.Minute || d > 7*24*time.Hour {
			return nil, fmt.Errorf("source cadence must be between 5m and 168h")
		}
		out[id] = d
	}
	return out, nil
}
func (s Schedule) Cadence(id string, fallback time.Duration) time.Duration {
	if d, ok := s[id]; ok {
		return d
	}
	switch {
	case id == "sbv:fx" || id == "ice:dxy":
		return time.Hour
	case id == "sbv:interbank" || id == "sbv:bills" || id == "spglobal:pmi:vn":
		return 2 * time.Hour
	case strings.HasPrefix(id, "sbv:statistics:"):
		return 12 * time.Hour
	default:
		return fallback
	}
}
func nextDelay(kind string, failures int, cadence time.Duration) time.Duration {
	if kind == "" {
		return cadence
	}
	if kind == "NOT_CONFIGURED" {
		return 24 * time.Hour
	}
	if kind == "ACCESS_DENIED" || kind == "DATA_QUALITY" {
		return 6 * time.Hour
	}
	d := 5 * time.Minute
	for i := 1; i < failures && d < 6*time.Hour; i++ {
		d *= 3
	}
	if d > 6*time.Hour {
		d = 6 * time.Hour
	}
	return d
}

// Tick claims providers in PostgreSQL; replicas and a manual run cannot overlap.
func (s *Service) Tick(ctx context.Context, store OperationalStore, schedule Schedule, fallback time.Duration) error {
	for _, p := range s.Providers {
		if err := ctx.Err(); err != nil {
			return err
		}
		opCtx, opCancel := context.WithTimeout(ctx, 10*time.Second)
		token, err := store.ClaimSource(opCtx, p.ID(), time.Now().UTC())
		opCancel()
		if err != nil {
			return err
		}
		if token == "" {
			continue
		}
		run, cancel := context.WithTimeout(ctx, 2*time.Minute)
		single := *s
		single.Logger = nil
		single.Providers = s.Providers[:0:0]
		single.Providers = append(single.Providers, p)
		result, runErr := single.Ingest(run)
		cancel()
		kind, message := "", "Source checked successfully"
		if runErr != nil {
			kind = "STORAGE"
			if errors.Is(runErr, context.DeadlineExceeded) {
				kind = "TIMEOUT"
			}
			message = "Source ingestion storage/run failure"
		} else if len(result.Failures) > 0 {
			kind = result.Failures[0].Kind
			message = result.Failures[0].Message
		}
		opCtx, opCancel = context.WithTimeout(ctx, 10*time.Second)
		states, err := store.SourceStates(opCtx)
		opCancel()
		if err != nil {
			return err
		}
		failures := 1
		recovered := false
		for _, st := range states {
			if st.Provider == p.ID() {
				failures = st.Failures + 1
				recovered = st.Kind != "" && kind == ""
			}
		}
		now := time.Now().UTC()
		next := now.Add(nextDelay(kind, failures, schedule.Cadence(p.ID(), fallback)))
		if len(result.Failures) > 0 && result.Failures[0].RetryAt != nil && result.Failures[0].RetryAt.After(next) {
			next = *result.Failures[0].RetryAt
		}
		opCtx, opCancel = context.WithTimeout(ctx, 10*time.Second)
		err = store.FinishSource(opCtx, p.ID(), token, now, next, kind, message)
		opCancel()
		if err != nil {
			return err
		}
		if s.Logger != nil {
			if kind != "" {
				s.Logger.Warn("macro source alert", "provider", p.ID(), "kind", kind, "consecutiveFailures", failures, "nextDue", next, "reason", message)
			} else {
				if recovered {
					s.Logger.Info("macro source recovered", "provider", p.ID())
				}
				s.Logger.Info("macro source checked", "provider", p.ID(), "processed", result.Saved, "nextDue", next)
			}
		}
		if runErr != nil {
			return runErr
		}
	}
	return nil
}
