#!/bin/bash
set -e -x

# TODO @fengjin Add golang version check

cd `dirname $0`/../../
output_dir=output
[[ -d $output_dir ]] && echo output dir $output_dir already exists, please remove it if you want to build again &&exit -0
mkdir -p output

# build framework and tools
function buildpkg() {
    echo flag  ${XCHAIN_BUILD_FLAG}
    output=$1
    pkg=$2
    buildVersion=`git rev-parse --abbrev-ref HEAD`
    buildDate=$(date "+%Y-%m-%d-%H:%M:%S")
    commitHash=`git rev-parse --short HEAD`
    go build -o $output_dir/bin/$output ${XCHAIN_BUILD_FLAG} -ldflags "-X main.buildVersion=$buildVersion -X main.buildDate=$buildDate -X main.commitHash=$commitHash" $pkg
}


buildpkg xchain-cli github.com/xuperchain/xuperchain/core/cmd/cli
buildpkg xchain github.com/xuperchain/xuperchain/core/cmd/xchain
buildpkg xdev github.com/xuperchain/xuperchain/core/cmd/xdev
buildpkg xchain-httpgw github.com/xuperchain/xuperchain/core/gateway
buildpkg dump_chain github.com/xuperchain/xuperchain/core/test
buildpkg relayer github.com/xuperchain/xuperchain/core/cmd/relayer/relayer

echo start build plugins
# build plugins
echo "OS:"${PLATFORM}
echo "## Build Plugins..."
mkdir -p  ${output_dir}/plugins/kv  ${output_dir}/plugins/crypto  ${output_dir}/plugins/consensus  ${output_dir}/plugins/contract
go build --buildmode=plugin ${XCHAIN_BUILD_FLAG} --tags multi -o ${output_dir}/plugins/kv/kv-ldb-multi.so.1.0.0 github.com/xuperchain/xuperchain/core/kv/kvdb/plugin-ldb
go build --buildmode=plugin ${XCHAIN_BUILD_FLAG} --tags single -o ${output_dir}/plugins/kv/kv-ldb-single.so.1.0.0 github.com/xuperchain/xuperchain/core/kv/kvdb/plugin-ldb
go build --buildmode=plugin ${XCHAIN_BUILD_FLAG} --tags cloud -o ${output_dir}/plugins/kv/kv-ldb-cloud.so.1.0.0 github.com/xuperchain/xuperchain/core/kv/kvdb/plugin-ldb
go build --buildmode=plugin ${XCHAIN_BUILD_FLAG} -o ${output_dir}/plugins/kv/kv-badger.so.1.0.0 github.com/xuperchain/xuperchain/core/kv/kvdb/plugin-badger
go build --buildmode=plugin ${XCHAIN_BUILD_FLAG} -o ${output_dir}/plugins/crypto/crypto-default.so.1.0.0 github.com/xuperchain/xuperchain/core/crypto/client/xchain/plugin_impl
go build --buildmode=plugin ${XCHAIN_BUILD_FLAG} -o ${output_dir}/plugins/crypto/crypto-schnorr.so.1.0.0 github.com/xuperchain/xuperchain/core/crypto/client/schnorr/plugin_impl
go build --buildmode=plugin ${XCHAIN_BUILD_FLAG} -o ${output_dir}/plugins/crypto/crypto-gm.so.1.0.0 github.com/xuperchain/xuperchain/core/crypto/client/gm/gmclient/plugin_impl
go build --buildmode=plugin ${XCHAIN_BUILD_FLAG} -o ${output_dir}/plugins/consensus/consensus-pow.so.1.0.0 github.com/xuperchain/xuperchain/core/consensus/pow
go build --buildmode=plugin ${XCHAIN_BUILD_FLAG} -o ${output_dir}/plugins/consensus/consensus-single.so.1.0.0 github.com/xuperchain/xuperchain/core/consensus/single
go build --buildmode=plugin ${XCHAIN_BUILD_FLAG} -o ${output_dir}/plugins/consensus/consensus-tdpos.so.1.0.0 github.com/xuperchain/xuperchain/core/consensus/tdpos/main
go build --buildmode=plugin ${XCHAIN_BUILD_FLAG} -o ${output_dir}/plugins/consensus/consensus-xpoa.so.1.0.0 github.com/xuperchain/xuperchain/core/consensus/xpoa/main
go build --buildmode=plugin ${XCHAIN_BUILD_FLAG} -o ${output_dir}/plugins/p2p/p2p-p2pv1.so.1.0.0 github.com/xuperchain/xuperchain/core/p2p/p2pv1/plugin_impl
go build --buildmode=plugin ${XCHAIN_BUILD_FLAG} -o ${output_dir}/plugins/p2p/p2p-p2pv2.so.1.0.0 github.com/xuperchain/xuperchain/core/p2p/p2pv2/plugin_impl
go build --buildmode=plugin ${XCHAIN_BUILD_FLAG} -o ${output_dir}/plugins/xendorser/xendorser-default.so.1.0.0 github.com/xuperchain/xuperchain/core/server/xendorser/plugin-default
go build --buildmode=plugin ${XCHAIN_BUILD_FLAG} -o ${output_dir}/plugins/xendorser/xendorser-proxy.so.1.0.0 github.com/xuperchain/xuperchain/core/server/xendorser/plugin-proxy

# TODO @fengjin  
# Add symbol link of binary file for compatibility
cp -rf core/data ${output_dir}

cp -rf core/conf ${output_dir}
cp -rf core/cmd/relayer/conf/relayer.yaml ${output_dir}/conf
cp -rf core/cmd/cli/conf/* ${output_dir}/conf
cp -rf core/cmd/quick_shell/* ${output_dir}
mkdir -p ${output_dir}/data/blockchain

