#!/bin/bash

set -e -x

cd `dirname $0`

:<<!
# install protoc 3.7.1 
# export GO111MODULES=on
# go install github.com/golang/protobuf/protoc-gen-go
# go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway

protoc -I pb pb/xchain.proto pb/xchain_spv.proto pb/xcheck.proto \
	-I pb/googleapis \
	--go_out=plugins=grpc:pb \
	--grpc-gateway_out=logtostderr=true:pb

protoc -I p2pv2/pb p2pv2/pb/message.proto  --go_out=p2pv2/pb

protoc -I xmodel/pb xmodel/pb/versioned_data.proto --go_out=xmodel/pb 

protoc -I contractsdk/pb contractsdk/pb/contract.proto \
       --go_out=plugins=grpc:contractsdk/go/pb
protoc -I contractsdk/pb contractsdk/pb/contract.proto \
       --go_out=contractsdk/go/litepb

!

# build wasm2c
make -C xvm/compile/wabt -j 4
cp xvm/compile/wabt/build/wasm2c ./

# build framework and tools
function buildpkg() {
    output=$1
    pkg=$2
    buildVersion=`git rev-parse --abbrev-ref HEAD`
    buildDate=$(date "+%Y-%m-%d-%H:%M:%S")
    commitHash=`git rev-parse --short HEAD`
    go build -o $output -ldflags "-X main.buildVersion=$buildVersion -X main.buildDate=$buildDate -X main.commitHash=$commitHash" $pkg
}

buildpkg xchain-cli github.com/xuperchain/xuperunion/cmd/cli
buildpkg xchain github.com/xuperchain/xuperunion/cmd/xchain
go build -o dump_chain test/dump_chain.go

# build plugins
echo "OS:"${PLATFORM}
echo "## Build Plugins..."
mkdir -p plugins/kv plugins/crypto plugins/consensus plugins/contract
go build --buildmode=plugin --tags multi -o plugins/kv/kv-ldb-multi.so.1.0.0 github.com/xuperchain/xuperunion/kv/kvdb/plugin-ldb
go build --buildmode=plugin --tags single -o plugins/kv/kv-ldb-single.so.1.0.0 github.com/xuperchain/xuperunion/kv/kvdb/plugin-ldb
go build --buildmode=plugin -o plugins/kv/kv-badger.so.1.0.0 github.com/xuperchain/xuperunion/kv/kvdb/plugin-badger
go build --buildmode=plugin -o plugins/crypto/crypto-default.so.1.0.0 github.com/xuperchain/xuperunion/crypto/client/xchain
go build --buildmode=plugin -o plugins/crypto/crypto-schnorr.so.1.0.0 github.com/xuperchain/xuperunion/crypto/client/schnorr
go build --buildmode=plugin -o plugins/consensus/consensus-pow.so.1.0.0 github.com/xuperchain/xuperunion/consensus/pow
go build --buildmode=plugin -o plugins/consensus/consensus-single.so.1.0.0 github.com/xuperchain/xuperunion/consensus/single
go build --buildmode=plugin -o plugins/consensus/consensus-tdpos.so.1.0.0 github.com/xuperchain/xuperunion/consensus/tdpos/main

# build output dir
mkdir -p output
output_dir=output
mv xchain-cli xchain ${output_dir}
mv wasm2c ${output_dir}
mv dump_chain ${output_dir}
cp -rf  plugins ${output_dir}
cp -rf data ${output_dir}
cp -rf conf ${output_dir}
cp -rf cmd/quick_shell/* ${output_dir}
mkdir -p ${output_dir}/data/blockchain
