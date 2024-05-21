#!/bin/bash

# Purpose: This script lints the code.
# Instructions: make lint

# set lint level
LINTLEVEL=${1:-DEFAULT}

GOPATH=$(go env GOPATH)

echo running lint on "$LINTLEVEL" level

set -eu

# Check if linter is installed
CURVER=v$("$GOPATH"/bin/golangci-lint version | cut -f 4 -d\ )
LATESTVER=$(curl -s --head https://github.com/golangci/golangci-lint/releases/latest | grep "location:" | cut -d/ -f 8)
LATESTVER=${LATESTVER%?} # remove trailing control character

# Check if we have the correct version, otherwise install it
echo "detected current lint version <$CURVER>"
echo "detected latest lint version <$LATESTVER>"
if ! [[ "$CURVER" == "$LATESTVER" ]]; then
    echo golangci-lint not installed or incorrect version, installing latest golangci-lint
    curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s -- -b "$(go env GOPATH)"/bin latest
else
    echo Lint tool already at correct version
fi

echo Running lint
case $LINTLEVEL in
# minimal amount of linting currently running clean. More linters and and rules will be added as more lint errors are fixed.
# currently no subrules disabled in "min"
MIN)
  "$GOPATH"/bin/golangci-lint run --timeout 5m
  ;;
# "full" set of linters
*)
  "$GOPATH"/bin/golangci-lint run --timeout 5m
  ;;
esac
