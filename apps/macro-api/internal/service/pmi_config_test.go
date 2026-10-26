package service

import (
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/providers"
	"testing"
)

func TestPMIConfigurationIsIndependentOfDXY(t *testing.T) {
	t.Setenv("DXY_FEED_URL", "https://dxy.example.test/data")
	t.Setenv("DXY_FEED_TOKEN", "dxy-secret")
	t.Setenv("PMI_FEED_URL", "https://pmi.example.test/data")
	t.Setenv("PMI_FEED_TOKEN", "pmi-secret")
	t.Setenv("PMI_RELEASE_FILES", "")
	found := false
	for _, p := range ProvidersFromEnv() {
		if v, ok := p.(*providers.PMI); ok {
			found = true
			if v.FeedURL != "https://pmi.example.test/data" || v.FeedToken != "pmi-secret" {
				t.Fatal("PMI configuration not independent")
			}
		}
	}
	if !found {
		t.Fatal("PMI connector missing")
	}
}
