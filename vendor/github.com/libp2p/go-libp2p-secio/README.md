# go-libp2p-secio

[![](https://img.shields.io/badge/made%20by-Protocol%20Labs-blue.svg?style=flat-square)](https://protocol.ai)
[![](https://img.shields.io/badge/project-libp2p-yellow.svg?style=flat-square)](https://libp2p.io/)
[![](https://img.shields.io/badge/freenode-%23libp2p-yellow.svg?style=flat-square)](http://webchat.freenode.net/?channels=%23libp2p)
[![Discourse posts](https://img.shields.io/discourse/https/discuss.libp2p.io/posts.svg)](https://discuss.libp2p.io)
[![GoDoc](https://godoc.org/github.com/libp2p/go-libp2p-secio?status.svg)](https://godoc.org/github.com/libp2p/go-libp2p-secio)
[![Build Status](https://travis-ci.org/libp2p/go-libp2p-secio.svg?branch=master)](https://travis-ci.org/libp2p/go-libp2p-secio)

> go-libp2p's secio encrypted transport

Package `go-libp2p-secio` is a libp2p [stream security transport](https://github.com/libp2p/go-stream-security). Connections wrapped by `secio` use secure sessions provided by this package to encrypt all traffic. A TLS-like handshake is used to setup the communication channel.

## Install

`go-libp2p-secio` is a standard Go module which can be installed with:

```sh
go get github.com/libp2p/go-libp2p-secio
```

This repo is [gomod](https://github.com/golang/go/wiki/Modules)-compatible, and users of
go 1.11 and later with modules enabled will automatically pull the latest tagged release
by referencing this package. Upgrades to future releases can be managed using `go get`,
or by editing your `go.mod` file as [described by the gomod documentation](https://github.com/golang/go/wiki/Modules#how-to-upgrade-and-downgrade-dependencies).

## Usage

For more information about how `go-libp2p-secio` is used in the libp2p context, you can see the [go-libp2p-conn](https://github.com/libp2p/go-libp2p-conn) module.

## Contribute

Feel free to join in. All welcome. Open an [issue](https://github.com/libp2p/go-libp2p-secio/issues)!

This repository falls under the IPFS [Code of Conduct](https://github.com/libp2p/community/blob/master/code-of-conduct.md).

### Want to hack on IPFS?

[![](https://cdn.rawgit.com/jbenet/contribute-ipfs-gif/master/img/contribute.gif)](https://github.com/ipfs/community/blob/master/contributing.md)

## License

MIT

---

The last gx published version of this module was: 2.0.30: QmSVaJe1aRjc78cZARTtf4pqvXERYwihyYhZWoVWceHnsK
