#!/bin/bash

set -e -x

cd `dirname $0`

:<<!
# install protoc 3.7.1 
# export GO111MODULES=on
# go install github.com/golang/protobuf/protoc-gen-go
# go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway

protoc -I core/pb core/pb/*.proto \
	-I core/pb/googleapis \
	--go_out=plugins=grpc:core/pb \
	--grpc-gateway_out=logtostderr=true:core/pb 

protoc -I core/p2p/pb core/p2p/pb/*.proto  --go_out=plugins=grpc:core/p2p/pb

protoc -I core/xmodel/pb core/xmodel/pb/versioned_data.proto --go_out=core/xmodel/pb 

protoc -I core/contractsdk/pb core/contractsdk/pb/contract_service.proto \
       --go_out=plugins=grpc,paths=source_relative:core/contractsdk/go/pbrpc
protoc -I core/contractsdk/pb core/contractsdk/pb/contract.proto \
       --go_out=paths=source_relative:core/contractsdk/go/pb
protoc -I core/cmd/relayer/pb core/cmd/relayer/pb/relayer.proto \
       --go_out=core/cmd/relayer/pb
protoc -I core/contract/teevm/pb core/contract/teevm/pb/tf.proto \
       --go_out=paths=source_relative:core/contract/teevm/pb
!

# build wasm2c
make -C core/xvm/compile/wabt -j 4
cp core/xvm/compile/wabt/build/wasm2c ./

# build framework and tools
function buildpkg() {
    args=($1)
    output=${args[0]}
    pkg=${args[1]}
    tags=${args[2]}
    tag_flags=""
    if test $tags ;then
        tag_flags="-tags $tags"
    fi
    echo -e "start build $pkg ..."
    buildVersion=`git rev-parse --abbrev-ref HEAD`
    buildDate=$(date "+%Y-%m-%d-%H:%M:%S")
    commitHash=`git rev-parse --short HEAD`
    go build $tag_flags -o $output -ldflags "-X main.buildVersion=$buildVersion -X main.buildDate=$buildDate -X main.commitHash=$commitHash" $pkg
}

targets=(
 'xchain-cli github.com/xuperchain/xuperchain/core/cmd/cli'
 'xchain github.com/xuperchain/xuperchain/core/cmd/xchain'
 'xdev github.com/xuperchain/xuperchain/core/cmd/xdev'
 'xchain-httpgw github.com/xuperchain/xuperchain/core/gateway'
 'dump_chain github.com/xuperchain/xuperchain/core/test'
 'event_client github.com/xuperchain/xuperchain/core/test/pubsub'
 'relayer github.com/xuperchain/xuperchain/core/cmd/relayer/relayer'
)

for i in "${targets[@]}" ; do
    buildpkg "${i[@]}"
done

# build plugins
echo "OS:"${PLATFORM}
echo "## Build Plugins..."
mkdir -p core/plugins/kv core/plugins/crypto core/plugins/consensus core/plugins/contract

# build plugins 
function build_plugin() {
    args=($1)
    target=${args[0]}
    src=${args[1]}
    tags=${args[2]}
    tag_flags=""
    if test $tags ;then
        tag_flags="-tags $tags" 
    fi 
    go build --buildmode=plugin ${tag_flags} -o $target $src 
}

plugins=(
  'core/plugins/kv/kv-ldb-multi.so.1.0.0 github.com/xuperchain/xuperchain/core/kv/kvdb/plugin-ldb  multi'
  'core/plugins/kv/kv-ldb-single.so.1.0.0 github.com/xuperchain/xuperchain/core/kv/kvdb/plugin-ldb single'
  'core/plugins/kv/kv-badger.so.1.0.0 github.com/xuperchain/xuperchain/core/kv/kvdb/plugin-badger'
  'core/plugins/crypto/crypto-default.so.1.0.0 github.com/xuperchain/xuperchain/core/crypto/client/xchain/plugin_impl'
  'core/plugins/crypto/crypto-schnorr.so.1.0.0 github.com/xuperchain/xuperchain/core/crypto/client/schnorr/plugin_impl'
  'core/plugins/crypto/crypto-gm.so.1.0.0 github.com/xuperchain/xuperchain/core/crypto/client/gm/gmclient/plugin_impl'
  'core/plugins/consensus/consensus-pow.so.1.0.0 github.com/xuperchain/xuperchain/core/consensus/pow'
  'core/plugins/consensus/consensus-single.so.1.0.0 github.com/xuperchain/xuperchain/core/consensus/single'
  'core/plugins/consensus/consensus-tdpos.so.1.0.0 github.com/xuperchain/xuperchain/core/consensus/tdpos/main'
  'core/plugins/p2p/p2p-p2pv1.so.1.0.0 github.com/xuperchain/xuperchain/core/p2p/p2pv1/plugin_impl'
  'core/plugins/p2p/p2p-p2pv2.so.1.0.0 github.com/xuperchain/xuperchain/core/p2p/p2pv2/plugin_impl'
)

for i in "${plugins[@]}" ; do
    build_plugin "${i[@]}"
done


# build output dir
mkdir -p output
output_dir=output
mv xchain-cli xchain ${output_dir}
mv xchain-httpgw ${output_dir}
mv wasm2c ${output_dir}
mv dump_chain ${output_dir}
mv xdev ${output_dir}
mv relayer ${output_dir}
cp -rf core/plugins ${output_dir}
cp -rf core/data ${output_dir}
cp -rf core/conf ${output_dir}
cp -rf core/cmd/quick_shell/* ${output_dir}
mkdir -p ${output_dir}/data/blockchain
