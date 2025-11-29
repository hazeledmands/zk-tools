#!/usr/bin/env bash
#
# Compare files in a backup tar.gz with the current Zettelkasten directory
#

set -euo pipefail

# Default directories
ZETTEL_DIR="${ZETTEL_DIR:-$HOME/Projects/Zettelkasten}"
BACKUP_DIR="${ZETTEL_BACKUP_DIR:-$HOME/.backups/zk}"

# Function to show usage
usage() {
    cat <<EOF
Usage: $0 [backup-file.tar.gz]

Compare files in a backup tar.gz with the current Zettelkasten directory.

Arguments:
    backup-file.tar.gz    Path to backup file (optional)
                         If not provided, uses the most recent backup

Environment Variables:
    ZETTEL_DIR           Path to Zettelkasten directory (default: ~/Projects/Zettelkasten)
    ZETTEL_BACKUP_DIR    Path to backup directory (default: ~/.backups/zk)

Examples:
    # Diff against most recent backup
    $0

    # Diff against specific backup
    $0 ~/.backups/zk/Zettelkasten-2025-11-28T23:25:41Z.tar.gz
EOF
    exit 1
}

# Parse arguments
if [[ "${1:-}" == "-h" ]] || [[ "${1:-}" == "--help" ]]; then
    usage
fi

# Determine which backup to use
if [[ -n "${1:-}" ]]; then
    BACKUP_FILE="$1"
    if [[ ! -f "$BACKUP_FILE" ]]; then
        echo "Error: Backup file not found: $BACKUP_FILE" >&2
        exit 1
    fi
else
    # Find most recent backup
    BACKUP_FILE=$(ls -t "$BACKUP_DIR"/*.tar.gz 2>/dev/null | head -1)
    if [[ -z "$BACKUP_FILE" ]]; then
        echo "Error: No backup files found in $BACKUP_DIR" >&2
        exit 1
    fi
    echo "Using most recent backup: $(basename "$BACKUP_FILE")"
    echo
fi

# Create temporary directory for extraction
TEMP_DIR=$(mktemp -d)
trap 'rm -rf "$TEMP_DIR"' EXIT

# Extract backup to temporary directory
echo "Extracting backup..."
tar -xzf "$BACKUP_FILE" -C "$TEMP_DIR"

# Find the extracted directory (handle both with and without parent directory)
if [[ -d "$TEMP_DIR/Zettelkasten" ]]; then
    EXTRACTED_DIR="$TEMP_DIR/Zettelkasten"
elif [[ $(find "$TEMP_DIR" -maxdepth 1 -type f -name "*.md" | wc -l) -gt 0 ]]; then
    EXTRACTED_DIR="$TEMP_DIR"
else
    echo "Error: Could not find extracted markdown files" >&2
    exit 1
fi

# Run diff
echo "Comparing with current Zettelkasten directory..."
echo "========================================"
echo

# Use diff with options for better readability
# -u: unified format
# -r: recursive
# -N: treat absent files as empty
# --brief: report only when files differ (remove this for full diff)
diff -ur --color=auto "$EXTRACTED_DIR" "$ZETTEL_DIR" || true

echo
echo "========================================"
echo "Comparison complete"
