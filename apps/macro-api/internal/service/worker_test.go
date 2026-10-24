package service

import (
	"context"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/macro"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/providers"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/testutil"
	"testing"
	"time"
)

type workerStore struct {
	state macro.SourceState
	token string
}

func (s *workerStore) ClaimSource(_ context.Context, id string, now time.Time) (string, error) {
	if s.token != "" || s.state.NextDue.After(now) {
		return "", nil
	}
	s.state.Provider = id
	s.token = "owner"
	return s.token, nil
}
func (s *workerStore) FinishSource(_ context.Context, id, token string, now, next time.Time, kind, message string) error {
	s.token = ""
	s.state.LastChecked = &now
	s.state.NextDue = next
	s.state.Kind = kind
	s.state.Message = message
	if kind == "" {
		s.state.Failures = 0
		s.state.LastSuccess = &now
	} else {
		s.state.Failures++
	}
	return nil
}
func (s *workerStore) SourceStates(context.Context) ([]macro.SourceState, error) {
	return []macro.SourceState{s.state}, nil
}
func TestWorkerBackoffSurvivesRestartAndRecovers(t *testing.T) {
	repo := &testutil.Memory{}
	state := &workerStore{}
	p := fakeProvider{code: "VN_SBV_TBILL", err: &providers.SourceError{Source: "SBV", Kind: "ACCESS_DENIED", Status: 403}}
	svc := Service{Repo: repo, Providers: []providers.Provider{p}}
	if err := svc.Tick(context.Background(), state, Schedule{}, time.Hour); err != nil {
		t.Fatal(err)
	}
	if state.state.Failures != 1 || state.state.Kind != "ACCESS_DENIED" || state.state.NextDue.Before(time.Now().Add(5*time.Hour)) {
		t.Fatal("403 cooldown not durable")
	}
	restarted := Service{Repo: repo, Providers: svc.Providers}
	if err := restarted.Tick(context.Background(), state, Schedule{}, time.Hour); err != nil {
		t.Fatal(err)
	}
	if state.state.Failures != 1 {
		t.Fatal("restart bypassed next due")
	}
	state.state.NextDue = time.Now().Add(-time.Second)
	p.err = nil
	p.batch = macro.Batch{Unavailable: map[string]string{"VN_SBV_TBILL": "NO_PUBLISHED_RATE: no bill award"}}
	restarted.Providers = []providers.Provider{p}
	if err := restarted.Tick(context.Background(), state, Schedule{}, time.Hour); err != nil {
		t.Fatal(err)
	}
	if state.state.Failures != 0 || state.state.Kind != "" || state.state.LastSuccess == nil {
		t.Fatal("expected no-award did not clear source failure")
	}
}
func TestScheduleAndQualityGates(t *testing.T) {
	if _, err := ParseSchedule(`{"sbv:fx":"15m"}`, []string{"sbv:fx"}); err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{`{"unknown":"1h"}`, `{"sbv:fx":"1s"}`, `{"sbv:fx":"bad"}`, `{"sbv:fx":"999h"}`} {
		if _, err := ParseSchedule(raw, []string{"sbv:fx"}); err == nil {
			t.Fatal("invalid schedule accepted")
		}
	}
	if nextDelay("TIMEOUT", 2, time.Hour) != 15*time.Minute || nextDelay("NOT_CONFIGURED", 1, time.Hour) != 24*time.Hour {
		t.Fatal("incorrect source backoff")
	}
	repo := &testutil.Memory{}
	end, _ := macro.MonthEnd("2026-08")
	o := macro.Observation{IndicatorCode: "VN_CPI_YOY", Period: "2026-08", PeriodEnd: end, Value: 3, Unit: "% YoY", Source: "NSO", SourceURL: "https://www.nso.gov.vn/test", SeriesID: "CPI", Method: "YOY"}
	repo.Rows = []macro.Observation{o}
	o.Value = 4
	doc := macro.RawDocument{Source: "NSO", URL: o.SourceURL, Body: []byte("duplicate fixture"), FetchedAt: time.Now()}
	svc := Service{Repo: repo, Providers: []providers.Provider{fakeProvider{code: o.IndicatorCode, batch: macro.Batch{Document: doc, Observations: []macro.Observation{o, o}}}}}
	result, err := svc.Ingest(context.Background())
	if err != nil || result.Failed != 1 || len(repo.Rows) != 1 || repo.Rows[0].Value != 3 || len(repo.Raw) != 1 {
		t.Fatalf("bad batch overwrote good data %+v %v", result, err)
	}
}

type operationalRepo struct {
	testutil.Memory
	workerStore
}

func TestOperationalAlertsSeparateOverdueAndExpectedNoAward(t *testing.T) {
	now := time.Now()
	repo := &operationalRepo{workerStore: workerStore{state: macro.SourceState{Provider: "sbv:bills", NextDue: now.Add(-time.Hour), Failures: 2, Kind: "ACCESS_DENIED", Message: "403"}}}
	repo.Checks = map[string]macro.Attempt{"VN_SBV_TBILL": {Code: "VN_SBV_TBILL", Success: false, Message: "NO_PUBLISHED_RATE: no published bill yield"}}
	svc := Service{Repo: repo}
	report, err := svc.Operations(context.Background(), now)
	if err != nil {
		t.Fatal(err)
	}
	overdue, denied := false, false
	for _, a := range report.Alerts {
		if a.Status == "SCHEDULER_OVERDUE" {
			overdue = true
		}
		if a.Status == "ACCESS_DENIED" {
			denied = true
		}
		if a.Code == "VN_SBV_TBILL" {
			t.Fatal("normal no-award is not an indicator outage")
		}
	}
	if !overdue || !denied {
		t.Fatal("worker failure alert missing")
	}
}
