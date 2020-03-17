#!/bin/bash

cd `dirname $0`
# install docker in precondition

# build examples
mkdir -p build
for elem in `ls example`; do
    cc=example/$elem

    # build single cc file
    if [[ -f $cc ]]; then
        out=build/$(basename $elem .cc).wasm
        echo "build $cc"
        xdev build -o $out $cc
    fi

    # build package
    if [[ -d $cc ]]; then
        echo "build $cc"
        bash -c "cd $cc && xdev build && mv $elem.wasm ../build"
    fi
    echo 
done

