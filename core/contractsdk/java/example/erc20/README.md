# 依赖
>=java1.8

# 编译

`mvn package -f ./core/contractsdk/java/example/erc20`
生成的目标包为./core/contractsdk/java/example/erc20/target/erc20-0.1.0-jar-with-dependencies.jar

# 部署

`./xchain-cli native deploy --account XC1111111111111113@xuper -a '{"totalSupply": "10000000000"}' --fee 15587517 --runtime java erc20-0.1.0-jar-with-dependencies.jar --cname erc20java`

其中-a 10000000000指定初始资产，
--account XC1111111111111113@xuper，指定合约账户

成功后会生成合约地址，后续使用这个地址来调用合约

# 调用

``` bash
# 查询资产总量
$ xchain-cli native query --method totalSupply erc20java
# 转账
$ xchain-cli native invoke --method transfer -a '{"to":"dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN","token":"10"}' erc20java --fee 200000 --account XC1111111111111113@xuper
# 余额查询
$ xchain-cli native query --method balance -a '{"account":"XC1111111111111113@xuper"}' erc20java
# 授权额度
$ xchain-cli native invoke --method approve -a '{"to":"dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN","token":"10000"}' erc20java --fee 200000 --account XC1111111111111113@xuper
# 查询授权额度
$ xchain-cli native query --method allowance -a '{"from":"XC1111111111111113@xuper","to":"dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN"}' erc20java
# 使用授权额度
$ xchain-cli native invoke --method transferFrom -a '{"from":"XC1111111111111113@xuper","to":"dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN","token":"88"}' erc20java --fee 200000
# 增发token
$ xchain-cli native invoke --method mint --account XC1111111111111113@xuper -a '{"amount":"100"}' erc20java --fee 200000
```