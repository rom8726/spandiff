package main

import (
	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapdiff",
		Short: "Snapshot-based PostgreSQL diff tool",
		Long: `snapdiff is a CLI tool for creating and comparing snapshots of PostgreSQL database tables.
It allows you to see changes between two points in time: inserted, updated, and deleted rows.

Useful for:
- Migration validation
- Snapshot-based testing
- Debugging side effects
- CI/CD change analysis`,
	}

	// Add commands
	cmd.AddCommand(newSnapshotCmd())
	cmd.AddCommand(newDiffCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newRmCmd())
	cmd.AddCommand(newAssertCmd())

	return cmd
}
