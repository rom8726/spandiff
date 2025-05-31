package snapshot

type Options struct {
	DSN           string   // PostgreSQL connection string
	Label         string   // Snapshot label
	Tables        []string // Specific tables to include
	IgnoreColumns []string // Columns to ignore in snapshot
	SortKeys      bool     // Sort keys in YAML output
	OutputDir     string   // Output base directory (default ".snapdiff")
}
