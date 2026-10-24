package service

import (
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/providers"
	"os"
	"strings"
)

func ProvidersFromEnv() []providers.Provider {
	key := os.Getenv("FRED_API_KEY")
	return []providers.Provider{
		&providers.DXY{URL: os.Getenv("DXY_FEED_URL"), Token: os.Getenv("DXY_FEED_TOKEN")},
		&providers.FRED{Code: "US_EFFR", Series: "DFF", APIKey: key},
		&providers.FRED{Code: "US_10Y", Series: "DGS10", APIKey: key},
		&providers.FRED{Code: "US_BROAD_USD", Series: "DTWEXBGS", APIKey: key},
		&providers.SDMX{Code: "VN_CPI_YOY", File: "CPIVNM", Series: "PCPI_IX"},
		&providers.SDMX{Code: "VN_IIP_YOY", File: "IIPVNM", Series: "AIP_ISIC4_IX"},
		&providers.SDMX{Code: "VN_EXPORT_YOY", File: "METVNM", Series: "TXG_FOB_USD"},
		&providers.SBVPage{Kind: "fx", URL: os.Getenv("SBV_FX_URL")},
		&providers.SBVPage{Kind: "interbank", URL: os.Getenv("SBV_INTERBANK_URL")},
		&providers.SBVStatistics{Kind: "m2", Endpoint: os.Getenv("SBV_ARTICLES_URL")},
		&providers.SBVStatistics{Kind: "credit", Endpoint: os.Getenv("SBV_ARTICLES_URL")},
		&providers.PPI{},
		&providers.PMI{IndexURL: os.Getenv("PMI_RELEASE_INDEX_URL"), ReleaseURLs: strings.FieldsFunc(os.Getenv("PMI_RELEASE_URLS"), func(r rune) bool { return r == ',' || r == ' ' || r == '\n' })},
		&providers.SBVBills{URL: os.Getenv("SBV_BILLS_URL")},
		&providers.Reports{Kind: "economic", URL: os.Getenv("NSO_ECONOMIC_FEED_URL")},
		&providers.Reports{Kind: "prices", URL: os.Getenv("NSO_PRICES_FEED_URL")},
	}
}
