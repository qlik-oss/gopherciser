#!/bin/bash
# Unit test cover
set -eu

TEST_REPORTS=${1:-./.cover}

if [[ ! -f c.out ]]; then
  echo "Error: file c.out missing."
  exit 1
fi

if mkdir -p "$TEST_REPORTS"; then
  # Print per-function coverage statistics
  go tool cover -func c.out
  # Generate interactive HTML pages
  go tool cover -html c.out -o "$TEST_REPORTS"/c.html
else
  echo "Insufficient privileges to write to $TEST_REPORTS, aborting"
fi