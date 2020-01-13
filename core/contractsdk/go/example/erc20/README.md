# 依赖
>=go1.11

# 编译

`GOOS=js GOARCH=wasm go build -o erc20.wasm erc20.go`

# 部署

`xchain-cli wasm deploy -a 1000000 erc20.wasm`

其中-a 1000000指定初始资产

成功后会生成合约地址，后续使用这个地址来调用合约

# 调用

``` bash
# 向addr1转账100
$ xchain-cli wasm invoke $codeAddr --action transfer -a addr1,100
# 查询addr1余额
$ xchain-cli wasm query $codeAddr --action balanceOf -a addr1
# 向addr1授权200
$ xchain-cli wasm invoke $codeAddr --action approve -a addr1,200
# 查询授权额度
$ xchain-cli wasm query $codeAddr --action allowance -a $myaddress,addr1
# 换一个钱包地址addr1
# 使用myaddress授权的额度向其他地址转账
$ xchain-cli wasm invoke $codeAddr --action transferFrom -a $myaddress,addr2,200
```
