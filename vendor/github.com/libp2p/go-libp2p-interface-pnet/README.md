go-libp2p-interface-pnet
==================

[![](https://img.shields.io/badge/made%20by-Protocol%20Labs-blue.svg?style=flat-square)](https://protocol.ai)
[![](https://img.shields.io/badge/project-libp2p-yellow.svg?style=flat-square)](https://libp2p.io/)
[![](https://img.shields.io/badge/freenode-%23libp2p-yellow.svg?style=flat-square)](http://webchat.freenode.net/?channels=%23ipfs)
[![Discourse posts](https://img.shields.io/discourse/https/discuss.libp2p.io/posts.svg)](https://discuss.libp2p.io)

> An interface providing abstraction of swarm protection for libp2p.


## Table of Contents

- [Usage](#usage)
- [Contribute](#contribute)
- [License](#license)

## Usage

Core of this interface in `Protector` that is used to protect the swarm.
It makes decisions about which streams are allowed to pass.

This interface is accepted in multiple places in libp2p but most importantly in
go-libp2p-swarm `NewSwarmWithProtector` and `NewNetworkWithProtector`.

## Implementations:

 - [go-libp2p-pnet](//github.com/libp2p/go-libp2p-pnet) - simple PSK based Protector, using XSalsa20

## Contribute

PRs are welcome!

Small note: If editing the Readme, please conform to the [standard-readme](https://github.com/RichardLitt/standard-readme) specification.

## License

MIT © Jeromy Johnson

---

The last gx published version of this module was: 3.0.0: QmW7Ump7YyBMr712Ta3iEVh3ZYcfVvJaPryfbCnyE826b4
