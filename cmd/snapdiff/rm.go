package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/rom8726/snapdiff/internal/storage"
)

func newRmCmd() *cobra.Command {
	var baseDir string

	cmd := &cobra.Command{
		Use:   "rm [snapshot-label]",
		Short: "Remove a snapshot",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			label := args[0]

			store, err := storage.NewStorage(baseDir)
			if err != nil {
				return fmt.Errorf("failed to create storage: %w", err)
			}

			tables, err := store.ListSnapshotTables(label)
			if err != nil {
				return fmt.Errorf("snapshot '%s' not found: %w", label, err)
			}

			if err := store.DeleteSnapshot(label); err != nil {
				return fmt.Errorf("failed to delete snapshot '%s': %w", label, err)
			}

			fmt.Printf("Snapshot '%s' with %d tables deleted successfully.\n", label, len(tables))

			return nil
		},
	}

	cmd.Flags().StringVar(&baseDir, "base-dir", ".snapdiff", "Base directory for snapshots")

	return cmd
}
