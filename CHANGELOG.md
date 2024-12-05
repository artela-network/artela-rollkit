# Changelog

All notable changes to this project will be documented in this file.

## v0.4.9 - 2024.12.6

### Enhancements

- Added support for ERC20-compatible mapping for IBC assets. [Issue #8](https://github.com/artela-network/artela-rollkit/issues/8) [Commit b0570ae](https://github.com/artela-network/artela-rollkit/commit/b0570ae5e6ce6f8f7d14711e83d0d34644ca037f)

### Improvements

- Refactored Aspect storage and Aspect System Contract. [Issue #8](https://github.com/artela-network/artela-rollkit/issues/8) [Commit a151dbc...afd1a43](https://github.com/artela-network/artela-rollkit/compare/a151dbc0a6f508de70763e76b978560b633bf020...afd1a43c8c7d24257d1f8cb9c50e4f106dc8682a)

### Fixes

- Fixed an encoding issue encountered during private key import. [Issue #6](https://github.com/artela-network/artela-rollkit/issues/6)
- Fixed support for the EthSecp256k1 private key algorithm. [Issue #9](https://github.com/artela-network/artela-rollkit/issues/9) [Commit 0b8e174](https://github.com/artela-network/artela-rollkit/commit/0b8e174331cc71014acf7edc58db699f9ecdd450)
- Fixed transaction conversion support during trace tx. [Issue #9](https://github.com/artela-network/artela-rollkit/issues/9) [Commit 4312d81](https://github.com/artela-network/artela-rollkit/commit/4312d8191412f76c434018431557f4634f97565a)
- Fixed an issue with block fetching caused by Rollkit block submission mechanism. [Issue #9](https://github.com/artela-network/artela-rollkit/issues/9) [Commit e0ec03a](https://github.com/artela-network/artela-rollkit/commit/e0ec03a4f675a289e576558726f11c48d2524dfa)
