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

## list-keys

Rewrites the YAML front-matter in all markdown files to keep only the `keywords` field, removing all other fields (ID, title, date, etc.).

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
- Rewrite each file's front-matter to keep only the `keywords` field
- Preserve the `keywords` values exactly as-is (including `#` characters)
- Remove all other front-matter fields (ID, title, date, etc.)
- Preserve all content after the front-matter

**Warning**: This tool modifies files in place. Make sure to back up your Zettelkasten before running it.
