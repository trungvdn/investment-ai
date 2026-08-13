package macrodata

import (
    "errors"
    "time"
)

type MacroCategory string
type MacroFrequency string
type QualityStatus string

const (
    CategoryLiquidity MacroCategory = "LIQUIDITY"
    CategoryCredit    MacroCategory = "CREDIT"
    CategoryExternal  MacroCategory = "EXTERNAL"

    FrequencyMonthly MacroFrequency = "MONTHLY"

    QualityValid   QualityStatus = "VALID"
    QualityWarning QualityStatus = "WARNING"
    QualityInvalid QualityStatus = "INVALID"
)

type MacroIndicator struct {
    Code      string         `json:"code"`
    Name      string         `json:"name"`
    Category  MacroCategory  `json:"category"`
    Unit      string         `json:"unit"`
    Frequency MacroFrequency `json:"frequency"`
    Country   string         `json:"country"`
    Active    bool           `json:"active"`
}

type MacroObservation struct {
    IndicatorCode  string        `json:"indicatorCode"`
    Value          float64       `json:"value"`
    PreviousValue  *float64      `json:"previousValue,omitempty"`
    ObservationDate time.Time    `json:"observationDate"`
    Unit           string        `json:"unit"`
    Source         string        `json:"source"`
    CollectedAt    time.Time     `json:"collectedAt"`
    QualityStatus  QualityStatus `json:"qualityStatus"`
}

var (
    ErrIndicatorNotFound = errors.New("indicator not found")
    ErrInvalidValue      = errors.New("invalid value")
    ErrInvalidDate       = errors.New("invalid date")
    ErrInvalidUnit       = errors.New("invalid unit")
    ErrInvalidSource     = errors.New("invalid source")
)

func ValidateObservation(ind MacroObservation) error {
    // indicator code
    if ind.IndicatorCode == "" {
        return ErrIndicatorNotFound
    }
    // value finite
    if !(ind.Value == ind.Value) { // NaN check
        return ErrInvalidValue
    }
    // negative checks for certain indicators
    switch ind.IndicatorCode {
    case "VN_INTERBANK_ON", "VN_SBV_TBILL", "VN_USD_VND":
        if ind.Value < 0 {
            return ErrInvalidValue
        }
    }
    // dates
    if ind.ObservationDate.IsZero() {
        return ErrInvalidDate
    }
    if ind.CollectedAt.IsZero() {
        return ErrInvalidDate
    }
    if ind.CollectedAt.Before(ind.ObservationDate) {
        return errors.New("collectedAt before observationDate")
    }
    if ind.Unit == "" {
        return ErrInvalidUnit
    }
    if ind.Source == "" {
        return ErrInvalidSource
    }
    return nil
}
