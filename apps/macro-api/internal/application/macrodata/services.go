package macrodata

import (
	"context"
	"time"

	dom "github.com/trungvdn/investment-ai/apps/macro-api/internal/domain/macrodata"
	infra "github.com/trungvdn/investment-ai/apps/macro-api/internal/infrastructure/macrodata"
)

type MacroDataProvider interface {
	GetObservations(ctx context.Context) ([]infra.RawMacroObservation, error)
}

type MacroObservationRepository interface {
	Save(ctx context.Context, o dom.MacroObservation) error
	FindLatest(ctx context.Context, indicatorCode string) (*dom.MacroObservation, error)
	FindByIndicator(ctx context.Context, indicatorCode string, from, to *time.Time) ([]dom.MacroObservation, error)
	FindByIndicators(ctx context.Context, indicatorCodes []string, from, to *time.Time) (map[string][]dom.MacroObservation, error)
}

type MacroDataIngestionService struct {
	provider   MacroDataProvider
	normalizer *infra.Normalizer
	repository MacroObservationRepository
}

func NewIngestionService(p MacroDataProvider, n *infra.Normalizer, r MacroObservationRepository) *MacroDataIngestionService {
	return &MacroDataIngestionService{provider: p, normalizer: n, repository: r}
}

type IngestionResult struct {
	Saved   int
	Skipped int
	Errors  []error
}

func (s *MacroDataIngestionService) Run(ctx context.Context) (IngestionResult, error) {
	raws, err := s.provider.GetObservations(ctx)
	if err != nil {
		return IngestionResult{}, err
	}
	var res IngestionResult
	for _, r := range raws {
		obs, err := s.normalizer.Normalize(r)
		if err != nil {
			res.Errors = append(res.Errors, err)
			continue
		}
		if err := dom.ValidateObservation(obs); err != nil {
			res.Errors = append(res.Errors, err)
			res.Skipped++
			continue
		}
		if err := s.repository.Save(ctx, obs); err != nil {
			res.Errors = append(res.Errors, err)
			continue
		}
		res.Saved++
	}
	return res, nil
}

type MacroDataQueryService struct {
	repo MacroObservationRepository
}

func NewQueryService(r MacroObservationRepository) *MacroDataQueryService {
	return &MacroDataQueryService{repo: r}
}

func (q *MacroDataQueryService) ListIndicators(ctx context.Context) []dom.MacroIndicator {
	// hard-coded catalog for now
	return []dom.MacroIndicator{
		{Code: "VN_M2_GROWTH", Name: "M2 Growth", Category: dom.CategoryLiquidity, Unit: "%", Frequency: dom.FrequencyMonthly, Country: "VN", Active: true},
		{Code: "VN_CREDIT_GROWTH", Name: "Credit Growth", Category: dom.CategoryCredit, Unit: "%", Frequency: dom.FrequencyMonthly, Country: "VN", Active: true},
		{Code: "VN_INTERBANK_ON", Name: "Interbank ON", Category: dom.CategoryLiquidity, Unit: "%", Frequency: dom.FrequencyMonthly, Country: "VN", Active: true},
		{Code: "VN_SBV_TBILL", Name: "SBV TBill", Category: dom.CategoryLiquidity, Unit: "%", Frequency: dom.FrequencyMonthly, Country: "VN", Active: true},
		{Code: "VN_USD_VND", Name: "USD/VND", Category: dom.CategoryExternal, Unit: "VND", Frequency: dom.FrequencyMonthly, Country: "VN", Active: true},
	}
}

func (q *MacroDataQueryService) GetLatest(ctx context.Context, code string) (*dom.MacroObservation, error) {
	return q.repo.FindLatest(ctx, code)
}

func (q *MacroDataQueryService) GetHistorical(ctx context.Context, code string, from, to *time.Time) ([]dom.MacroObservation, error) {
	return q.repo.FindByIndicator(ctx, code, from, to)
}

func (q *MacroDataQueryService) QueryByIndicators(ctx context.Context, codes []string, from, to *time.Time) (map[string][]dom.MacroObservation, error) {
	return q.repo.FindByIndicators(ctx, codes, from, to)
}
