package service

import (
	"context"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/macro"
	"time"
)

type Alert struct {
	Code     string `json:"code"`
	Provider string `json:"provider,omitempty"`
	Status   string `json:"status"`
	Message  string `json:"message"`
}
type Operations struct {
	GeneratedAt time.Time           `json:"generatedAt"`
	Sources     []macro.SourceState `json:"sources"`
	Alerts      []Alert             `json:"alerts"`
}

func (s *Service) Operations(ctx context.Context, now time.Time) (Operations, error) {
	out := Operations{GeneratedAt: now, Sources: []macro.SourceState{}, Alerts: []Alert{}}
	if store, ok := s.Repo.(OperationalStore); ok {
		states, err := store.SourceStates(ctx)
		if err != nil {
			return out, err
		}
		out.Sources = states
		for _, st := range states {
			if st.Kind != "" {
				out.Alerts = append(out.Alerts, Alert{Provider: st.Provider, Status: st.Kind, Message: st.Message})
			}
			if !st.NextDue.IsZero() && now.After(st.NextDue.Add(15*time.Minute)) && (st.LeaseUntil == nil || now.After(*st.LeaseUntil)) {
				out.Alerts = append(out.Alerts, Alert{Provider: st.Provider, Status: "SCHEDULER_OVERDUE", Message: "Source collection is overdue; inspect worker health"})
			}
		}
	}
	dashboard, err := s.Dashboard(ctx, now)
	if err != nil {
		return out, err
	}
	for _, c := range dashboard.Indicators {
		if c.Status == "STALE" || c.Status == "MISSING" || c.Status == "NOT_CONFIGURED" || c.Status == "SOURCE_ERROR" || c.Status == "ERROR" {
			message := c.Description
			if c.LastAttempt != nil {
				message = c.LastAttempt.Message
			}
			out.Alerts = append(out.Alerts, Alert{Code: c.Code, Status: c.Status, Message: message})
		}
	}
	return out, nil
}
