package duckdb

import "database/sql"

type DuckdbCache struct {
	Database *sql.DB
}
