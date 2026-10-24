package migrations

import _ "embed"

//go:embed 001_macro.sql
var Schema string
