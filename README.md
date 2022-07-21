# Luanet Node

[![GoDoc](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square&cacheSeconds=3600)](https://godoc.org/github.com/ipfs/kubo)
[![CircleCI](https://img.shields.io/circleci/build/github/ipfs/kubo?style=flat-square&cacheSeconds=3600)](https://circleci.com/gh/ipfs/kubo)

## What is go-lua?

The earliest and most widely used implementation of IPFS node for Luanet.

It includes:
- an IPFS daemon server
- extensive [command line tooling](https://docs.ipfs.io/reference/cli/)
- an [HTTP Gateway](https://github.com/ipfs/specs/tree/main/http-gateways#readme) (`/ipfs/`, `/ipns/`) for serving content to HTTP browsers with an extra authentication layer to protect private file saved by Luanet renters.
- an HTTP RPC API (`/api/v0`) for controlling the daemon node

## What is IPFS?

IPFS is a global, versioned, peer-to-peer filesystem. It combines good ideas from previous systems such as Git, BitTorrent, Kademlia, SFS, and the Web. It is like a single BitTorrent swarm, exchanging git objects. IPFS provides an interface as simple as the HTTP web, but with permanence built-in. You can also mount the world at /ipfs.

For more info see: https://docs.ipfs.tech/concepts/what-is-ipfs/

## Architecture Diagram
![](https://user-images.githubusercontent.com/106291312/180193047-89c198bc-2151-40de-a974-ee5721ed6116.png)


## Table of Contents

- [kubo](#kubo)
  - [What is IPFS?](#what-is-ipfs)
  - [Install](#install)
    - [System Requirements](#system-requirements)
    - [Install prebuilt binaries](#install-prebuilt-binaries)
    - [Build from Source](#build-from-source)
      - [Install Go](#install-go)
      - [Download and Compile IPFS](#download-and-compile-ipfs)
        - [Cross Compiling](#cross-compiling)
        - [OpenSSL](#openssl)
      - [Troubleshooting](#troubleshooting)
    - [Updating](#updating)
      - [Using ipfs-update](#using-ipfs-update)
      - [Downloading builds using IPFS](#downloading-builds-using-ipfs)
  - [Unofficial Linux packages](#unofficial-linux-packages)
    - [ArchLinux](#arch-linux)
    - [Nix](#nix)
    - [Solus](#solus)
    - [openSUSE](#opensuse)
    - [Guix](#guix)
    - [Snap](#snap)
  - [Unofficial MacOS packages](#unofficial-macos-packages)
    - [MacPorts](#macports)
    - [Nix](#nix-1)
    - [Homebrew](#homebrew)  
  - [Unofficial Windows packages](#unofficial-windows-packages)
    - [Chocolatey](#chocolatey)
    - [Scoop](#scoop)
  - [Build from Source](#build-from-source)
    - [Install Go](#install-go)
    - [Download and Compile IPFS](#download-and-compile-ipfs)
      - [Cross Compiling](#cross-compiling)
      - [OpenSSL](#openssl)
    - [Troubleshooting](#troubleshooting)
- [Getting Started](#getting-started)
  - [Usage](#usage)
  - [Some things to try](#some-things-to-try)
  - [Troubleshooting](#troubleshooting-1)
- [Packages](#packages)
- [Development](#development)
  - [Map of Implemented Subsystems](#map-of-implemented-subsystems)
  - [CLI, HTTP-API, Architecture Diagram](#cli-http-api-architecture-diagram)
  - [Testing](#testing)
  - [Development Dependencies](#development-dependencies)
  - [Developer Notes](#developer-notes)
- [Maintainer Info](#maintainer-info)
- [Contributing](#contributing)
- [License](#license)

## Install

The canonical download instructions for IPFS are over at: https://docs.ipfs.tech/install/. It is **highly recommended** you follow those instructions if you are not interested in working on IPFS development.

### System Requirements

IPFS can run on most Linux, macOS, and Windows systems. We recommend running it on a machine with at least 2 GB of RAM and 2 CPU cores. On systems with less memory, it may not be completely stable.

If your system is resource-constrained, we recommend:

1. Installing OpenSSL and rebuilding go-lua manually with `make build GOTAGS=openssl`. See the [download and compile](#download-and-compile-ipfs) section for more information on compiling go-lua.
2. Initializing your daemon with `ipfs init --profile=lowpower`

### Install prebuilt binaries

[![Downloads](https://img.shields.io/github/v/release/ipfs/kubo?label=dist.ipfs.io&logo=ipfs&style=flat-square&cacheSeconds=3600)](https://github.com/luanet/go-lua/releases)

### Build from Source

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/ipfs/kubo?label=Requires%20Go&logo=go&style=flat-square&cacheSeconds=3600)

go-lua's build system requires Go and some standard POSIX build tools:

* GNU make
* Git
* GCC (or some other go compatible C Compiler) (optional)

To build without GCC, build with `CGO_ENABLED=0` (e.g., `make build CGO_ENABLED=0`).

#### Install Go

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/ipfs/kubo?label=Requires%20Go&logo=go&style=flat-square&cacheSeconds=3600)

If you need to update: [Download latest version of Go](https://golang.org/dl/).

You'll need to add Go's bin directories to your `$PATH` environment variable e.g., by adding these lines to your `/etc/profile` (for a system-wide installation) or `$HOME/.profile`:

```
export PATH=$PATH:/usr/local/go/bin
export PATH=$PATH:$GOPATH/bin
```

(If you run into trouble, see the [Go install instructions](https://golang.org/doc/install)).

#### Download and Compile

```
$ git clone https://github.com/luanet/go-lua.git

$ cd go-lua
$ make install
```

Alternatively, you can run `make build` to build the go-lua binary (storing it in `cmd/ipfs/ipfs`) without installing it.

**NOTE:** If you get an error along the lines of "fatal error: stdlib.h: No such file or directory", you're missing a C compiler. Either re-run `make` with `CGO_ENABLED=0` or install GCC.

##### Cross Compiling

Compiling for a different platform is as simple as running:

```
make build GOOS=myTargetOS GOARCH=myTargetArchitecture
```

##### OpenSSL

To build go-lua with OpenSSL support, append `GOTAGS=openssl` to your `make` invocation. Building with OpenSSL should significantly reduce the background CPU usage on nodes that frequently make or receive new connections.

Note: OpenSSL requires CGO support and, by default, CGO is disabled when cross-compiling. To cross-compile with OpenSSL support, you must:

1. Install a compiler toolchain for the target platform.
2. Set the `CGO_ENABLED=1` environment variable.

#### Troubleshooting

- Separate [instructions are available for building on Windows](docs/windows.md).
- `git` is required in order for `go get` to fetch all dependencies.
- Package managers often contain out-of-date `golang` packages.
  Ensure that `go version` reports at least 1.10. See above for how to install go.
- If you are interested in development, please install the development
dependencies as well.
- Shell command completions can be generated with one of the `ipfs commands completion` subcommands. Read [docs/command-completion.md](docs/command-completion.md) to learn more.
- See the [misc folder](https://github.com/ipfs/kubo/tree/master/misc) for how to connect IPFS to systemd or whatever init system your distro uses.

## Getting Started

### Usage

[![docs: Command-line quick start](https://img.shields.io/static/v1?label=docs&message=Command-line%20quick%20start&color=blue&style=flat-square&cacheSeconds=3600)](https://docs.ipfs.tech/how-to/command-line-quick-start/)
[![docs: Command-line reference](https://img.shields.io/static/v1?label=docs&message=Command-line%20reference&color=blue&style=flat-square&cacheSeconds=3600)](https://docs.ipfs.tech/reference/kubo/cli/)

To start using IPFS, you must first initialize IPFS's config files on your
system, this is done with `ipfs init`. See `ipfs init --help` for information on
the optional arguments it takes. After initialization is complete, you can use
`ipfs mount`, `ipfs add` and any of the other commands to explore!

### Join Lua Network

Go to [lua node](https://node.luanet.io/) and register your node ID.


### Start Your Node

`ipfs daemon`

### Troubleshooting

If you have previously installed IPFS before and you are running into problems getting a newer version to work, try deleting (or backing up somewhere else) your IPFS config directory (~/.ipfs by default) and rerunning `ipfs init`. This will reinitialize the config file to its defaults and clear out the local datastore of any bad entries.

Please direct general questions and help requests to our [forums](https://discuss.ipfs.tech).

If you believe you've found a bug, check the [issues list](https://github.com/ipfs/kubo/issues) and, if you don't see your problem there, either come talk to us on [Matrix chat](https://docs.ipfs.tech/community/chat/), or file an issue of your own!

## Development

Some places to get you started on the codebase:

- Main file: [./cmd/ipfs/main.go](https://github.com/ipfs/kubo/blob/master/cmd/ipfs/main.go)
- CLI Commands: [./core/commands/](https://github.com/ipfs/kubo/tree/master/core/commands)
- Bitswap (the data trading engine): [go-bitswap](https://github.com/ipfs/go-bitswap)
- libp2p
  - libp2p: https://github.com/libp2p/go-libp2p
  - DHT: https://github.com/libp2p/go-libp2p-kad-dht
  - PubSub: https://github.com/libp2p/go-libp2p-pubsub
- [IPFS : The `Add` command demystified](https://github.com/ipfs/kubo/tree/master/docs/add-code-flow.md)

### Map of Implemented Subsystems
**WIP**: This is a high-level architecture diagram of the various sub-systems of this specific implementation. To be updated with how they interact. Anyone who has suggestions is welcome to comment [here](https://docs.google.com/drawings/d/1OVpBT2q-NtSJqlPX3buvjYhOnWfdzb85YEsM_njesME/edit) on how we can improve this!
<img src="https://docs.google.com/drawings/d/e/2PACX-1vS_n1FvSu6mdmSirkBrIIEib2gqhgtatD9awaP2_WdrGN4zTNeg620XQd9P95WT-IvognSxIIdCM5uE/pub?w=1446&amp;h=1036">

### CLI, HTTP-API, Architecture Diagram

![](./docs/cli-http-api-core-diagram.png)

> [Origin](https://github.com/ipfs/pm/pull/678#discussion_r210410924)

Description: Dotted means "likely going away". The "Legacy" parts are thin wrappers around some commands to translate between the new system and the old system. The grayed-out parts on the "daemon" diagram are there to show that the code is all the same, it's just that we turn some pieces on and some pieces off depending on whether we're running on the client or the server.

### Testing

```
make test
```

### Development Dependencies

If you make changes to the protocol buffers, you will need to install the [protoc compiler](https://github.com/google/protobuf).

### Developer Notes

Find more documentation for developers on [docs](./docs)

## License

This project is dual-licensed under Apache 2.0 and MIT terms:

- Apache License, Version 2.0, ([LICENSE-APACHE](https://github.com/ipfs/kubo/blob/master/LICENSE-APACHE) or http://www.apache.org/licenses/LICENSE-2.0)
- MIT license ([LICENSE-MIT](https://github.com/ipfs/kubo/blob/master/LICENSE-MIT) or http://opensource.org/licenses/MIT)
