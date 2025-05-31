package diff

import (
	"context"
	"fmt"
	"reflect"
	"sort"

	"github.com/rom8726/snapdiff/internal/storage"
)

// TableDiff represents the differences between two snapshots of a table
type TableDiff struct {
	TableName string
	Inserted  []map[string]any
	Updated   []UpdatedRow
	Deleted   []map[string]any
}

// UpdatedRow represents a row that was updated
type UpdatedRow struct {
	PrimaryKey map[string]any
	Before     map[string]any
	After      map[string]any
}

// Result contains all table diffs
type Result struct {
	Tables map[string]*TableDiff
}

// Run executes the diff command
func Run(_ context.Context, opts Options) (*Result, error) {
	store, err := storage.NewStorage(opts.BaseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	var tables []string
	if len(opts.Tables) > 0 {
		tables = opts.Tables
	} else {
		fromTables, err := store.ListSnapshotTables(opts.From)
		if err != nil {
			return nil, fmt.Errorf("failed to list tables in 'from' snapshot: %w", err)
		}

		toTables, err := store.ListSnapshotTables(opts.To)
		if err != nil {
			return nil, fmt.Errorf("failed to list tables in 'to' snapshot: %w", err)
		}

		tableMap := make(map[string]bool)
		for _, t := range fromTables {
			tableMap[t] = true
		}
		for _, t := range toTables {
			tableMap[t] = true
		}

		for t := range tableMap {
			tables = append(tables, t)
		}

		sort.Strings(tables)
	}

	ignoreColumnsMap := make(map[string]bool)
	for _, col := range opts.IgnoreColumns {
		ignoreColumnsMap[col] = true
	}

	result := &Result{
		Tables: make(map[string]*TableDiff),
	}

	for _, tableName := range tables {
		fromData, err := store.LoadSnapshot(opts.From, tableName, storage.FormatJSON)
		if err != nil {
			// If table doesn't exist in 'from', all rows are inserted
			fromData = []any{}
		}

		toData, err := store.LoadSnapshot(opts.To, tableName, storage.FormatJSON)
		if err != nil {
			// If table doesn't exist in 'to', all rows are deleted
			toData = []any{}
		}

		fromRows := convertToMapSlice(fromData)
		toRows := convertToMapSlice(toData)

		tableDiff, err := compareRows(tableName, fromRows, toRows, ignoreColumnsMap)
		if err != nil {
			return nil, fmt.Errorf("failed to compare rows for table %s: %w", tableName, err)
		}

		if opts.OnlyChanged && len(tableDiff.Inserted) == 0 && len(tableDiff.Updated) == 0 && len(tableDiff.Deleted) == 0 {
			continue
		}

		result.Tables[tableName] = tableDiff
	}

	return result, nil
}

// convertToMapSlice converts an any to []map[string]any
func convertToMapSlice(data any) []map[string]any {
	if data == nil {
		return []map[string]any{}
	}

	if rows, ok := data.([]map[string]any); ok {
		return rows
	}

	if rows, ok := data.([]any); ok {
		result := make([]map[string]any, 0, len(rows))
		for _, row := range rows {
			if rowMap, ok := row.(map[string]any); ok {
				result = append(result, rowMap)
			}
		}
		return result
	}

	return []map[string]any{}
}

// compareRows compares two sets of rows and returns the differences
func compareRows(tableName string, fromRows, toRows []map[string]any, ignoreColumns map[string]bool) (*TableDiff, error) {
	result := &TableDiff{
		TableName: tableName,
	}

	fromMap := make(map[string]map[string]any)
	toMap := make(map[string]map[string]any)

	for _, row := range fromRows {
		key := generateRowKey(row)
		fromMap[key] = row
	}

	for _, row := range toRows {
		key := generateRowKey(row)
		toMap[key] = row
	}

	for key, row := range toMap {
		if _, exists := fromMap[key]; !exists {
			result.Inserted = append(result.Inserted, row)
		}
	}

	for key, row := range fromMap {
		if _, exists := toMap[key]; !exists {
			result.Deleted = append(result.Deleted, row)
		}
	}

	for key, fromRow := range fromMap {
		if toRow, exists := toMap[key]; exists {
			if !rowsEqual(fromRow, toRow, ignoreColumns) {
				updatedRow := UpdatedRow{
					PrimaryKey: extractPrimaryKey(fromRow),
					Before:     filterIgnoredColumns(fromRow, ignoreColumns),
					After:      filterIgnoredColumns(toRow, ignoreColumns),
				}
				result.Updated = append(result.Updated, updatedRow)
			}
		}
	}

	return result, nil
}

// generateRowKey generates a unique key for a row based on its values
func generateRowKey(row map[string]any) string {
	// For simplicity, we'll use the "id" column if it exists
	if id, ok := row["id"]; ok {
		return fmt.Sprintf("%v", id)
	}

	var key string
	keys := make([]string, 0, len(row))
	for k := range row {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		key += fmt.Sprintf("%s:%v;", k, row[k])
	}

	return key
}

// rowsEqual checks if two rows are equal, ignoring specified columns
func rowsEqual(row1, row2 map[string]any, ignoreColumns map[string]bool) bool {
	for key, val1 := range row1 {
		if ignoreColumns[key] {
			continue
		}

		val2, exists := row2[key]
		if !exists {
			return false
		}

		if !reflect.DeepEqual(val1, val2) {
			return false
		}
	}

	// Check if row2 has extra fields
	for key := range row2 {
		if ignoreColumns[key] {
			continue
		}

		if _, exists := row1[key]; !exists {
			return false
		}
	}

	return true
}

// extractPrimaryKey extracts the primary key from a row
func extractPrimaryKey(row map[string]any) map[string]any {
	result := make(map[string]any)

	// For simplicity, we'll use the "id" column if it exists
	if id, ok := row["id"]; ok {
		result["id"] = id
		return result
	}

	return row
}

// filterIgnoredColumns returns a copy of the row with ignored columns removed
func filterIgnoredColumns(row map[string]any, ignoreColumns map[string]bool) map[string]any {
	result := make(map[string]any)
	for key, value := range row {
		if !ignoreColumns[key] {
			result[key] = value
		}
	}

	return result
}
