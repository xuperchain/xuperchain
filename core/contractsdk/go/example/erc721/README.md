#ERC721参考手册

##ERC721简介
ERC721是数字资产合约,交易的商品是非同质性商品
其中，每一份资产，也就是token_id都是独一无二的
类似收藏品交易

##ERC721具备哪些功能
* 通过initialize方法，向交易池注入自己的token_id
  * 注意token_id必须是全局唯一
* 通过invoke方法，执行不同的交易功能
  * transfer:     userA将自己的某个收藏品token_id转给userB
  * approve:      userA将自己的某个收藏品token_id的售卖权限授予userB
  * transferFrom: userB替userA将赋予权限的收藏品token_id卖给userC
  * approveAll:   userA将自己的所有收藏品token_id的售卖权限授予userB
* 通过query方法，执行不同的查询功能
  * balanceOf: userA的所有收藏品的数量
  * totalSupply: 交易池中所有的收藏品的数量
  * approvalOf: userA授权给userB的收藏品的数量

##调用json文件示例
###initialize
```go
{
    "module_name": "native",      // native或wasm
    "contract_name": "erc721",    // contract name
    "method_name": "initialize",  // initialize or query or invoke
    "args": {
        "from": "dudu",           // userName
        "supply": "1,2"          //  token_ids
    }
}
```

###invoke
```go
{
    "module_name": "native",      // native或wasm
    "contract_name": "erc721",    // contract name
    "method_name": "invoke",      // initialize or query or invoke
    "args": {
        "action": "transfer",    // action name
        "from": "dudu",           // usera
        "to": "chengcheng",       // userb
        "token_id": "1"           // token_ids
    }
}

{
    "module_name": "native",      // native或wasm
    "contract_name": "erc721",    // contract name
    "method_name": "invoke",      // initialize or query or invoke
    "args": {
        "action": "transferFrom", // action name
        "from": "dudu",           // userA
        "caller": "chengcheng",   // userB
        "to": "miaomiao",         // userC
        "token_id": "1"           // token_ids
    }
}

{
    "module_name": "native",      // native或wasm
    "contract_name": "erc721",    // contract name
    "method_name": "invoke",      // initialize or query or invoke
    "args": {
        "action": "approve",      // action name
        "from": "dudu",           // userA
        "to": "chengcheng",       // userB
        "token_id": "1"           // token_ids
    }
}
```

###query
```go
{
    "module_name": "native",     // native或wasm
    "contract_name": "erc721",   // contract name
    "method_name": "query",      // initialize or query or invoke
    "args": {
        "action": "balanceOf",   // action name
        "from": "dudu"           // userA
    }
}

{
    "module_name": "native",     // native或wasm
    "contract_name": "erc721",   // contract name
    "method_name": "query",      // initialize or query or invoke
    "args": {
        "action": "totalSupply"  // action name
    }
}

{
    "module_name": "native",      // native或wasm
    "contract_name": "erc721",    // contract name
    "method_name": "query",       // initialize or query or invoke
    "args": {
        "action": "approvalOf",   // action name
        "from": "dudu",           // userA
        "to": "chengcheng"       // userB
    }
}
```

