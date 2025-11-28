#!/bin/bash

# Set default paths
ZETTEL_DIR="${ZETTEL_DIR:-$HOME/Projects/Zettelkasten}"
ZETTEL_BACKUP_DIR="${ZETTEL_BACKUP_DIR:-$HOME/.backups/zk}"

# Create backup directory if it doesn't exist
mkdir -p "$ZETTEL_BACKUP_DIR"

# Generate ISO-8601 timestamp
timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Get the basename of the zettelkasten directory
zettel_basename=$(basename "$ZETTEL_DIR")

# Create tar.gz archive
tar -czf "$ZETTEL_BACKUP_DIR/${zettel_basename}-${timestamp}.tar.gz" -C "$(dirname "$ZETTEL_DIR")" "$zettel_basename"

echo "Backup created: $ZETTEL_BACKUP_DIR/${zettel_basename}-${timestamp}.tar.gz"
