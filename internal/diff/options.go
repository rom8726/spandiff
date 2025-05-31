package diff

// Options contains configuration for the diff command
type Options struct {
	From          string   // Source snapshot label
	To            string   // Target snapshot label
	Tables        []string // Specific tables to include
	IgnoreColumns []string // Columns to ignore in comparison
	OnlyChanged   bool     // Show only changed tables
	Format        string   // Output format (cli, yaml, markdown)
	OutputFile    string   // Output file path (stdout if empty)
	SortKeys      bool     // Sort keys in output
	Limit         int      // Limit the number of rows in output
	BaseDir       string   // Base directory for snapshots
}
