#!/bin/bash

set -x -e

c_file=$1
wasm_file="${c_file%%.*}".wasm
wat_file="${c_file%%.*}".wat
xemcc $c_file -s WASM=1 -o main.js -Os -s TOTAL_MEMORY=10MB -s ERROR_ON_UNDEFINED_SYMBOLS=0
mv -f main.wasm $wasm_file
wasm2wat -o $wat_file $wasm_file
rm -f main.js
#rm -f $wasm_file
