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

## verify-ids

Verifies that the ID in the YAML front-matter matches the ID in the filename for all markdown files in your Zettelkasten.

### Prerequisites

- Go 1.21 or higher

### Installation

```bash
go mod download
go build -o verify-ids verify-ids.go
```

### Usage

```bash
./verify-ids
```

Or specify a custom directory:

```bash
./verify-ids /path/to/zettelkasten
```

### Environment Variables

- `ZETTEL_DIR` - Path to your Zettelkasten directory (default: `~/Projects/Zettelkasten`)

### Example

```bash
# Use default directory
./verify-ids

# Use custom directory
./verify-ids ~/Documents/notes

# Use environment variable
export ZETTEL_DIR="$HOME/Documents/notes"
./verify-ids
```

The tool will:
- Check all `.md` files in the specified directory
- Parse the YAML front-matter to extract the `ID` field
- Compare it with the filename (without `.md` extension)
- Report any mismatches with an error exit code
