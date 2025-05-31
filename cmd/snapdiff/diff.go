package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/rom8726/snapdiff/internal/diff"
	"github.com/rom8726/snapdiff/internal/formatter"
)

var diffOpts diff.Options
var formatOpts formatter.Options
var formatStr string

func newDiffCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Compare two database snapshots",
		RunE:  runDiffCmd,
	}

	// Add flags
	cmd.Flags().StringVar(&diffOpts.From, "from", "", "Source snapshot label (required)")
	cmd.Flags().StringVar(&diffOpts.To, "to", "", "Target snapshot label (required)")
	cmd.Flags().StringSliceVar(&diffOpts.Tables, "table", nil, "Filter by tables (comma-separated)")
	cmd.Flags().StringSliceVar(&diffOpts.IgnoreColumns, "ignore-columns", nil, "Columns to ignore")
	cmd.Flags().BoolVar(&diffOpts.OnlyChanged, "only-changed", false, "Show only changed tables")
	cmd.Flags().StringVar(&formatStr, "format", "cli", "Output format (cli, yaml, markdown)")
	cmd.Flags().StringVar(&formatOpts.OutputFile, "out", "", "Output file (stdout if not specified)")
	cmd.Flags().BoolVar(&formatOpts.SortKeys, "sort-keys", false, "Sort keys in output")
	cmd.Flags().IntVar(&formatOpts.Limit, "limit", 0, "Limit the number of rows in output")
	cmd.Flags().StringVar(&diffOpts.BaseDir, "base-dir", ".snapdiff", "Base directory for snapshots")

	return cmd
}

func runDiffCmd(cmd *cobra.Command, _ []string) error {
	if diffOpts.From == "" || diffOpts.To == "" {
		return fmt.Errorf("both --from and --to are required")
	}

	switch formatStr {
	case "cli":
		formatOpts.Format = formatter.FormatCLI
	case "yaml":
		formatOpts.Format = formatter.FormatYAML
	case "markdown":
		formatOpts.Format = formatter.FormatMarkdown
	default:
		return fmt.Errorf("unsupported format: %s", formatStr)
	}

	result, err := diff.Run(cmd.Context(), diffOpts)
	if err != nil {
		return fmt.Errorf("failed to run diff: %w", err)
	}

	if err := formatter.FormatDiff(result, formatOpts); err != nil {
		return fmt.Errorf("failed to format diff: %w", err)
	}

	return nil
}
