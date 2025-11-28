#!/bin/bash

set -euo pipefail

# Set default paths
ZETTEL_DIR="${ZETTEL_DIR:-$HOME/Projects/Zettelkasten}"
ZETTEL_BACKUP_DIR="${ZETTEL_BACKUP_DIR:-$HOME/.backups/zk}"

# Check if source directory exists
if [[ ! -d "$ZETTEL_DIR" ]]; then
    echo "Error: Zettelkasten directory not found: $ZETTEL_DIR" >&2
    exit 1
fi

# Create backup directory if it doesn't exist
mkdir -p "$ZETTEL_BACKUP_DIR"

# Generate ISO-8601 timestamp
timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Get the basename of the zettelkasten directory
zettel_basename=$(basename "$ZETTEL_DIR")

# Create tar.gz archive
backup_file="$ZETTEL_BACKUP_DIR/${zettel_basename}-${timestamp}.tar.gz"
if tar -czf "$backup_file" -C "$(dirname "$ZETTEL_DIR")" "$zettel_basename"; then
    echo "Backup created: $backup_file"
else
    echo "Error: Failed to create backup" >&2
    exit 1
fi
