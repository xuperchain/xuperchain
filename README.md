### 调研:区块链底层开源技术哪家强？
每每到了区块链技术选型的时刻， 是性能？还是安全？还是易用性更重要？你是否经常犯难？填写这份小调研https://www.wjx.cn/m/80446725.aspx  助你快速总结出区块链开源技术选型比较的方法，得到此次调研的结论，同时还有机会获得小度智能音箱、区块链应用案例集、京东卡等精美小礼品一份~ 🎁

# XuperChain

[![Build Status](https://travis-ci.org/xuperchain/xuperchain.svg?branch=master)](https://travis-ci.org/xuperchain/xuperchain)
[![Go Report Card](https://goreportcard.com/badge/github.com/xuperchain/xuperchain)](https://goreportcard.com/report/github.com/xuperchain/xuperchain)
[![GolangCI](https://golangci.com/badges/github.com/golangci/golangci-lint.svg)](https://golangci.com)
[![License](https://img.shields.io/github/license/xuperchain/xuperchain?style=flat-square)](/LICENSE)
[![Release](https://img.shields.io/github/v/release/xuperchain/xuperchain?style=flat-square)](https://github.com/xuperchain/xuperchain/releases/latest)

[中文说明](#中文说明-1)
-----
## What is XuperChain

**XuperChain**, the first open source project of **XuperChain Lab**, introduces a highly flexible blockchain architecture with great transaction performance.
 
XuperChain is the underlying solution for union networks with following highlight features:

* **High Performance**

    * Creative XuperModel technology makes contract execution and verification run parallelly.
    * TDPoS ensures quick consensus in a large scale network.
    * WASM VM using AOT technology.

* **Solid Security**

    * Contract account protected by multiple private keys ensures assets safety.
    * Flexible authorization system supports weight threshold, AK sets and could be easily extended.

* **High Scalability**

    * Robust P2P network supports a large scale network with thousands of nodes.
    * Branch management on ledger makes automatic convergence consistency and supports global deployment.

* **Multi-Language Support**: Support pluggable multi-language contract VM using XuperBridge technology.

* **Flexibility**:  Modular and pluggable design provides high flexibility for users to build their blockchain solutions for various business scenarios.

## Contact
E-mail: xchain-help@baidu.com

## Quick start

### Requirements

* OS Support: Linux and Mac OS
* Go 1.12.x or later
* GCC 4.8.x or later
* Git

### Build

Clone the repository

```
git clone https://github.com/xuperchain/xuperchain
```

**Note**: `master` branch contains latest features but might be **unstable**. for production use, please checkout our release branch. the latest release branch is `v3.7`.

Enter the xuperchain folder and build the code:

```
cd xuperchain
make
```

Note that if you are using Go 1.11 or later, go modules are used to download 3rd-party dependencies by default. You can also disable go modules and use the prepared dependencies under vendor folder.

Run test:
```
make test
```

Use Docker to build xuperchain see [docker build](core/scripts/README.md)

### Run 

There is an output folder if build successfully. Enter the output folder, create a default chain firstly:

```
cd ./output
./xchain-cli createChain
```

By doing this, a blockchain named "xuper" is created, you can find the data of this blockchain at `./data/blockchain/xuper/`.

Then start the node and run XuperChain full node servers:

```
nohup ./xchain &
```

By default, the `xuper` chain will produce a block every 3 seconds, try the following command to see the `trunkHeight` of chain and make sure it's growing.

```
./xchain-cli status
```

## Run with Docker

### Build image

```bash
docker build . -t xuperchain
```

### Run image

```bash
# run xchain daemon
docker run -d -p 37101:37101 -p 47101:47101 --rm --name xchain xuperchain
# enter running container
docker exec -ti xchain bash
# run command
./xchain-cli status
```

> This is only a demo for local single container, you can use volume to mount and overwrite configurations.

## Documentation

Please refer to our [wiki](https://github.com/xuperchain/xuperchain/wiki) for more  information, including how to build multi-node network, transfer to others, deploy and invoke smart contract.

We have new documentation of Chinese version at [XuperChain Chinese Docs](https://xuperchain.readthedocs.io/zh/latest/), the English version is comming soon.

## How to Contribute

We encourage you to contribute to XuperChain.

Please review the [Contribution guidelines](https://github.com/xuperchain/xuperchain/blob/master/CONTRIBUTING.md)  for information on how to get started contributing to the project.

## License

XuperChain is under the [Apache License, Version 2.0](https://github.com/xuperchain/xuperchain/blob/master/LICENSE).


=====

# 中文说明

## XuperChain是什么?

**XuperUion**是超级链体系下的第一个开源项目，是构建超级联盟网络的底层方案。

核心特点

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
* 开发语言：Go 1.12.x及以上
* 编译器：GCC 4.8.x及以上
* 版本控制工具：Git

### 构建

克隆XuperChain仓库
```
git clone https://github.com/xuperchain/xuperchain
```

**注意**: `master` 分支是日常开发分支，会包含最新的功能，但是 **不稳定**。生产环境请使用最新的已发布分支，最新的已发布分支是`v3.7`。

编译
```
cd xuperchain
make
```

跑单测
```
make test
```

使用docker来编译xueprchain见[docker build](core/scripts/README.md)

单机版xchain
```
cd ./output
./xchain-cli createChain
nohup ./xchain &
./xchain-cli status
```

## 容器运行

### 编译镜像

```bash
docker build . -t xuperchain
```

### 运行镜像

```bash
# 运行容器 daemon
docker run -d -p 37101:37101 -p 47101:47101 --rm --name xchain xuperchain
# 进入容器
docker exec -ti xchain bash
# 运行指令
./xchain-cli status
```

> 本地容器化运行的示例，实际场景中可以用卷的方式挂载并覆盖配置。

## 文档


关于XuperChain更详细、更深入的使用方法链接请查看[XuperChain文档库](https://xuperchain.readthedocs.io/zh/latest/)和[wiki](https://github.com/xuperchain/xuperchain/wiki)

## 如何参与开发
1. 阅读源代码，了解我们当前的开发方向
2. 找到自己感兴趣的功能或模块
3. 进行开发，开发完成后自测功能是否正确，并运行make & make test
4. 发起pull request
5. 更多详情请参见[链接](https://github.com/xuperchain/xuperchain/blob/master/CONTRIBUTING_CN.md)

## 许可证
XuperChain使用的许可证是Apache 2.0

## 联系我们
商务合作，请Email：xchain-help@baidu.com, 来源请注明Github。
如果你对XuperChain开源技术及应用感兴趣，欢迎添加“百度超级链·小助手“微信，回复“技术论坛进群”，加入“百度超级链开发者社区”，与百度资深工程师深度交流!微信二维码如下:

![微信二维码](https://github.com/ToWorld/xuperchain-image/blob/master/baidu-image-xuperchain.png)


