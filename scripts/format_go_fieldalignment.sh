#!/usr/bin/env bash
set -euo pipefail

# Format Go field alignment
# Usage: format_go_fieldalignment.sh

until fieldalignment -fix ./...; do
  true
done > /dev/null
