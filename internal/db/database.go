// Package db provides database abstraction and connectivity
package db

import (
	"context"
	"errors"
)

// Database is an interface that abstracts database operations
// to allow for different database implementations
type Database interface {
	// Connect establishes a connection to the database
	Connect(ctx context.Context) error

	// Close closes the database connection
	Close() error

	// GetTableNames returns a list of all tables in the database
	GetTableNames(ctx context.Context, schema string) ([]string, error)

	// GetTableColumns returns all columns for a table
	GetTableColumns(ctx context.Context, schema, tableName string) ([]string, error)

	// GetPrimaryKeyColumns returns the primary key columns for a table
	GetPrimaryKeyColumns(ctx context.Context, schema, tableName string) ([]string, error)

	// QueryTableData executes a query to get all data from a table with the specified columns
	QueryTableData(ctx context.Context, schema, tableName string, columns []string) ([]map[string]any, error)
}

// Config contains configuration for database connections
type Config struct {
	// DSN is the data source name (connection string)
	DSN string

	// MaxOpenConns is the maximum number of open connections to the database
	MaxOpenConns int

	// MaxIdleConns is the maximum number of connections in the idle connection pool
	MaxIdleConns int

	// ConnMaxLifetime is the maximum amount of time a connection may be reused (in seconds)
	ConnMaxLifetime int
}

// Error definitions
var (
	// ErrUnsupportedDatabaseType is returned when an unsupported database type is specified
	ErrUnsupportedDatabaseType = errors.New("unsupported database type")
)

// DatabaseType represents the type of database
type DatabaseType string

const (
	// PostgreSQL database type
	PostgreSQL DatabaseType = "postgres"
	// MySQL database type (for future use)
	MySQL DatabaseType = "mysql"
	// SQLite database type (for future use)
	SQLite DatabaseType = "sqlite"
)

// NewDatabase creates a new database instance of the specified type
func NewDatabase(dbType DatabaseType, config Config) (Database, error) {
	switch dbType {
	case PostgreSQL:
		return NewPostgresDB(config), nil
	// Add cases for other database types as they are implemented
	default:
		return nil, ErrUnsupportedDatabaseType
	}
}
