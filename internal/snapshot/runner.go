package snapshot

import (
	"context"
	"fmt"
	"log"

	"github.com/rom8726/snapdiff/internal/db"
	"github.com/rom8726/snapdiff/internal/storage"
)

// TableData represents the data for a single table
type TableData []map[string]any

// Run executes the snapshot command
func Run(ctx context.Context, opts Options) error {
	// Create database configuration
	dbConfig := db.Config{
		DSN:             opts.DSN,
		MaxOpenConns:    10, // Default values for connection pooling
		MaxIdleConns:    5,
		ConnMaxLifetime: 300, // 5 minutes
	}

	// Create database instance
	database, err := db.NewDatabase(db.PostgreSQL, dbConfig)
	if err != nil {
		return fmt.Errorf("failed to create database instance: %w", err)
	}

	// Connect to the database
	if err := database.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer database.Close()

	store, err := storage.NewStorage(opts.OutputDir)
	if err != nil {
		return fmt.Errorf("failed to create storage: %w", err)
	}

	// Get schema (default to public)
	schema := ""

	var tables []string
	if len(opts.Tables) > 0 {
		tables = opts.Tables
	} else {
		allTables, err := database.GetTableNames(ctx, schema)
		if err != nil {
			return fmt.Errorf("failed to get table names: %w", err)
		}
		tables = allTables
	}

	ignoreColumnsMap := make(map[string]bool)
	for _, col := range opts.IgnoreColumns {
		ignoreColumnsMap[col] = true
	}

	for _, tableName := range tables {
		log.Printf("Processing table: %s", tableName)

		columns, err := database.GetTableColumns(ctx, schema, tableName)
		if err != nil {
			return fmt.Errorf("failed to get columns for table %s: %w", tableName, err)
		}

		var filteredColumns []string
		for _, col := range columns {
			if !ignoreColumnsMap[col] {
				filteredColumns = append(filteredColumns, col)
			}
		}

		if len(filteredColumns) == 0 {
			log.Printf("Skipping table %s: all columns are ignored", tableName)
			continue
		}

		// Use the new QueryTableData method instead of building and executing the query directly
		tableData, err := database.QueryTableData(ctx, schema, tableName, filteredColumns)
		if err != nil {
			return fmt.Errorf("failed to query table %s: %w", tableName, err)
		}

		format := storage.FormatJSON
		if err := store.SaveSnapshot(opts.Label, tableName, tableData, format); err != nil {
			return fmt.Errorf("failed to save snapshot for table %s: %w", tableName, err)
		}

		log.Printf("Saved snapshot for table %s with %d rows", tableName, len(tableData))
	}

	log.Printf("Snapshot '%s' created successfully", opts.Label)

	return nil
}
