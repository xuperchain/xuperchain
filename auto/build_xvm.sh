#!/bin/bash

cd `dirname $0`/../

HOMEDIR=`pwd`
OUTDIR="$HOMEDIR/.compile_cache/xvm"
XVMPKG="https://codeload.github.com/xuperchain/xvm/zip/main"

function buildxvm() {
    # clean dir
    rm -rf $OUTDIR
    mkdir -p $OUTDIR

    # download pkg
    echo "start downloading xvm pkg..."
    curl -s -L -k -o "$OUTDIR/xvm.zip" "$XVMPKG"
    if [ $? != 0 ]; then
        echo "download xvm failed"
        exit 1
    fi

    # unzip
    unzip -d "$OUTDIR" "$OUTDIR/xvm.zip"
    mv "$OUTDIR/xvm-main" "$OUTDIR/xvm"

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
