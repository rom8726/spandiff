package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/rom8726/snapdiff/internal/storage"
)

var listBaseDir string

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available snapshots",
		RunE:  runListCmd,
	}

	cmd.Flags().StringVar(&listBaseDir, "base-dir", ".snapdiff", "Base directory for snapshots")

	return cmd
}

func runListCmd(*cobra.Command, []string) error {
	store, err := storage.NewStorage(listBaseDir)
	if err != nil {
		return fmt.Errorf("failed to create storage: %w", err)
	}

	snapshots, err := store.ListSnapshots()
	if err != nil {
		return fmt.Errorf("failed to list snapshots: %w", err)
	}

	if len(snapshots) == 0 {
		fmt.Println("No snapshots found.")

		return nil
	}

	fmt.Println("Available snapshots:")
	for _, snapshot := range snapshots {
		tables, err := store.ListSnapshotTables(snapshot)
		if err != nil {
			return fmt.Errorf("failed to list tables for snapshot %s: %w", snapshot, err)
		}

		fmt.Printf("  📦 %s (%d tables)\n", snapshot, len(tables))
		for _, table := range tables {
			fmt.Printf("    - %s\n", table)
		}

		fmt.Println()
	}

	return nil
}
