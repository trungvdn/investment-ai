package migrations

import _ "embed"

//go:embed 001_macro.sql
var Schema string

//go:embed 002_operations.sql
var Operations string
