#!/bin/bash

CHFILE=${1:-changelog.md}

if [ -z "$CHFILE" ]; then
    echo no changelog file defined
    exit 1
fi


# Is this release? then skip one tag from rev-list
SKIP="--skip=1"
if [[ $(git describe --tags --match="v[0-9]*") == *"-"* ]]; then
    # not a release, don't skip
    SKIP=
fi

CURVER=$(git describe --tags --match="v[0-9]*")
PREVVER=$(git describe --abbrev=0 --tags $(git rev-list --tags="v[0-9]*" $SKIP --max-count=1))

echo "# Changes $PREVVER -> $CURVER" > "$CHFILE"
git log "$PREVVER"..@ --pretty=format:'[%h](https://github.com/qlik-oss/gopherciser/commit/%H) %s
' >> "$CHFILE"
