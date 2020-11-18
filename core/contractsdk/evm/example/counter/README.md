# 编译

`solc --abi --bin Counter.sol -o .`
生成的合约二进制文件和abi文件，Counter.bin和Counter.abi

# 部署

`./xchain-cli evm deploy --account XC1111111111111113@xuper --cname counterevm --fee 22787517 Counter.bin --abi Counter.abi`

成功后会生成合约地址，后续使用这个地址来调用合约

# 调用

``` bash
# 计数器，增加值
$ ./xchain-cli evm invoke --method increase -a '{"key":"stones"}' counterevm --fee 22787517 --abi Counter.abi
# 计数器，查询值
$ ./xchain-cli evm query --method get -a '{"key":"stones"}' counterevm --abi Counter.abi 
```