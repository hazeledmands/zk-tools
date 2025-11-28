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
