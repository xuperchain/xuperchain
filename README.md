# XuperUnion

[![Build Status](https://travis-ci.org/xuperchain/xuperunion.svg?branch=master)](https://travis-ci.org/xuperchain/xuperunion)
[![Go Report Card](https://goreportcard.com/badge/github.com/xuperchain/xuperunion)](https://goreportcard.com/report/github.com/xuperchain/xuperunion)

## What is XuperUnion

**XuperUnion**, the first open source project of **XuperChain**, introduces a highly flexible blockchain architecture with great transaction performance.
 
XuperUnion is the underlying solution for union networks with following highlight features:

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

## Quick start

### Requirements

* OS Support: Linux and Mac OS
* Go 1.12.x or later
* G++ 4.8.x or later
* Git

### Build

Clone the repository

```
git clone https://github.com/xuperchain/xuperunion
```

Enter the xuperunion folder and build the code:

```
make
```

Note that if you are using Go 1.11 or later, go modules are used to download 3rd-party dependencies by default. You can also disable go modules and use the prepared dependencies under vendor folder.

Run test:
```
make test
```

### Run 

There is an output folder if build successfully. Enter the output folder, create a default chain firstly:

```
./xchain-cli createChain
```

By doing this, a blockchain named "xuper" is created, you can find the data of this blockchain at `./data/blockchain/xuper/`.

Then start the node and run XuperUnion full node servers:

```
nohup ./xchain &
```

By default, the `xuper` chain will produce a block every 3 seconds, try the following command to see the `trunkHeight` of chain and make sure it's growing.

```
./xchain-cli status
```


### Documentation

Please refer to our [wiki](https://github.com/xuperchain/xuperunion/wiki) for more  information, including how to build multi-node network, transfer to others, deploy and invoke smart contract.

## How to Contribute

We encourage you to contribute to XuperUnion.

Please review the [Contribution guidelines](https://github.com/xuperchain/xuperunion/blob/master/CONTRIBUTING.md)  for information on how to get started contributing to the project.

## License

XuperUnion is under the [Apache License, Version 2.0](https://github.com/xuperchain/xuperunion/blob/master/LICENSE).

