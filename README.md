# XuperChain

[![Build Status](https://travis-ci.org/xuperchain/xuperchain.svg?branch=master)](https://travis-ci.org/xuperchain/xuperchain)
[![Go Report Card](https://goreportcard.com/badge/github.com/xuperchain/xuperchain)](https://goreportcard.com/report/github.com/xuperchain/xuperchain)
[![GolangCI](https://golangci.com/badges/github.com/golangci/golangci-lint.svg)](https://golangci.com)
[![License](https://img.shields.io/github/license/xuperchain/xuperchain?style=flat-square)](/LICENSE)
[![Release](https://img.shields.io/github/v/release/xuperchain/xuperchain?style=flat-square)][LatestRelease]

[中文](README-CN.md) | English

---

## What is XuperChain

**XuperChain**, the first open source project of **XuperChain Lab**, 
introduces an underlying solution to build the super alliance network, 
based on the dynamic kernel of xupercore. 
You can use xuperchain as a blockchain infrastructure to build a compliant blockchain network.

XuperChain is the underlying solution for union networks with following highlight features:
* **Dynamic kernel**

    * Based on the dynamic kernel technology, the free extension kernel components without kernel code intrusion and lightweight extension customized kernel engine are implemented to meet the needs of blockchain implementation for various scenarios.
    * It provides a comprehensive and high-performance implementation of standard kernel components.
    * Comprehensively reduce the cost of blockchain research and development, and open a new era of one click chain development.

* **High Performance**

    * Creative XuperModel technology makes contract execution and verification run parallel.
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

## Quick start

### Requirements

* OS Support: Linux and Mac OS
* Go 1.14.x or later
* GCC 4.8.x or later
* Git

### Build

Clone the repository

```
git clone https://github.com/xuperchain/xuperchain
```

> **Note**:
> 
> `master` branch contains the latest features but might be **unstable**.
> 
> For production use, please check out [the latest release][LatestRelease].

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

### Run 

#### Run single node blockchain
There is an output folder if build successfully. Enter the output folder, create a default chain & start blockchains:

```
cd ./output
sh control.sh start
```

By doing this, a blockchain named "xuper" is created, you can find the data of this blockchain at `./data/blockchain/xuper/`.

By default, the `xuper` chain will produce a block every 3 seconds, try the following command to see the `trunkHeight` of chain and make sure it's growing.

```
./bin/xchain-cli status
```

#### Run multi nodes blockchain

Generate multi nodes.
Before running the following command, make sure you have run `make` to make the code.
```
make testnet
```

Enter the testnet directory, and then start three nodes separately (make sure the port is not used)
```
cd ./testnet/node1
sh ./control.sh start
cd ../node2
sh ./control.sh start
cd ../node3
sh ./control.sh start
```

Observe the status of each node
```
./bin/xchain-cli status -H :37101
./bin/xchain-cli status -H :37102
./bin/xchain-cli status -H :37103
```

## Documentation

We have new documentation of Chinese version at [XuperChain Chinese Docs][Docs],
the English version is coming soon.

## How to Contribute

We encourage you to contribute to XuperChain.

Please review the [Contribution guidelines][Contribution] for information on how to get started contributing to the project.

## License

XuperChain is under the [Apache License, Version 2.0](https://github.com/xuperchain/xuperchain/blob/master/LICENSE).

## Contact

For business cooperation, please email：xchain-help@baidu.com (Note with source: GitHub)。

If you are interested in the open source technology and application of XuperChain, 
welcome to add `百度超级链·小助手` in WeChat,
join the Baidu Super Chain Developer Community by replying `技术论坛进群`,
and have in-depth exchanges with Baidu senior engineers!

WeChat QR code is as follows:
![微信二维码](<img width="291" alt="496bd829f51cda8f4c8027daf0e6b543" src="https://user-images.githubusercontent.com/51440377/186586870-1c147ed5-6d8b-4bb6-9151-c3110b19f318.png">)

[Contribution]: docs/en_us/contribute/contribute-guideline.md
[LatestRelease]: https://github.com/xuperchain/xuperchain/releases/latest
[Docs]: https://xuper.baidu.com/n/xuperdoc/index.html