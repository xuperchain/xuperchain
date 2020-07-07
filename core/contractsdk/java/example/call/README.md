# 依赖
>=java1.8

# 编译

`mvn package -f ./core/contractsdk/java/example/call/c1`
`mvn package -f ./core/contractsdk/java/example/call/c2`
生成的目标包为./core/contractsdk/java/example/erc20/target/c1-0.1.0-jar-with-dependencies.jar
生成的目标包为./core/contractsdk/java/example/erc20/target/c2-0.1.0-jar-with-dependencies.jar

# 部署

`./xchain-cli native deploy --account XC1111111111111113@xuper --fee 15587517 --runtime java c2-0.1.0-jar-with-dependencies.jar --cname c2`
`./xchain-cli native deploy --account XC1111111111111113@xuper --fee 15587517 --runtime java c1-0.1.0-jar-with-dependencies.jar --cname c1`

成功后会生成合约地址，后续使用这个地址来调用合约

# 调用

``` bash
# 转账到c1和c2合约
$ ./xchain-cli transfer --to c1 --amount 100000
$ ./xchain-cli transfer --to c2 --amount 100000
# 跨合约调用
$ ./xchain-cli native invoke --method invoke -a '{"to":"test"}' c1 --fee 200000
```