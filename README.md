# snapdiff

A CLI tool for creating and comparing snapshots of PostgreSQL database tables. It allows you to see changes between two points in time: inserted, updated, and deleted rows.

## Features

- Create snapshots of PostgreSQL tables (in YAML/JSON)
- Compare snapshots to identify inserted, updated, and deleted rows
- Output in CLI, YAML, or Markdown formats
- Filter tables and ignore columns
- Assert functionality for CI and snapshot testing
- Local storage of snapshots

## Installation

```bash
go install github.com/rom8726/snapdiff/cmd/snapdiff@latest
```

## Usage

### Create a snapshot

```bash
snapdiff snapshot --dsn "postgresql://user:password@localhost:5432/dbname" --label pre
```

### Make changes to your database

Run your migrations, tests, or other operations that modify the database.

### Create another snapshot

```bash
snapdiff snapshot --dsn "postgresql://user:password@localhost:5432/dbname" --label post
```

### Compare snapshots

```bash
snapdiff diff --from pre --to post
```

### List available snapshots

```bash
snapdiff list
```

### Remove a snapshot

```bash
snapdiff rm pre
```

### Assert changes match expectations (for CI)

```bash
snapdiff assert --from pre --to post --expected expected-changes.json
```

## Command Options

### Global Options

- `--base-dir`: Base directory for snapshots (default: `.snapdiff`)

### Snapshot Options

- `--dsn`: PostgreSQL connection string (required)
- `--label`: Snapshot label (required)
- `--table`: Filter by tables (comma-separated)
- `--ignore-columns`: Columns to ignore (comma-separated)
- `--sort-keys`: Sort keys in YAML output

### Diff Options

- `--from`: Source snapshot label (required)
- `--to`: Target snapshot label (required)
- `--table`: Filter by tables (comma-separated)
- `--ignore-columns`: Columns to ignore (comma-separated)
- `--only-changed`: Show only changed tables
- `--format`: Output format (`cli`, `yaml`, `markdown`)
- `--out`: Output file (stdout if not specified)
- `--sort-keys`: Sort keys in output
- `--limit`: Limit the number of rows in output

### Assert Options

- `--from`: Source snapshot label (required)
- `--to`: Target snapshot label (required)
- `--expected`: Expected changes file (required)
- `--table`: Filter by tables (comma-separated)
- `--ignore-columns`: Columns to ignore (comma-separated)
- `--only-changed`: Show only changed tables

## Example Workflow

1. Before running a migration:
   ```bash
   snapdiff snapshot --dsn "postgresql://user:password@localhost:5432/dbname" --label pre
   ```

2. Run your migration:
   ```bash
   migrate -database "postgresql://user:password@localhost:5432/dbname" -path ./migrations up
   ```

3. After the migration:
   ```bash
   snapdiff snapshot --dsn "postgresql://user:password@localhost:5432/dbname" --label post
   ```

4. Compare the changes:
   ```bash
   snapdiff diff --from pre --to post
   ```

## Use Cases

- **Migration Validation**: Verify that migrations produce the expected changes
- **Snapshot Testing**: Create snapshot-based tests for your database operations
- **Debugging**: Identify unexpected side effects of database operations
- **CI/CD**: Validate database changes in your CI/CD pipeline

## License

Apache-2.0
