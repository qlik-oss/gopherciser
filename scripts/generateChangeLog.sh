#!/bin/bash

CHFILE=${1:-changelog.md}

if [ -z "$CHFILE" ]; then
    echo no changelog file defined
    exit 1
fi

CURVER=$(git describe --match "v[0-9]*")
PREVVER=$(git describe --match "v[0-9]*" --abbrev=0 HEAD~1)

echo "# Changes $PREVVER -> $CURVER" > "$CHFILE"
git log "$PREVVER"..@ --pretty=format:'[%h](https://github.com/qlik-oss/gopherciser/commit/%H) %s
' >> "$CHFILE"
