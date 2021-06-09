#!/bin/bash

# Purpose: This script lints the code.
# Instructions: make lint

# On darwin you can also install with homebrew 
# brew install golangci/tap/golangci-lint

# set OS 
OS= 
if [[ "$OSTYPE" == "linux-gnu" ]]; then
        OS=Linux
elif [[ "$OSTYPE" == "darwin"* ]]; then
        # Mac OSX
        OS=Darwin
elif [[ "$OSTYPE" == "cygwin" ]]; then
        # POSIX compatibility layer and Linux environment emulation for Windows
        OS=Windows
else
        # Unknown, assume Linux
        OS=Linux
fi

# set lint level
LINTLEVEL=${1:-DEFAULT}

GOPATH=$(go env GOPATH)

echo running lint on "$LINTLEVEL" level

set -eu

export GO111MODULE=on

GLVERSION=1.40.1
CURVER=NULL

# Determine OS and act accordingly to extract MD5 sum as an override due to -version being missing in golangci-lint CLI

# Check if linter is installed
CURVER=$("$GOPATH"/bin/golangci-lint --version | cut -f 4 -d\ )

# Check if we have the correct version, otherwise install it
echo detected lint version "$CURVER"
if ! [[ "$CURVER" == "$GLVERSION" ]]; then
    echo golangci-lint not installed or incorrect version, installing golangci-lint v$GLVERSION
    curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s -- -b "$(go env GOPATH)"/bin v$GLVERSION
else
    echo Lint tool already at correct version
fi

echo Running lint
case $LINTLEVEL in
# minimal amount of linting currently running clean. More linters and and rules will be added as more lint errors are fixed.
# Current status:
# linters currently looked at: govet
# subrules currently disabled:
#   * structtag : We override tags in engima, so structtag complains about repeating tags, found no way to tell it this is intended.
MIN)
  "$GOPATH"/bin/golangci-lint run -D structcheck
  ;;
# Default set of linters
*)
  "$GOPATH"/bin/golangci-lint run
  ;;
esac
