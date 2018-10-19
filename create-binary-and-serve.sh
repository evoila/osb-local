#!/bin/bash

set -eu

FOLDER=$1
FILENAME=$2

if [[ -d "tmp" ]]; then
  rm -R tmp
fi
mkdir tmp
cd $FOLDER
tar -czvf  ../tmp/$FILENAME ./

cd ../tmp
python -m SimpleHTTPServer