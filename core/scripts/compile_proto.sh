#!/bin/bash
set -e -x

cd `dirname $0`/../../

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