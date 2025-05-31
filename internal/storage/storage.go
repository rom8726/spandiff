package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// SnapshotFormat represents the format of the snapshot file
type SnapshotFormat string

const (
	FormatYAML SnapshotFormat = "yaml"
	FormatJSON SnapshotFormat = "json"
)

// Storage handles saving and loading snapshots
type Storage struct {
	baseDir string
}

// NewStorage creates a new storage instance
func NewStorage(baseDir string) (*Storage, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &Storage{
		baseDir: baseDir,
	}, nil
}

// SaveSnapshot saves a snapshot of a table to disk
func (s *Storage) SaveSnapshot(label, tableName string, data any, format SnapshotFormat) error {
	snapshotDir := filepath.Join(s.baseDir, "snapshots", label)
	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		return fmt.Errorf("failed to create snapshot directory: %w", err)
	}

	ext := string(format)

	filePath := filepath.Join(snapshotDir, fmt.Sprintf("%s.%s", tableName, ext))

	var content []byte
	var err error

	switch format {
	case FormatJSON:
		content, err = json.MarshalIndent(data, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
	case FormatYAML:
		content, err = yaml.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal YAML: %w", err)
		}
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return fmt.Errorf("failed to write snapshot file: %w", err)
	}

	return nil
}

// LoadSnapshot loads a snapshot from disk
func (s *Storage) LoadSnapshot(label, tableName string, format SnapshotFormat) (any, error) {
	ext := string(format)

	filePath := filepath.Join(s.baseDir, "snapshots", label, fmt.Sprintf("%s.%s", tableName, ext))

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("snapshot file does not exist: %s", filePath)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read snapshot file: %w", err)
	}

	var data any

	switch format {
	case FormatJSON:
		if err := json.Unmarshal(content, &data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
		}
	case FormatYAML:
		if err := yaml.Unmarshal(content, &data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	return data, nil
}

// ListSnapshots returns a list of all available snapshots
func (s *Storage) ListSnapshots() ([]string, error) {
	snapshotsDir := filepath.Join(s.baseDir, "snapshots")

	if err := os.MkdirAll(snapshotsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create snapshots directory: %w", err)
	}

	entries, err := os.ReadDir(snapshotsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read snapshots directory: %w", err)
	}

	var snapshots []string
	for _, entry := range entries {
		if entry.IsDir() {
			snapshots = append(snapshots, entry.Name())
		}
	}

	return snapshots, nil
}

// ListSnapshotTables returns a list of all tables in a snapshot
func (s *Storage) ListSnapshotTables(label string) ([]string, error) {
	snapshotDir := filepath.Join(s.baseDir, "snapshots", label)

	if _, err := os.Stat(snapshotDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("snapshot does not exist: %s", label)
	}

	entries, err := os.ReadDir(snapshotDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read snapshot directory: %w", err)
	}

	var tables []string
	for _, entry := range entries {
		if !entry.IsDir() {
			name := entry.Name()
			tableName := strings.TrimSuffix(name, filepath.Ext(name))
			tables = append(tables, tableName)
		}
	}

	return tables, nil
}

// DeleteSnapshot deletes a snapshot
func (s *Storage) DeleteSnapshot(label string) error {
	snapshotDir := filepath.Join(s.baseDir, "snapshots", label)

	if _, err := os.Stat(snapshotDir); os.IsNotExist(err) {
		return fmt.Errorf("snapshot does not exist: %s", label)
	}

	if err := os.RemoveAll(snapshotDir); err != nil {
		return fmt.Errorf("failed to delete snapshot: %w", err)
	}

	return nil
}
