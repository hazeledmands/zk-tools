#!/bin/bash

# Create backup directory if it doesn't exist
mkdir -p ~/.backups/zk/

# Generate ISO-8601 timestamp
timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Create tar.gz archive
tar -czf ~/.backups/zk/zettelkasten-${timestamp}.tar.gz -C ~/Projects Zettelkasten

echo "Backup created: ~/.backups/zk/zettelkasten-${timestamp}.tar.gz"
