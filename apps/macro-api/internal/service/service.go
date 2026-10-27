package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/trungvdn/investment-ai/apps/macro-api/internal/macro"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/providers"
)

type Repository interface {
	SaveDocument(context.Context, macro.RawDocument) (int64, error)
	Save(context.Context, macro.Observation) error
	RecordAttempt(context.Context, string, macro.Attempt) error
	History(context.Context, string, *time.Time, *time.Time, bool) ([]macro.Observation, error)
	Attempts(context.Context) (map[string]macro.Attempt, error)
}
type Service struct {
	Repo      Repository
	Providers []providers.Provider
	Logger    *slog.Logger
}
type Failure struct {
	RetryAt  *time.Time `json:"retryAt,omitempty"`
	Provider string     `json:"provider"`
	Kind     string     `json:"kind"`
	Message  string     `json:"message"`
}
type RunResult struct {
	Failures []Failure `json:"failures"`
	Saved    int       `json:"saved"`
	Failed   int       `json:"failedSources"`
}

func (s *Service) Ingest(ctx context.Context) (RunResult, error) {
	result := RunResult{}
	// List of providers to ingest from, in order. If a provider fails, we still try the others.
	if s.Logger != nil {
		s.Logger.Debug("ingest run started", "providers", s.Providers)
	}
	for _, p := range s.Providers {
		if ctx.Err() != nil {
			return result, ctx.Err()
		}
		batch, fetchErr := p.Fetch(ctx)
		if s.Logger != nil {
			s.Logger.Debug("provider fetch completed", "provider", p.ID(), "observations", len(batch.Observations), "warnings", len(batch.Warnings), "unavailable", batch.Unavailable, "error", fetchErr)
		}
		docs := append([]macro.RawDocument{batch.Document}, batch.Documents...)
		ids := make([]int64, len(docs))
		for i, doc := range docs {
			if len(doc.Body) == 0 {
				continue
			}
			id, err := s.Repo.SaveDocument(ctx, doc)
			if err != nil {
				return result, fmt.Errorf("persist raw document: %w", err)
			}
			ids[i] = id
		}
		counts := map[string]int{}
		if fetchErr == nil {
			seen := make(map[struct {
				code     string
				period   string
				seriesID string
			}]struct{}, len(batch.Observations))
			for _, o := range batch.Observations {
				if err := macro.Validate(o, time.Now()); err != nil || o.RawIndex < 0 || o.RawIndex >= len(ids) || ids[o.RawIndex] == 0 {
					fetchErr = fmt.Errorf("invalid normalized data")
					break
				}
				key := struct {
					code     string
					period   string
					seriesID string
				}{o.IndicatorCode, o.Period, o.SeriesID}
				if _, exists := seen[key]; exists {
					fetchErr = fmt.Errorf("duplicate normalized observation")
					break
				}
				seen[key] = struct{}{}
			}
			if fetchErr == nil {
				for _, o := range batch.Observations {
					o.RawID = ids[o.RawIndex]
					o.CollectedAt = docs[o.RawIndex].FetchedAt
					if err := s.Repo.Save(ctx, o); err != nil {
						return result, fmt.Errorf("persist normalized observation: %w", err)
					}
					counts[o.IndicatorCode]++
					result.Saved++
				}
			}
		}
		missing := false
		for _, code := range p.Codes() {
			if counts[code] == 0 && !strings.HasPrefix(batch.Unavailable[code], "AWAITING_RELEASE:") &&
				!strings.HasPrefix(batch.Unavailable[code], "NOT_COMPARABLE:") &&
				!strings.HasPrefix(batch.Unavailable[code], "NO_PUBLISHED_RATE:") {
				missing = true
			}
		}
		if fetchErr == nil && missing {
			result.Failed++
		}
		if fetchErr != nil {
			result.Failed++
			result.Failures = append(result.Failures, Failure{
				RetryAt:  providers.RetryAt(fetchErr),
				Provider: p.ID(),
				Kind:     providers.ErrorKind(fetchErr),
				Message:  fetchErr.Error(),
			})
			if s.Logger != nil {
				s.Logger.Warn("macro source unavailable", "provider", p.ID(), "reason", fetchErr.Error())
			}
		}
		for _, code := range p.Codes() {
			a := macro.Attempt{Code: code, CheckedAt: time.Now().UTC(), Success: fetchErr == nil && counts[code] > 0 && batch.Unavailable[code] == ""}
			if fetchErr != nil {
				a.Message = fetchErr.Error()
			} else if !a.Success {
				a.Message = batch.Unavailable[code]
				if a.Message == "" {
					a.Message = "No unambiguous observations for this indicator in source response"
				}
			} else {
				a.Message = fmt.Sprintf("Processed %d observations; %d parser warnings", counts[code], len(batch.Warnings))
			}
			if err := s.Repo.RecordAttempt(ctx, p.ID(), a); err != nil {
				return result, err
			}
		}
	}
	return result, nil
}

type Dashboard struct {
	Mode        string       `json:"mode"`
	GeneratedAt time.Time    `json:"generatedAt"`
	Available   int          `json:"available"`
	Enabled     int          `json:"enabled"`
	Indicators  []macro.Card `json:"indicators"`
	Score       *float64     `json:"score"`
	ScoreStatus string       `json:"scoreStatus"`
}

func (s *Service) Dashboard(ctx context.Context, now time.Time) (Dashboard, error) {
	attempts, err := s.Repo.Attempts(ctx)
	if err != nil {
		return Dashboard{}, err
	}
	out := Dashboard{Mode: "live", GeneratedAt: now, Indicators: []macro.Card{}, ScoreStatus: "NOT_CONFIGURED"}
	for _, ind := range macro.Catalog() {
		card := macro.Card{Indicator: ind, History: []macro.Observation{}}
		if ind.Enabled {
			out.Enabled++
			all, err := s.Repo.History(ctx, ind.Code, nil, nil, false)
			if err != nil {
				return Dashboard{}, err
			}
			if len(all) > 0 {
				latest := all[len(all)-1]
				card.Latest = &latest
				out.Available++
			}
			cutoff := now.AddDate(-3, 0, 0)
			for _, o := range all {
				if !o.PeriodEnd.Before(cutoff) {
					card.History = append(card.History, o)
				}
			}
		}
		if a, ok := attempts[ind.Code]; ok {
			card.LastAttempt = &a
		}
		card.Status = macro.Status(ind, card.Latest, card.LastAttempt, now)
		out.Indicators = append(out.Indicators, card)
	}
	return out, nil
}
