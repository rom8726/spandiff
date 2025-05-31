// Package db provides database abstraction and connectivity
package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// PostgresDB implements the Database interface for PostgreSQL
type PostgresDB struct {
	config Config
	db     *sql.DB
}

// NewPostgresDB creates a new PostgreSQL database instance
func NewPostgresDB(config Config) *PostgresDB {
	return &PostgresDB{
		config: config,
	}
}

// Connect establishes a connection to the PostgreSQL database
func (p *PostgresDB) Connect(ctx context.Context) error {
	db, err := sql.Open("postgres", p.config.DSN)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pooling
	if p.config.MaxOpenConns > 0 {
		db.SetMaxOpenConns(p.config.MaxOpenConns)
	}
	if p.config.MaxIdleConns > 0 {
		db.SetMaxIdleConns(p.config.MaxIdleConns)
	}
	if p.config.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(time.Duration(p.config.ConnMaxLifetime) * time.Second)
	}

	// Test the connection
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	p.db = db
	return nil
}

// Close closes the database connection
func (p *PostgresDB) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// GetTableNames returns a list of all tables in the database
func (p *PostgresDB) GetTableNames(ctx context.Context, schema string) ([]string, error) {
	if schema == "" {
		schema = "public"
	}

	query := `
SELECT table_name 
FROM information_schema.tables 
WHERE table_schema = $1 
AND table_type = 'BASE TABLE'
ORDER BY table_name`

	rows, err := p.db.QueryContext(ctx, query, schema)
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
func (p *PostgresDB) GetPrimaryKeyColumns(ctx context.Context, schema, tableName string) ([]string, error) {
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

	rows, err := p.db.QueryContext(ctx, query, schema, tableName)
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
func (p *PostgresDB) GetTableColumns(ctx context.Context, schema, tableName string) ([]string, error) {
	if schema == "" {
		schema = "public"
	}

	query := `
SELECT column_name
FROM information_schema.columns
WHERE table_schema = $1
AND table_name = $2
ORDER BY ordinal_position`

	rows, err := p.db.QueryContext(ctx, query, schema, tableName)
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

// QueryTableData executes a query to get all data from a table with the specified columns
func (p *PostgresDB) QueryTableData(ctx context.Context, schema, tableName string, columns []string) ([]map[string]any, error) {
	if schema == "" {
		schema = "public"
	}

	// Build query with proper escaping
	var queryBuilder strings.Builder
	queryBuilder.WriteString("SELECT ")

	for i, col := range columns {
		if i > 0 {
			queryBuilder.WriteString(", ")
		}
		// Quote column names to prevent SQL injection
		queryBuilder.WriteString(fmt.Sprintf("\"%s\"", col))
	}

	// Quote schema and table names
	queryBuilder.WriteString(fmt.Sprintf(" FROM \"%s\".\"%s\"", schema, tableName))

	query := queryBuilder.String()
	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query table %s.%s: %w", schema, tableName, err)
	}
	defer rows.Close()

	var tableData []map[string]any

	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))

		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row for table %s.%s: %w", schema, tableName, err)
		}

		rowMap := make(map[string]any)
		for i, col := range columns {
			rowMap[col] = values[i]
		}

		tableData = append(tableData, rowMap)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows for table %s.%s: %w", schema, tableName, err)
	}

	return tableData, nil
}
