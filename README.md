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
