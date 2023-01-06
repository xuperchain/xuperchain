# XuperChain

[![Build Status](https://travis-ci.org/xuperchain/xuperchain.svg?branch=master)](https://travis-ci.org/xuperchain/xuperchain)
[![Go Report Card](https://goreportcard.com/badge/github.com/xuperchain/xuperchain)](https://goreportcard.com/report/github.com/xuperchain/xuperchain)
[![GolangCI](https://golangci.com/badges/github.com/golangci/golangci-lint.svg)](https://golangci.com)
[![License](https://img.shields.io/github/license/xuperchain/xuperchain?style=flat-square)](/LICENSE)
[![Release](https://img.shields.io/github/v/release/xuperchain/xuperchain?style=flat-square)][LatestRelease]

中文 | [English](README.md)

---

## XuperChain是什么?

**XuperChain**是超级链体系下的第一个开源项目，是构建超级联盟网络的底层方案。
基于XuperCore动态内核实现的。您可以使用XuperChain，作为区块链基础设施，构建合规的区块链网络。
。

核心特点
* **动态内核**
    * 基于动态内核技术，实现无内核代码侵入的自由扩展内核核心组件和轻量级的扩展订制内核引擎，满足面向各类场景的区块链实现的需要。
    * 提供了全面的、高性能的标准内核组件实现。
    * 全面降低区块链研发成本，开启一键发链新时代。

* **高性能**
    * 原创的XuperModel模型，真正实现了智能合约的并发执行和验证。
    * TDPOS算法确保大规模节点下的快速共识。
    * 使用AOT加速的WASM虚拟机，合约运行速度接近native程序。

* **更安全**
    * 多私钥保护的账户体系。
    * 鉴权支持权重累计、集合运算等灵活的策略。

* **易扩展**
    * 鲁棒的P2P网络，支持广域网超大规模节点。
    * 底层账本支持分叉管理，自动收敛一致性，实现真正全球化部署。

* **多语言开发智能合约**
    * 通过原创的XuperBridge技术，可插拔多语言虚拟机。

* **高灵活性**
    * 可插拔、插件化的设计使得用户可以方便选择适合自己业务场景的解决方案。

## 快速试用

### 环境配置

* 操作系统：支持Linux以及Mac OS
* 开发语言：Go 1.14.*及以上
* 编译器：GCC 4.8.x及以上
* 版本控制工具：Git

### 构建

克隆XuperChain仓库
```
git clone https://github.com/xuperchain/xuperchain
```

> **注意**: 
> 
> `master` 分支是日常开发分支，会包含最新的功能，但是 **不稳定**。
> 
> 生产环境请使用[最新的已发布分支][LatestRelease]。

编译
```
cd xuperchain
make
```

跑单测
```
make test
```

单机版xchain
```
cd ./output
sh ./control.sh start
./bin/xchain-cli status
```

多节点xchain

生成多节点。
在运行下面的命令之前，请确保已经运行`make`去编译代码。
```
make testnet
```

进入testnet目录，分别启动三个节点(确保端口未被占用)。
```
cd ./testnet/node1
sh ./control.sh start
cd ../node2
sh ./control.sh start
cd ../node3
sh ./control.sh start
```

观察每个节点状态
```
./bin/xchain-cli status -H :37101
./bin/xchain-cli status -H :37102
./bin/xchain-cli status -H :37103
```

## 文档


关于XuperChain更详细、更深入的使用方法请查看 [XuperChain文档][Docs]

## 如何参与开发
1. 阅读源代码，了解我们当前的开发方向
2. 找到自己感兴趣的功能或模块
3. 进行开发，开发完成后自测功能是否正确，并运行make & make test
4. 发起pull request
5. 更多详情请参见 [贡献指南][Contribution]

## 许可证
XuperChain使用的许可证是Apache 2.0

## 联系我们
商务合作，请Email：xchain-help@baidu.com, 来源请注明 GitHub。

如果你对XuperChain开源技术及应用感兴趣，
欢迎添加`百度超级链·小助手`微信，回复`技术论坛进群`，
加入“百度超级链开发者社区”，
与百度资深工程师深度交流!

微信二维码如下:

<img width="291" alt="496bd829f51cda8f4c8027daf0e6b543" src="https://user-images.githubusercontent.com/51440377/210507301-84a45cc8-0841-4c55-9398-6d03f395c0b7.png">

[Contribution]: docs/zh_cn/contribute/contribute-guideline.md
[LatestRelease]: https://github.com/xuperchain/xuperchain/releases/latest
[Docs]: https://xuper.baidu.com/n/xuperdoc/index.html
