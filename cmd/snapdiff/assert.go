package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/rom8726/snapdiff/internal/diff"
)

var assertOpts diff.Options
var expectedFile string

func newAssertCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assert",
		Short: "Assert that diff matches expected changes",
		RunE:  runAssertCmd,
	}

	cmd.Flags().StringVar(&assertOpts.From, "from", "", "Source snapshot label (required)")
	cmd.Flags().StringVar(&assertOpts.To, "to", "", "Target snapshot label (required)")
	cmd.Flags().StringVar(&expectedFile, "expected", "", "Expected changes file (required)")
	cmd.Flags().StringSliceVar(&assertOpts.Tables, "table", nil, "Filter by tables (comma-separated)")
	cmd.Flags().StringSliceVar(&assertOpts.IgnoreColumns, "ignore-columns", nil, "Columns to ignore")
	cmd.Flags().BoolVar(&assertOpts.OnlyChanged, "only-changed", false, "Show only changed tables")
	cmd.Flags().StringVar(&assertOpts.BaseDir, "base-dir", ".snapdiff", "Base directory for snapshots")

	return cmd
}

func runAssertCmd(cmd *cobra.Command, _ []string) error {
	if assertOpts.From == "" || assertOpts.To == "" {
		return fmt.Errorf("both --from and --to are required")
	}

	if expectedFile == "" {
		return fmt.Errorf("--expected is required")
	}

	result, err := diff.Run(cmd.Context(), assertOpts)
	if err != nil {
		return fmt.Errorf("failed to run diff: %w", err)
	}

	expectedData, err := os.ReadFile(expectedFile)
	if err != nil {
		return fmt.Errorf("failed to read expected file: %w", err)
	}

	var expected map[string]any
	if err := json.Unmarshal(expectedData, &expected); err != nil {
		return fmt.Errorf("failed to parse expected file: %w", err)
	}

	resultMap := make(map[string]any)
	for tableName, tableDiff := range result.Tables {
		tableMap := make(map[string]any)

		if len(tableDiff.Inserted) > 0 {
			tableMap["inserted"] = tableDiff.Inserted
		}

		if len(tableDiff.Updated) > 0 {
			updatedRows := make([]map[string]any, 0, len(tableDiff.Updated))
			for _, row := range tableDiff.Updated {
				updatedRow := map[string]any{
					"primary_key": row.PrimaryKey,
					"before":      row.Before,
					"after":       row.After,
				}
				updatedRows = append(updatedRows, updatedRow)
			}
			tableMap["updated"] = updatedRows
		}

		if len(tableDiff.Deleted) > 0 {
			tableMap["deleted"] = tableDiff.Deleted
		}

		if len(tableMap) > 0 {
			resultMap[tableName] = tableMap
		}
	}

	if !compareJSON(expected, resultMap) {
		resultJSON, _ := json.MarshalIndent(resultMap, "", "  ")

		return fmt.Errorf("diff does not match expected changes:\nExpected:\n%s\n\nActual:\n%s", expectedData, resultJSON)
	}

	fmt.Println("✅ Diff matches expected changes.")

	return nil
}

// compareJSON compares two JSON objects for equality
func compareJSON(expected, actual any) bool {
	expectedJSON, err := json.Marshal(expected)
	if err != nil {
		return false
	}

	actualJSON, err := json.Marshal(actual)
	if err != nil {
		return false
	}

	return string(expectedJSON) == string(actualJSON)
}
