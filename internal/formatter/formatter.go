package formatter

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/rom8726/snapdiff/internal/diff"
)

// FormatType represents the output format type
type FormatType string

const (
	FormatCLI      FormatType = "cli"
	FormatYAML     FormatType = "yaml"
	FormatMarkdown FormatType = "markdown"
)

// Options contains configuration for the formatter
type Options struct {
	Format     FormatType
	SortKeys   bool
	Limit      int
	OutputFile string
}

// FormatDiff formats the diff result according to the specified format
func FormatDiff(result *diff.Result, opts Options) error {
	var writer io.Writer
	if opts.OutputFile == "" {
		writer = os.Stdout
	} else {
		file, err := os.Create(opts.OutputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()
		writer = file
	}

	switch opts.Format {
	case FormatCLI:
		return formatCLI(writer, result, opts)
	case FormatYAML:
		return formatYAML(writer, result, opts)
	case FormatMarkdown:
		return formatMarkdown(writer, result, opts)
	default:
		return fmt.Errorf("unsupported format: %s", opts.Format)
	}
}

// formatCLI formats the diff result in CLI format
func formatCLI(w io.Writer, result *diff.Result, opts Options) error {
	tableNames := getSortedTableNames(result)

	for _, tableName := range tableNames {
		tableDiff := result.Tables[tableName]

		_, _ = fmt.Fprintf(w, "📄 Table: %s\n", tableName)

		if len(tableDiff.Inserted) > 0 {
			_, _ = fmt.Fprintf(w, "  ✅  Inserted: %d\n", len(tableDiff.Inserted))

			rows := tableDiff.Inserted
			if opts.Limit > 0 && len(rows) > opts.Limit {
				rows = rows[:opts.Limit]
			}

			for _, row := range rows {
				_, _ = fmt.Fprintf(w, "      + %s\n", formatRow(row, opts.SortKeys))
			}

			if opts.Limit > 0 && len(tableDiff.Inserted) > opts.Limit {
				_, _ = fmt.Fprintf(w, "      ... and %d more\n", len(tableDiff.Inserted)-opts.Limit)
			}
		}

		if len(tableDiff.Updated) > 0 {
			_, _ = fmt.Fprintf(w, "  ✅  Updated: %d\n", len(tableDiff.Updated))

			rows := tableDiff.Updated
			if opts.Limit > 0 && len(rows) > opts.Limit {
				rows = rows[:opts.Limit]
			}

			for _, row := range rows {
				pk := formatRow(row.PrimaryKey, opts.SortKeys)
				changes := formatChanges(row.Before, row.After, opts.SortKeys)
				_, _ = fmt.Fprintf(w, "      ~ %s, %s\n", pk, changes)
			}

			if opts.Limit > 0 && len(tableDiff.Updated) > opts.Limit {
				_, _ = fmt.Fprintf(w, "      ... and %d more\n", len(tableDiff.Updated)-opts.Limit)
			}
		}

		if len(tableDiff.Deleted) > 0 {
			_, _ = fmt.Fprintf(w, "  ❌  Deleted: %d\n", len(tableDiff.Deleted))

			rows := tableDiff.Deleted
			if opts.Limit > 0 && len(rows) > opts.Limit {
				rows = rows[:opts.Limit]
			}

			for _, row := range rows {
				_, _ = fmt.Fprintf(w, "      - %s\n", formatRow(row, opts.SortKeys))
			}

			if opts.Limit > 0 && len(tableDiff.Deleted) > opts.Limit {
				_, _ = fmt.Fprintf(w, "      ... and %d more\n", len(tableDiff.Deleted)-opts.Limit)
			}
		}

		_, _ = fmt.Fprintln(w)
	}

	return nil
}

// formatYAML formats the diff result in YAML format
func formatYAML(w io.Writer, result *diff.Result, opts Options) error {
	// For simplicity, we'll use JSON as a placeholder
	// In a real implementation, you'd use a YAML library

	// Convert to a map for JSON marshaling
	output := make(map[string]any)

	// Get sorted table names
	tableNames := getSortedTableNames(result)

	for _, tableName := range tableNames {
		tableDiff := result.Tables[tableName]

		tableOutput := make(map[string]any)

		// Add inserted rows
		if len(tableDiff.Inserted) > 0 {
			// Apply limit
			rows := tableDiff.Inserted
			if opts.Limit > 0 && len(rows) > opts.Limit {
				rows = rows[:opts.Limit]
			}

			tableOutput["inserted"] = rows
		}

		// Add updated rows
		if len(tableDiff.Updated) > 0 {
			// Apply limit
			rows := tableDiff.Updated
			if opts.Limit > 0 && len(rows) > opts.Limit {
				rows = rows[:opts.Limit]
			}

			// Convert to a format suitable for YAML
			updatedRows := make([]map[string]any, 0, len(rows))
			for _, row := range rows {
				updatedRow := map[string]any{
					"primary_key": row.PrimaryKey,
					"before":      row.Before,
					"after":       row.After,
				}
				updatedRows = append(updatedRows, updatedRow)
			}

			tableOutput["updated"] = updatedRows
		}

		// Add deleted rows
		if len(tableDiff.Deleted) > 0 {
			// Apply limit
			rows := tableDiff.Deleted
			if opts.Limit > 0 && len(rows) > opts.Limit {
				rows = rows[:opts.Limit]
			}

			tableOutput["deleted"] = rows
		}

		output[tableName] = tableOutput
	}

	// Marshal to JSON
	data, err := yaml.Marshal(output)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	// Write to output
	_, err = w.Write(data)

	return err
}

// formatMarkdown formats the diff result in Markdown format
func formatMarkdown(w io.Writer, result *diff.Result, opts Options) error {
	tableNames := getSortedTableNames(result)

	for _, tableName := range tableNames {
		tableDiff := result.Tables[tableName]

		_, _ = fmt.Fprintf(w, "## Table: %s\n\n", tableName)

		if len(tableDiff.Inserted) > 0 {
			_, _ = fmt.Fprintf(w, "### ✅ Inserted (%d rows)\n", len(tableDiff.Inserted))

			rows := tableDiff.Inserted
			if opts.Limit > 0 && len(rows) > opts.Limit {
				rows = rows[:opts.Limit]
			}

			columnNames := getAllColumnNames(rows)

			_, _ = fmt.Fprintf(w, "| %s |\n", strings.Join(columnNames, " | "))
			_, _ = fmt.Fprintf(w, "|%s|\n", strings.Repeat("-----|", len(columnNames)))

			for _, row := range rows {
				var values []string
				for _, col := range columnNames {
					val, ok := row[col]
					if !ok {
						values = append(values, "")
					} else {
						values = append(values, fmt.Sprintf("%v", val))
					}
				}
				_, _ = fmt.Fprintf(w, "| %s |\n", strings.Join(values, " | "))
			}

			if opts.Limit > 0 && len(tableDiff.Inserted) > opts.Limit {
				_, _ = fmt.Fprintf(w, "\n_... and %d more rows_\n", len(tableDiff.Inserted)-opts.Limit)
			}

			_, _ = fmt.Fprintln(w)
		}

		if len(tableDiff.Updated) > 0 {
			_, _ = fmt.Fprintf(w, "### ✅ Updated (%d rows)\n", len(tableDiff.Updated))

			rows := tableDiff.Updated
			if opts.Limit > 0 && len(rows) > opts.Limit {
				rows = rows[:opts.Limit]
			}

			_, _ = fmt.Fprintln(w, "| ID | Field | Before | After |")
			_, _ = fmt.Fprintln(w, "|-----|-------|--------|-------|")

			for _, row := range rows {
				var id string
				if idVal, ok := row.PrimaryKey["id"]; ok {
					id = fmt.Sprintf("%v", idVal)
				} else {
					id = formatRow(row.PrimaryKey, opts.SortKeys)
				}

				changedFields := getChangedFields(row.Before, row.After)

				for i, field := range changedFields {
					before := fmt.Sprintf("%v", row.Before[field])
					after := fmt.Sprintf("%v", row.After[field])

					if i == 0 {
						_, _ = fmt.Fprintf(w, "| %s | %s | %s | %s |\n", id, field, before, after)
					} else {
						_, _ = fmt.Fprintf(w, "| | %s | %s | %s |\n", field, before, after)
					}
				}
			}

			if opts.Limit > 0 && len(tableDiff.Updated) > opts.Limit {
				_, _ = fmt.Fprintf(w, "\n_... and %d more rows_\n", len(tableDiff.Updated)-opts.Limit)
			}

			_, _ = fmt.Fprintln(w)
		}

		if len(tableDiff.Deleted) > 0 {
			_, _ = fmt.Fprintf(w, "### ❌ Deleted (%d rows)\n", len(tableDiff.Deleted))

			rows := tableDiff.Deleted
			if opts.Limit > 0 && len(rows) > opts.Limit {
				rows = rows[:opts.Limit]
			}

			columnNames := getAllColumnNames(rows)

			_, _ = fmt.Fprintf(w, "| %s |\n", strings.Join(columnNames, " | "))
			_, _ = fmt.Fprintf(w, "|%s|\n", strings.Repeat("-----|", len(columnNames)))

			for _, row := range rows {
				var values []string
				for _, col := range columnNames {
					val, ok := row[col]
					if !ok {
						values = append(values, "")
					} else {
						values = append(values, fmt.Sprintf("%v", val))
					}
				}
				_, _ = fmt.Fprintf(w, "| %s |\n", strings.Join(values, " | "))
			}

			if opts.Limit > 0 && len(tableDiff.Deleted) > opts.Limit {
				_, _ = fmt.Fprintf(w, "\n_... and %d more rows_\n", len(tableDiff.Deleted)-opts.Limit)
			}

			_, _ = fmt.Fprintln(w)
		}
	}

	return nil
}

// getSortedTableNames returns a sorted list of table names
func getSortedTableNames(result *diff.Result) []string {
	tableNames := make([]string, 0, len(result.Tables))
	for tableName := range result.Tables {
		tableNames = append(tableNames, tableName)
	}

	sort.Strings(tableNames)

	return tableNames
}

// formatRow formats a row as a string
func formatRow(row map[string]any, sortKeys bool) string {
	var parts []string

	keys := make([]string, 0, len(row))
	for k := range row {
		keys = append(keys, k)
	}

	if sortKeys {
		sort.Strings(keys)
	}

	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s: %v", k, row[k]))
	}

	return strings.Join(parts, ", ")
}

// formatChanges formats the changes between two rows
func formatChanges(before, after map[string]any, sortKeys bool) string {
	var parts []string

	keys := make(map[string]bool)
	for k := range before {
		keys[k] = true
	}
	for k := range after {
		keys[k] = true
	}

	keySlice := make([]string, 0, len(keys))
	for k := range keys {
		keySlice = append(keySlice, k)
	}

	if sortKeys {
		sort.Strings(keySlice)
	}

	for _, k := range keySlice {
		beforeVal, beforeOK := before[k]
		afterVal, afterOK := after[k]

		if !beforeOK {
			parts = append(parts, fmt.Sprintf("%s: null → %v", k, afterVal))
		} else if !afterOK {
			parts = append(parts, fmt.Sprintf("%s: %v → null", k, beforeVal))
		} else if beforeVal != afterVal {
			parts = append(parts, fmt.Sprintf("%s: %v → %v", k, beforeVal, afterVal))
		}
	}

	return strings.Join(parts, ", ")
}

// getAllColumnNames returns all column names from a slice of rows
func getAllColumnNames(rows []map[string]any) []string {
	columnMap := make(map[string]bool)

	for _, row := range rows {
		for col := range row {
			columnMap[col] = true
		}
	}

	columns := make([]string, 0, len(columnMap))
	for col := range columnMap {
		columns = append(columns, col)
	}

	sort.Strings(columns)

	return columns
}

// getChangedFields returns a list of fields that changed between two rows
func getChangedFields(before, after map[string]any) []string {
	var changedFields []string

	for field, beforeVal := range before {
		if afterVal, ok := after[field]; ok {
			if beforeVal != afterVal {
				changedFields = append(changedFields, field)
			}
		} else {
			changedFields = append(changedFields, field)
		}
	}

	for field := range after {
		if _, ok := before[field]; !ok {
			changedFields = append(changedFields, field)
		}
	}

	sort.Strings(changedFields)

	return changedFields
}
