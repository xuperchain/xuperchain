# 依赖
>=java1.8

# 编译

`mvn package -f ./core/contractsdk/java/example/builtin-types`
生成的目标包为./core/contractsdk/java/example/erc20/target/builtin-types-0.1.0-jar-with-dependencies.jar

# 部署

`./xchain-cli native deploy --account XC1111111111111113@xuper --fee 15587517 --runtime java builtin-types-0.1.0-jar-with-dependencies.jar --cname builtintypes`

成功后会生成合约地址，后续使用这个地址来调用合约

# 调用

``` bash
# 从合约里转账到其他账户
$ xchain-cli native invoke --method transfer -a '{"to":"dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN","amount":"10"}' builtintypes --fee 200000 --account XC1111111111111113@xuper
# 交易查询
$ xchain-cli native query --method getTx -a '{"txid":"e74f44d613c30637b6b0abbfa1f0ad4dc4fad3f36a947d0e7af8cdb216abd7b5"}' builtintypes
# 区块查询
$ xchain-cli native query --method getBlock -a '{"blockid":"a18f905a1ce81a78d0ea8c56002870cc046e3fc86064201bc398a4b3a2758ce2"}' builtintypes
# k-v 存储
$ xchain-cli native invoke --method put -a '{"key":"test1","value":"value1"}' builtintypes
# k-v 查询
$ xchain-cli native query --method get -a '{"key":"test1"}' builtintypes
# 迭代器查询
$ xchain-cli native query --method getList -a '{"start":"test"}' builtintypes
# invoke调用合约任意方法，并设置--amount会同时转账给合约，转账并打印转账金额
$ xchain-cli native invoke --method transferAndPrintAmount builtintypes --amount 1 --fee 10
```