# IBC -> ERC20 Mapping

This document introduces the design, structure, working principles, and how to use the ERC20 protocol to interact with IBC assets transferred across chains.

## Design and Structure

Design document references:

[Overview Design](https://forum.artela.network/t/erc20-wrapper-of-x-bank-module-allows-eoas-and-smart-contracts-on-evm-to-interact-with-the-bank-module-through-the-wrapper-contract/12/6)

[Detailed Design](https://forum.artela.network/t/detailed-desgined-of-ibc-erc20/13/13)

### Goals

The integration of IBC (Inter-Blockchain Communication) and ERC20 (Ethereum’s token standard) provides an efficient and simple management method for cross-chain asset transfer. By implementing ERC20 mapping, Cosmos ecosystem tokens (such as ATOM) can be managed using the ERC20 protocol on Artela.

### Key Components

- **Proxy**: An ERC20-compatible proxy for IBC assets, where all operations for IBC assets are carried out through this contract interface.
- **erc20.go**: This is an implementation of the proxy interface, used to wrap Cosmos tokens on Ethereum-compatible chains to comply with the ERC20 standard. It allows Cosmos tokens to interact with Ethereum DApps and smart contracts.
- **store.go**: Used to manage the relationship pairs between IBC and ERC20.

### Mapping Process

1. **Native Token Transfer**: The user initiates a transfer on the Cosmos chain and transfers tokens through IBC to the Artela chain.
2. **ERC20 Mapping Contract**: Query the denomination of the transferred asset on Artela. A proxy contract is manually deployed on Artela, and the address of this proxy contract is the ERC20 interface contract corresponding to the asset.
3. **Transfer**: Call the deployed proxy contract to transfer assets, and the underlying IBC asset will be transferred accordingly.

## Working Principle

### 1. IBC Transfer

When an IBC transfer is initiated on the Cosmos chain, the user generates a cross-chain transfer request via a Cosmos wallet and sends the tokens to an IBC transfer relay chain. This chain forwards the message to the target chain, where the target chain verifies the signature, checks the transfer status, and ensures the cross-chain transfer is legitimate.

### 2. Deploying the ERC20 Proxy Contract

The ERC20 Proxy contract on the target chain acts as a bridge between Cosmos tokens and ERC20 tokens. It allows users to interact with the underlying IBC assets using the ERC20 protocol on the target chain. The ERC20 Proxy contract is located in `x/evm/precompile/erc20/proxy/ERC20Proxy.sol`, and the denomination specified in the constructor corresponds to the IBC asset’s denomination.

### 3. Token Mapping

By freely mapping IBC assets to the ERC20 interface, the underlying IBC assets can circulate through the ERC20 protocol. The ERC20 standard serves as the mapping for the IBC asset.

## Usage

### Deploying the ERC20 Wrapper Contract

To deploy the ERC20 Wrapper contract on the target chain, follow these steps:

1. **IBC Transfer**

    Use the Cosmos IBC transfer protocol to transfer assets from other chains to Artela, using wallets like Keplr.

2. **Query Transferred Assets**

    Use `artelad query bank balances {cosmos address}` to query the denomination of the asset to be mapped. For example, the query structure is as follows:

    ```sh
    artelad query bank balances art1dsgnmxpeuwsuvrc92qyy2v6jgsu2clajnfsxee
    balances:
    - amount: "1"
      denom: ibc/725907476F79A96A2650A4D124501B5D236AB9DDFAF216F929833C6B51E42902
    - amount: "999999999999999999999"
      denom: uart
    pagination:
      next_key: null
      total: "0"
    ```

3. **Deploy the Wrapper Contract**

    Deploy the `x/evm/precompile/erc20/proxy/ERC20Proxy.sol` contract, where the constructor parameter specifies the IBC asset to be mapped (in this example, `ibc/725907476F79A96A2650A4D124501B5D236AB9DDFAF216F929833C6B51E42902`).
    Copy the address of the deployed contract.
    | If a mapping contract already exists, you can query the contract address. See step 5 for the query method.

4. **Use ERC20**

    Import the contract address from step 3 into your wallet and add it as an asset. You can then use this asset for querying or transferring.

5. **Query the Mapped Token Pairs**

    a. Query the mapped contract address by the denomination:

    ```sh
    curl localhost:8545 -H "Content-Type:application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_getAddressByDenom","params":["{ibc denom}", "latest"],"id":1}'
    ```

    b. Query the denomination by the contract address:

    ```sh
    curl localhost:8545 -H "Content-Type:application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_getDenomByAddress","params":["{proxy address}", "latest"],"id":1}'
    ```
