#!/bin/bash
set -eu

PREFIX=$1
BIN=$2
TEST_REPORTS=$3

if [[ "$#" -eq 3 ]]; then

    rm -Rf "${PREFIX:?}/$BIN"
    rm -Rf "${PREFIX:?}/$TEST_REPORTS"
    rm -f c.out

else
   echo "Illegal number of arguments, please run with ./clean.sh <PREFIX> <BIN_FOLDER> <REPORTS_FOLDER>"
fi
