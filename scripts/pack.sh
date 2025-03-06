#!/bin/bash
set -eo pipefail

# Optional arguments enumerated
# 1 - Build folder
# 2 - Pack folder : folder where to put zipped files

BIN=${1:-build}
PACK=${2:-pack}

# Does build folder exist?
if [[ ! -d "$BIN" ]]; then
  echo "no build folder";
  exit 1;
fi

# Delete current pack folder if existing
if [[ -d "$PACK" ]]; then
   rm -rf "${PACK:?}"
fi

# check for all binaries
# TODO check for correct version of all binaries
if [[ ! -x "$BIN"/gopherciser ]]; then
  echo "Linux binary not found"
  exit 3
fi

# check for zip command
if [[ ! -x $(command -v zip) ]]; then
  echo "zip command not found"
  exit 4
fi

# check for attributions file
if [[ ! -e licenses.txt ]]; then
  echo "licenses.txt not found"
  exit 5
fi

# check for documentation
if [[ ! -e "$BIN"/Readme.txt ]]; then
  echo "Readme.txt not found"
  exit 6
fi

Pack(){
  if [ $# -ne 2 ]; then
    echo "Pack called with $# arguments"
    exit 7
  fi

  local binary="$1"
  if [[ ! -x "$binary" ]]; then
    echo "$binary binary not found"
    exit 8
  fi

  local destination="$2"
  if [[ -z "$destination" ]]; then
    echo "Pack called with empty destination"
    exit 9
  fi

  echo "Packing" "$destination"

  # Create os subfolders
  mkdir -p "${PACK:?}"/"$destination" # error if destination is blank

  # copy binaries
  cp "$binary" "$PACK"/"$destination"

  # copy changelog
  if [[ ! -e changelog.md ]]; then
    # only warn since running this locally should require changelog
    echo "warning: changelog.md not found"
  else
    cp changelog.md "$PACK"/"$destination"
  fi

  # copy attributions file
  cp licenses.txt "$PACK"/"$destination"

  # copy documentation
  cp -r "$BIN"/Readme.txt "$PACK"/"$destination"/Readme.txt


  # zip folders
  cwd=$(pwd)
  cd "$PACK"
  zip -r "$destination".zip "$destination"
  rm -rf "$destination"
  cd "$cwd"
}

Pack "${BIN:?}"/gopherciser gopherciser_linux
Pack "$BIN"/gopherciser.exe gopherciser_windows
Pack "$BIN"/gopherciser_osx gopherciser_osx
