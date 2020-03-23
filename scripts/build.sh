#!/bin/bash

# make sure we can build within GOPATH
export GO111MODULE=on

# Purpose: This script compiles the Go packages and dependencies.
# Instructions: make build <BINARY>

set -eu
PREFIX=$1
BIN=$2
BRANCH=$(git rev-parse --abbrev-ref HEAD)
REVISION=$(git rev-parse --short HEAD)
BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)
VERSIONPKG=$(go list -mod readonly ./version)
BRANCH_FLAG="-X $VERSIONPKG=$BRANCH"
REVISION_FLAG="-X $VERSIONPKG.Revision=$REVISION"
VERSION_FLAG=""
BUILD_TIME_FLAG="-X $VERSIONPKG.BuildTime=$BUILD_TIME"
RELEASE_VERSION=""


# Check if integration or release based on if tag was set
CURRENT_REVISION=$(git rev-parse HEAD)
CURRENT_TAG=$(git describe --tags --match "v[0-9]*")
TAG_REVISION=$(git rev-list -n 1 "$CURRENT_TAG")
if [ "$CURRENT_REVISION" == "$TAG_REVISION" ]; then
  VERSION_FLAG="-X $VERSIONPKG.Version=$CURRENT_TAG"
  RELEASE_VERSION=$CURRENT_TAG
else
  BUILD_TIME=$(date -u +%Y%m%d%H%M%S)
  INTEGRATION_VERSION="$CURRENT_TAG-$BUILD_TIME"
  VERSION_FLAG="-X $VERSIONPKG.Version=$INTEGRATION_VERSION"
  RELEASE_VERSION=$INTEGRATION_VERSION
fi

# Trim away the initial "v" in the version tag if it exists
if [[ "$RELEASE_VERSION" =~ [0-9]+\.[0-9]+\.[0-9]+.* ]]; then
  RELEASE_VERSION="${BASH_REMATCH[0]}";
fi

# Store RELEASE_TYPE and VERSION_FLAG as status and version files respectively to be used by upstream ivy publish step
if mkdir -p "$PREFIX/$BIN"; then

    echo "$RELEASE_VERSION" > "$PREFIX/$BIN/version"

    # Build for three targets Windows (amd64), Linux (amd64), Darwin (amd64)
    # rm -f "$BINARY" done already via clean?

    # Linux amd64
    BINARY="$3"
    GOOS=linux GOARCH=amd64 go build -a -mod=readonly -tags netgo -installsuffix netgo -o "$PREFIX/$BIN/$BINARY" -ldflags "$BRANCH_FLAG $REVISION_FLAG $VERSION_FLAG $BUILD_TIME_FLAG -d -s -w"

    # Windows amd64
    BINARY="$3.exe"
    GOOS=windows GOARCH=amd64 go build -a -mod=readonly -tags netgo -installsuffix netgo -o "$PREFIX/$BIN/$BINARY" -ldflags "$BRANCH_FLAG $REVISION_FLAG $VERSION_FLAG $BUILD_TIME_FLAG -s -w"

    # Darwin amd64
    BINARY="$3_osx"
    GOOS=darwin GOARCH=amd64 go build -a -mod=readonly -tags netgo -installsuffix netgo -o "$PREFIX/$BIN/$BINARY" -ldflags "$BRANCH_FLAG $REVISION_FLAG $VERSION_FLAG $BUILD_TIME_FLAG -s -w"

else

    echo "Insufficient privileges to write to $PREFIX/$BIN, aborting"

fi
