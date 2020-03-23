#!/bin/bash
set -eox pipefail

BIN=${1:-build}
USER=$(id -u)
PROJECTFOLDER=${2:-.}
echo PROJECTFOLDER:"$PROJECTFOLDER"

# Test project folder and change to project folder
if [[ ! -e $PROJECTFOLDER/docs ]]; then echo "$PROJECTFOLDER"/docs not found; exit 1; fi
cd "$PROJECTFOLDER"

# Remove any existing docs folder
if [[ -d "$BIN/docs" ]]; then rm -rf "$BIN"/docs; fi

if [[ -x $(command -v docker) ]]; then
    # circleci don't allow for volume to mounted directly so need to handle build through volume container
    echo creating gopherciser data volume...
    GOPHERCISERDATAVOLUME=$(docker create -v /data alpine:latest /bin/true)

    echo creating pandoc container...
    PANDOCCONTAINER=$(docker create --volumes-from "$GOPHERCISERDATAVOLUME" -it --entrypoint /bin/ash pandoc/core:latest)
    docker start "$PANDOCCONTAINER"

    # Create folder structure needed
    echo Creating folder structure...
    find docs -type d ! -iname "*md" -print0 | xargs -0 -I{} docker exec "$PANDOCCONTAINER" mkdir -p /data/{}
    find docs -type d ! -iname "*md" -print0 | xargs -0 -I{} docker exec "$PANDOCCONTAINER" mkdir -p /data/build/{}

    # copy filter to volume
    echo copying files...
    docker cp scripts/docfilter.lua "$GOPHERCISERDATAVOLUME":/data

    # Generate all html files
    for f in $(find docs -iname "*.md"); do
        echo Generating "${f%.*}".html...
        docker cp "$f" "$GOPHERCISERDATAVOLUME":/data/"$f"
        docker exec "$PANDOCCONTAINER" pandoc /data/"$f" -o /data/build/"${f%.*}".html --lua-filter=/data/docfilter.lua
    done

    echo "copying generated files from volume to $BIN..."
    docker cp "$GOPHERCISERDATAVOLUME":/data/build/. "$BIN"

    # Copy non-md files
    echo copying non-md files to "$BIN"...
    find docs -type f ! -iname "*md" -print0 | xargs -0 -I{} cp {} "$BIN"/{}

    # Clean up
    docker stop "$PANDOCCONTAINER"
    docker rm "$PANDOCCONTAINER"
    docker rm "$GOPHERCISERDATAVOLUME"
else
    echo WARNING! Docker not found markdown documentation not converted to HTML

    # Copy docs folder as is
    echo Copying markdown documents to "$BIN"
    cp -r docs "$BIN"
fi
