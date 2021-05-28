#!/bin/bash

# install protoc 3.7.1
# export GO111MODULES=on
# go install github.com/golang/protobuf/protoc-gen-go@v1.3.3
# go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v1.16.0

protoc -I ./ ./*.proto \
    -I ./googleapis \
    --go_opt=paths=source_relative \
    --go_out=plugins=grpc:./ \
    --grpc-gateway_out=logtostderr=true:./
