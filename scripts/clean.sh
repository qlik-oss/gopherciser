#!/bin/bash
set -eu

PREFIX=$1
BIN=$2

if [[ "$#" -eq 2 ]]; then

    rm -Rf "${PREFIX:?}/$BIN"
    rm -f "${PREFIX:?}/coverage.csv"
    rm -f "${PREFIX:?}/coverage.html"

else
   echo "Illegal number of arguments, please run with ./clean.sh <PREFIX> <BIN_FOLDER>"
fi
