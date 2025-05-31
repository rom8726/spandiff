package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/rom8726/snapdiff/internal/snapshot"
)

func newSnapshotCmd() *cobra.Command {
	var opts snapshot.Options

	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Create a database snapshot",
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.DSN == "" || opts.Label == "" {
				return fmt.Errorf("both --dsn and --label are required")
			}

			return snapshot.Run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.DSN, "dsn", "", "PostgreSQL DSN (required)")
	cmd.Flags().StringVar(&opts.Label, "label", "", "Snapshot label (required)")
	cmd.Flags().StringSliceVar(&opts.Tables, "table", nil, "Filter by tables (comma-separated)")
	cmd.Flags().StringSliceVar(&opts.IgnoreColumns, "ignore-columns", nil, "Columns to ignore")
	cmd.Flags().BoolVar(&opts.SortKeys, "sort-keys", false, "Sort keys in YAML output")
	cmd.Flags().StringVar(&opts.OutputDir, "output-dir", ".snapdiff", "Snapshot output directory")

	return cmd
}
