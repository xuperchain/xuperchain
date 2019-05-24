# go-libp2p-discovery

[![](https://img.shields.io/badge/made%20by-Protocol%20Labs-blue.svg?style=flat-square)](http://ipn.io)
[![](https://img.shields.io/badge/project-libp2p-blue.svg?style=flat-square)](http://libp2p.io/)
[![](https://img.shields.io/badge/freenode-%23ipfs-blue.svg?style=flat-square)](http://webchat.freenode.net/?channels=%23ipfs)
[![standard-readme compliant](https://img.shields.io/badge/standard--readme-OK-green.svg?style=flat-square)](https://github.com/RichardLitt/standard-readme)

> Interfaces for active peer discovery

This package contains interfaces and utilities for active peer discovery.
Peers providing a service use the interface to advertise their presence in some namespace.
Vice versa, peers seeking a service use the interface to discover peers that have previously advertised
as service providers.
The package also includes a baseline implementation for discovery through [Content Routing](https://github.com/libp2p/go-libp2p-routing).

## Documenation

See https://godoc.org/github.com/libp2p/go-libp2p-discovery.

## Contribute

Feel free to join in. All welcome. Open an [issue](https://github.com/libp2p/go-libp2p-discovery/issues)!

This repository falls under the IPFS [Code of Conduct](https://github.com/ipfs/community/blob/master/code-of-conduct.md).

## License

MIT
