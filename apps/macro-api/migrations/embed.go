package migrations

import _ "embed"

//go:embed 001_macro.sql
var Schema string

//go:embed 002_operations.sql
var Operations string

//go:embed 003_ingestion_pilot.sql
var IngestionPilot string

//go:embed 004_cpi_reports.sql
var CPIReports string

//go:embed 005_te_pmi.sql
var TradingEconomicsPMI string

//go:embed 006_source_selection.sql
var SourceSelection string

//go:embed 007_pmi_history.sql
var PMIHistory string
