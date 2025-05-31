package db

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// Connect establishes a connection to the PostgreSQL database
func Connect(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()

		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// GetTableNames returns a list of all tables in the database
func GetTableNames(ctx context.Context, db *sql.DB, schema string) ([]string, error) {
	if schema == "" {
		schema = "public"
	}

	query := `
SELECT table_name 
FROM information_schema.tables 
WHERE table_schema = $1 
AND table_type = 'BASE TABLE'
ORDER BY table_name`

	rows, err := db.QueryContext(ctx, query, schema)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}

		tables = append(tables, tableName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating table rows: %w", err)
	}

	return tables, nil
}

// GetPrimaryKeyColumns returns the primary key columns for a table
func GetPrimaryKeyColumns(ctx context.Context, db *sql.DB, schema, tableName string) ([]string, error) {
	if schema == "" {
		schema = "public"
	}

	query := `
SELECT a.attname
FROM pg_index i
JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
WHERE i.indrelid = ($1 || '.' || $2)::regclass
AND i.indisprimary
ORDER BY a.attnum`

	rows, err := db.QueryContext(ctx, query, schema, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query primary key for %s.%s: %w", schema, tableName, err)
	}
	defer rows.Close()

	var pkColumns []string
	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			return nil, fmt.Errorf("failed to scan primary key column: %w", err)
		}

		pkColumns = append(pkColumns, columnName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating primary key rows: %w", err)
	}

	return pkColumns, nil
}

// GetTableColumns returns all columns for a table
func GetTableColumns(ctx context.Context, db *sql.DB, schema, tableName string) ([]string, error) {
	if schema == "" {
		schema = "public"
	}

	query := `
SELECT column_name
FROM information_schema.columns
WHERE table_schema = $1
AND table_name = $2
ORDER BY ordinal_position`

	rows, err := db.QueryContext(ctx, query, schema, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns for %s.%s: %w", schema, tableName, err)
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			return nil, fmt.Errorf("failed to scan column name: %w", err)
		}

		columns = append(columns, columnName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating column rows: %w", err)
	}

	return columns, nil
}
