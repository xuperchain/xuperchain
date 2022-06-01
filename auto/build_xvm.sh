#!/bin/bash
set -e 

cd `dirname $0`/../

HOMEDIR=`pwd`
OUTDIR="$HOMEDIR/.compile_cache/xvm"
XVMPKG="https://github.com/xuperchain/xvm.git"
XVM_VERSION=v0.1.0

function buildxvm() {
    # clean dir
    rm -rf $OUTDIR
    mkdir -p $OUTDIR

    # download pkg
    echo "start downloading xvm pkg..."
    git clone -b ${XVM_VERSION} ${XVMPKG} ${OUTDIR}/xvm

    # make
    make -C "$OUTDIR/xvm/compile/wabt" -j 4
    if [ $? != 0 ]; then
        echo "complie xvm failed"
        exit 1
    fi

    cp -r "$OUTDIR/xvm/compile/wabt/build/wasm2c" "$OUTDIR"
}

# build xvm
if [ ! -f "$OUTDIR/wasm2c" ]; then
    buildxvm
fi

echo "compile done!"
