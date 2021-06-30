#!/bin/bash

cd `dirname $0`/../

HOMEDIR=`pwd`
OUTDIR="$HOMEDIR/output"
XVMDIR="$HOMEDIR/.compile_cache/xvm"

# make output dir
if [ ! -d "$OUTDIR" ];then
    mkdir $OUTDIR
fi
rm -rf "$OUTDIR/*"

# check xvm
if [ ! -f "$XVMDIR/wasm2c" ];then
    echo "please first execute: make xvm"
    exit 1
fi

function buildpkg() {
    output=$1
    pkg=$2

    version=`git rev-parse --abbrev-ref HEAD`
    if [ $? != 0 ]; then
        version="unknow"
    fi
    
    commitId=`git rev-parse --short HEAD`
    if [ $? != 0 ]; then
        commitId="unknow"
    fi

    buildTime=$(date "+%Y-%m-%d-%H:%M:%S")
    
    
    # build
    if [ ! -d "$OUTDIR/bin" ]; then
        mkdir "$OUTDIR/bin"
    fi

    ldflags="-X main.Version=$version -X main.BuildTime=$buildTime -X main.CommitID=$commitId"
    echo "go build -o "$OUTDIR/bin/$output" -ldflags \"$ldflags\" $pkg"

    go build -o "$OUTDIR/bin/$output" -ldflags \
        "-X main.Version=$version -X main.BuildTime=$buildTime -X main.CommitID=$commitId" $pkg
}

# build xchain
buildpkg xchain "$HOMEDIR/cmd/xchain/main.go"
# adapetr client
buildpkg xchain-cli "$HOMEDIR/cmd/client/main.go"

# build output
cp -r "$HOMEDIR/conf" "$OUTDIR"
cp "$HOMEDIR/auto/control.sh" "$OUTDIR"
mkdir -p "$OUTDIR/data"
cp -r "$HOMEDIR/data/genesis" "$OUTDIR/data"
cp -r "$HOMEDIR/data/mock/data/keys" "$OUTDIR/data"
cp -r "$HOMEDIR/data/mock/data/netkeys" "$OUTDIR/data"
cp "$XVMDIR/wasm2c" "$OUTDIR/bin"

echo "compile done!"
