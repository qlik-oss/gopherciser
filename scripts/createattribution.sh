#!/bin/bash
#set -eo pipefail

PROJECT=${1:-.}

echoerr() { cat <<< "$@" 1>&2; }

if [[ ! -x $(command -v go-licenses) ]]; then
  echoerr go-licenses not found, please install https://github.com/google/go-licenses
  exit 1
fi

if ! go list "$PROJECT" >/dev/null 2>&1; then
  echoerr "$PROJECT" is not a go project
  exit 2
fi

echo "THE FOLLOWING SETS FORTH ATTRIBUTION NOTICES FOR THIRD PARTY SOFTWARE THAT MAY BE CONTAINED IN PORTIONS OF THE GOPHERCISER PRODUCT." > licenses.txt
{ echo; echo "------";  } >> licenses.txt

LICENSES=$(go-licenses csv "$PROJECT" 2>/dev/null | grep -vi unknown)

for i in $LICENSES; do
  # todo get from mod cache instead
  MOD=$(echo "$i" | cut -f 1 -d,)
  FP=$(echo "$i" | cut -f 2 -d, | sed 's/github.com/raw.githubusercontent.com/g' | sed 's/blob\/master/master/g' | sed 's/master\/.*\/LICENSE/master\/LICENSE/g')
  LICENSE=$(curl -s --fail "$FP")
  if [[ $? -eq 0 ]]; then
     { echo ; echo "$MOD"; echo ; echo "$LICENSE"; echo ; echo "------";} >> licenses.txt
    else
      echoerr curl of "$FP" failed
  fi
done
