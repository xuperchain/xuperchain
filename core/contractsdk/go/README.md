## 使用本地环境构建
1. 安装依赖工具
    go get -u -v github.com/xuperchain/xuperchain/core/cmd/xdev

2. 合约编译和测试
   go get -u -v github.com/xuperchain/xuperchain/core/contractsdk/go
   make -C core/contractsdk/go 构建合约模板，所有构建的产出在  xuperchain/core/contractsdk/go/example/wasm 目录下
   make -C core/contractsdk/go test 执行合约模板单测
   
3.删除构建产物
    make -C core/contractsdk/go clean

## 使用容器环境构建
TBD
