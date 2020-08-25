#!/bin/bash
set -e -x

cd `dirname $0`/../../

# build wasm2c
mkdir -p core_tmp
cd core_tmp
git clone https://github.com/xuperchain/xupercore.git
cd xupercore
make -C xvm/compile/wabt -j 4
cp xvm/compile/wabt/build/wasm2c ../../
cd ../../
rm -rf core_tmp

# build framework and tools
function buildpkg() {
    output=$1
    pkg=$2
    buildVersion=`git rev-parse --abbrev-ref HEAD`
    buildDate=$(date "+%Y-%m-%d-%H:%M:%S")
    commitHash=`git rev-parse --short HEAD`
    go build -o $output -ldflags "-X main.buildVersion=$buildVersion -X main.buildDate=$buildDate -X main.commitHash=$commitHash" $pkg
}

buildpkg xchain-cli github.com/xuperchain/xupercore/cmd/cli
buildpkg xchain github.com/xuperchain/xuperchain/core/cmd/xchain
buildpkg xdev github.com/xuperchain/xupercore/cmd/xdev
buildpkg xchain-httpgw github.com/xuperchain/xuperchain/core/gateway
buildpkg dump_chain github.com/xuperchain/xupercore/test
buildpkg relayer github.com/xuperchain/xupercore/cmd/relayer/relayer

# build plugins
echo "OS:"${PLATFORM}
echo "## Build Plugins..."
mkdir -p core/plugins/kv core/plugins/crypto core/plugins/consensus core/plugins/contract
go build --buildmode=plugin --tags multi -o core/plugins/kv/kv-ldb-multi.so.1.0.0 github.com/xuperchain/xuperchain/core/kv/kvdb/plugin-ldb
go build --buildmode=plugin --tags single -o core/plugins/kv/kv-ldb-single.so.1.0.0 github.com/xuperchain/xupercore/kv/kvdb/plugin-ldb
go build --buildmode=plugin --tags cloud -o core/plugins/kv/kv-ldb-cloud.so.1.0.0 github.com/xuperchain/xuperchain/core/kv/kvdb/plugin-ldb
go build --buildmode=plugin -o core/plugins/kv/kv-badger.so.1.0.0 github.com/xuperchain/xuperchain/core/kv/kvdb/plugin-badger
go build --buildmode=plugin -o core/plugins/crypto/crypto-default.so.1.0.0 github.com/xuperchain/xupercore/crypto/client/xchain/plugin_impl
go build --buildmode=plugin -o core/plugins/crypto/crypto-schnorr.so.1.0.0 github.com/xuperchain/xupercore/crypto/client/schnorr/plugin_impl
go build --buildmode=plugin -o core/plugins/crypto/crypto-gm.so.1.0.0 github.com/xuperchain/xuperchain/core/crypto/client/gm/gmclient/plugin_impl
go build --buildmode=plugin -o core/plugins/consensus/consensus-pow.so.1.0.0 github.com/xuperchain/xuperchain/core/consensus/pow
go build --buildmode=plugin -o core/plugins/consensus/consensus-single.so.1.0.0 github.com/xuperchain/xupercore/consensus/single
go build --buildmode=plugin -o core/plugins/consensus/consensus-tdpos.so.1.0.0 github.com/xuperchain/xuperchain/core/consensus/tdpos/main
go build --buildmode=plugin -o core/plugins/consensus/consensus-xpoa.so.1.0.0 github.com/xuperchain/xuperchain/core/consensus/xpoa/main
go build --buildmode=plugin -o core/plugins/p2p/p2p-p2pv1.so.1.0.0 github.com/xuperchain/xupercore/p2p/p2pv1/plugin_impl
go build --buildmode=plugin -o core/plugins/p2p/p2p-p2pv2.so.1.0.0 github.com/xuperchain/xupercore/p2p/p2pv2/plugin_impl
go build --buildmode=plugin -o core/plugins/xendorser/xendorser-default.so.1.0.0 github.com/xuperchain/xupercore/server/xendorser/plugin-default
go build --buildmode=plugin -o core/plugins/xendorser/xendorser-proxy.so.1.0.0 github.com/xuperchain/xupercore/server/xendorser/plugin-proxy

# build output dir
mkdir -p output
output_dir=output
mv xchain xchain-cli ${output_dir}
mv xchain-httpgw ${output_dir}
mv wasm2c ${output_dir}
mv xdev ${output_dir}
mv dump_chain ${output_dir}
# cp -rf core/plugins ${output_dir}
cp -rf core/data ${output_dir}
cp -rf core/conf ${output_dir}
mkdir -p ${output_dir}/data/blockchain
