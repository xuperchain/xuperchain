#!/bin/bash

cd `dirname $0`
export PATH=`pwd`/../../../output/:$PATH
export XDEV_ROOT=`pwd`

# install docker in precondition
if ! command -v docker &>/dev/null; then
    echo "missing docker command, please install docker first."
    exit 1
fi

# check if xdev available
if ! command -v xdev &>/dev/null; then
    project_root=$(cd ../../.. && pwd)
    echo "missing xdev command, please cd ${project_root} && make"
    exit 1
fi

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
        bash -c "cd $cc && xdev build && mv -v $elem.wasm ../../build/"
    fi
    echo 
done

