#!/bin/bash

# Unit test cover
set -eu

# make sure we can build within GOPATH
export GO111MODULE=on

TEST_REPORTS=${1:-./.cover} # Test report folder from first argument or default to ./.cover
GO_PACKAGES=$(go list -mod readonly ./... | grep -v "test/" | grep -v "gen/" | grep -v "design")

if mkdir -p "$TEST_REPORTS"; then
  for pkg in $GO_PACKAGES; do
      go test -mod=readonly -count=1 -race -coverprofile="$TEST_REPORTS"/"$(echo "$pkg" | tr / -)".cover "$pkg";
  done
  echo "mode: set" > c.out
  grep -h -v "^mode:" "$TEST_REPORTS"/*.cover >> c.out
else
    echo "Insufficient privileges to write to $TEST_REPORTS, aborting"
fi