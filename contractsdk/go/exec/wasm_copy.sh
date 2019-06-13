#!/bin/bash

set -e

# 为了减少wasm 合约的代码体积，wasm引用剔除了grpc的pb文件
# 当代码同时引用包含grpc和不包含grcp的pb文件的时候回出现`proto: duplicate proto type registered`这样的错误
# 因此单独拷贝一份wasm的实现代码，唯一的区别就是对pb的引用

sed -e 's|!wasm|wasm|g; s|contractsdk/go/pb|contractsdk/go/litepb|g' contract_context.go > contract_context_wasm.go

gofmt -w contract_context_wasm.go
