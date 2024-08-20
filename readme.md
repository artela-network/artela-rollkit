<h1 align="center"> Artela Rollkit </h1>

<div align="center">
  <a href="https://t.me/artela_official" target="_blank">
    <img alt="Telegram Chat" src="https://img.shields.io/badge/chat-telegram-blue?logo=telegram&chat">
  </a>
  <a href="https://twitter.com/Artela_Network" target="_blank">
    <img alt="Twitter Follow" src="https://img.shields.io/twitter/follow/Artela_Network">
  <a href="https://discord.gg/artela">
   <img src="https://img.shields.io/badge/chat-discord-green?logo=discord&chat" alt="Discord">
  </a>
  <a href="https://www.artela.network/">
   <img src="https://img.shields.io/badge/Artela%20Network-3282f8" alt="Artela Network">
  </a>
</div>

## Introduction


Artela Rollkit is a rollup kit built with [Celestia's Rollkit](https://github.com/rollkit/rollkit), it empowers developers to add user-defined native extensions and build feature-rich dApps. It offers extensibility that goes beyond EVM-equivalence, inter-domain interoperability, and boundless scalability with its Elastic Block Space design.

As the first rollup kit equipped with Aspects, Artela network aims to **maximize the value of Aspect and enable developers to build feature-rich dApps.**
<p align="center">
  <img src="https://docs.artela.network/assets/images/2-a4045260ad64e65eaa2af9fc50c06a4a.png" width="500" height="500">
</p>

* **Base Layer:** Provide basic functions, including consensus engine, networking, EVM environments for the smart contract execution, and WASM environments for the Aspects execution. This layer is launched by Artela.

* **Extension Layer:** Provide the Aspect SDK. Developers are able to build Aspects. Aspects have access to all APIs within the base layer and can be freely combined with smart contracts and other Aspects. Aspect is securely isolated from Base Layer, ensuring that it has no impact on the security or availability of the core network.

* **Application Layer:** Developers can build smart contracts as usual. Initially, EVM will be provided for the seamless landing of most dApps in crypto.


## Build the source

1). Set Up Your Go Development Environment<br />
Make sure you have set up your Go programming language development environment.

2). Install ignite-cli<br />

```sh
curl https://get.ignite.com/cli@v28.4.0! | bash
```

3). Install rollkit-cli <br />

```sh
curl -sSL https://rollkit.dev/install.sh | sh -s v0.13.6
```

4). Download the Source Code<br />
Obtain the project source code using the following method:

```
git clone https://github.com/artela-network/artela-rollkit.git
```

5). Compile<br />
Compile the source code and generate the executable using the Go compiler:

```
ignite chain build
```

## Executables

|  Command   | Description|
| :--------: | --------------------------------------------------------------------------------------------------------------|
| **`artela-rollkitd`** | artela-rollkitd is the core software of the Artela network. |

## Running a local dev node

1. Start `local-da` for the local development environment, make sure you have docker installed:

```sh
docker run --name local-da -tid -p 7980:7980 ghcr.io/rollkit/local-da:v0.2.1
```

if you prefer to run the local-da from the source code, you can find it out [here](https://github.com/rollkit/local-da).

2. Start `artela-rollkitd`:

```sh

```

---
Learn more about Artela in <https://artela.network/>


## License
Copyright © Artela Network, Inc. All rights reserved.

Licensed under the [Apache v2](LICENSE) License.