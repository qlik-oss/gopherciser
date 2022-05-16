#!/bin/bash
#set -eo pipefail

PROJECT=${1:-.}
# TODO default to temp folder
LICENSEFOLDER=${2:-$(mktemp -d)}

echoerr() { cat <<< "$@" 1>&2; }

function rmlicensefolder() {
	if [[ -d "$LICENSEFOLDER" ]]; then
		rm -rf "$LICENSEFOLDER"
	fi
}

if [[ ! -x $(command -v go-licenses) ]]; then
  echoerr go-licenses not found, please install https://github.com/google/go-licenses
  exit 1
fi

if ! go list "$PROJECT" >/dev/null 2>&1; then
  echoerr "$PROJECT" is not a go project
  exit 2
fi

echo "Saving licenses to $LICENSEFOLDER"

rmlicensefolder
trap rmlicensefolder EXIT

echo "THE FOLLOWING SETS FORTH ATTRIBUTION NOTICES FOR THIRD PARTY SOFTWARE THAT MAY BE CONTAINED IN PORTIONS OF THE GOPHERCISER PRODUCT." > licenses.txt
{ echo; echo "------";  } >> licenses.txt

go-licenses save --save_path "$LICENSEFOLDER" "$PROJECT"

echo "Collecting license files..."
for p in $(find "$LICENSEFOLDER" -type f -iname "*license*"|sort); do
	TITLE=${p#"$LICENSEFOLDER/"}
	echo "Adding $TITLE..."
	{ echo ; echo "PROJECT: $TITLE" ; echo ; cat "$p"; echo ; echo "------";} >> licenses.txt
done
