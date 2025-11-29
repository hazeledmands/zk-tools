# zk-tools

Tools for managing Zettelkasten.

## backup-zettelkasten.sh

Creates a timestamped tar.gz archive of your Zettelkasten directory.

### Usage

```bash
./backup-zettelkasten.sh
```

### Environment Variables

- `ZETTEL_DIR` - Path to your Zettelkasten directory (default: `~/Projects/Zettelkasten`)
- `ZETTEL_BACKUP_DIR` - Path to store backups (default: `~/.backups/zk`)

### Example

```bash
# Use custom paths
export ZETTEL_DIR="$HOME/Documents/notes"
export ZETTEL_BACKUP_DIR="$HOME/backups/notes"
./backup-zettelkasten.sh
```

Backups are named with ISO-8601 timestamps, e.g., `zettelkasten-2025-11-28T14:30:45Z.tar.gz`

## diff-backup.sh

Compares files in a backup tar.gz with the current Zettelkasten directory to see what has changed.

### Usage

```bash
# Compare with most recent backup
./diff-backup.sh

# Compare with specific backup
./diff-backup.sh ~/.backups/zk/Zettelkasten-2025-11-28T23:25:41Z.tar.gz
```

### Environment Variables

- `ZETTEL_DIR` - Path to your Zettelkasten directory (default: `~/Projects/Zettelkasten`)
- `ZETTEL_BACKUP_DIR` - Path to backup directory (default: `~/.backups/zk`)

### Example

```bash
# Use custom paths
export ZETTEL_DIR="$HOME/Documents/notes"
export ZETTEL_BACKUP_DIR="$HOME/backups/notes"
./diff-backup.sh
```

The script will extract the backup to a temporary directory and run a unified diff against your current Zettelkasten directory, showing all changes.

## list-keys

Moves keywords from YAML front-matter to a bulleted list below the first heading, and removes the front-matter entirely.

### Prerequisites

- Go 1.21 or higher

### Installation

```bash
go mod download
go build -o list-keys list-keys.go
```

### Usage

```bash
./list-keys
```

Or specify a custom directory:

```bash
./list-keys /path/to/zettelkasten
```

### Environment Variables

- `ZETTEL_DIR` - Path to your Zettelkasten directory (default: `~/Projects/Zettelkasten`)

### Example

```bash
# Use default directory
./list-keys

# Use custom directory
./list-keys ~/Documents/notes

# Use environment variable
export ZETTEL_DIR="$HOME/Documents/notes"
./list-keys
```

The tool will:
- Scan all `.md` files in the specified directory
- Extract `keywords` from the YAML front-matter
- Insert keywords as a bulleted list immediately below the first heading
- Remove the entire YAML front-matter section
- Preserve keyword values exactly as-is (including `#` characters)
- Preserve all other content

**Warning**: This tool modifies files in place. Make sure to back up your Zettelkasten before running it.

## remove-id-from-headings

Removes timestamp ID prefixes from headings in markdown files.

### Prerequisites

- Go 1.21 or higher

### Installation

```bash
go mod download
go build -o remove-id-from-headings remove-id-from-headings.go
```

### Usage

```bash
./remove-id-from-headings
```

Or specify a custom directory:

```bash
./remove-id-from-headings /path/to/zettelkasten
```

### Environment Variables

- `ZETTEL_DIR` - Path to your Zettelkasten directory (default: `~/Projects/Zettelkasten`)

### Example

```bash
# Use default directory
./remove-id-from-headings

# Use custom directory
./remove-id-from-headings ~/Documents/notes

# Use environment variable
export ZETTEL_DIR="$HOME/Documents/notes"
./remove-id-from-headings
```

The tool will:
- Scan all `.md` files in the specified directory
- Find headings matching the pattern `# 202003092000: Title`
- Remove the ID prefix, leaving just `# Title`
- Work on all heading levels (H1, H2, H3, etc.)
- Preserve all other content

**Warning**: This tool modifies files in place. Make sure to back up your Zettelkasten before running it.
