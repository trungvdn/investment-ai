package service

import (
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/providers"
	"os"
)

func ProvidersFromEnv() []providers.Provider {
	key := os.Getenv("FRED_API_KEY")
	return []providers.Provider{
		&providers.FRED{Code: "US_EFFR", Series: "DFF", APIKey: key},
		&providers.FRED{Code: "US_10Y", Series: "DGS10", APIKey: key},
		&providers.FRED{Code: "US_BROAD_USD", Series: "DTWEXBGS", APIKey: key},
		&providers.SDMX{Code: "VN_CPI_YOY", File: "CPIVNM", Series: "PCPI_IX"},
		&providers.SDMX{Code: "VN_IIP_YOY", File: "IIPVNM", Series: "AIP_ISIC4_IX"},
		&providers.SDMX{Code: "VN_EXPORT_YOY", File: "METVNM", Series: "TXG_FOB_USD"},
		&providers.Reports{Kind: "economic", URL: os.Getenv("NSO_ECONOMIC_FEED_URL")},
		&providers.Reports{Kind: "prices", URL: os.Getenv("NSO_PRICES_FEED_URL")},
	}
}
