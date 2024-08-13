package evm

import (
	"bytes"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	artela "github.com/artela-network/artela-rollkit/ethereum/types"
	"github.com/artela-network/artela-rollkit/x/evm/keeper"
	"github.com/artela-network/artela-rollkit/x/evm/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, accountKeeper types.AccountKeeper, genState types.GenesisState) {
	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(fmt.Errorf("error setting params %s", err))
	}

	// ensure evm module account is set
	if addr := accountKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic("the EVM module account has not been set")
	}

	for _, account := range genState.Accounts {
		address := common.HexToAddress(account.Address)
		accAddress := sdk.AccAddress(address.Bytes())

		// check that the EVM balance the matches the account balance
		acc := accountKeeper.GetAccount(ctx, accAddress)
		if acc == nil {
			panic(fmt.Errorf("account not found for address %s", account.Address))
		}

		ethAcct, ok := acc.(artela.EthAccountI)
		if !ok {
			panic(
				fmt.Errorf("account %s must be an EthAccount interface, got %T",
					account.Address, acc,
				),
			)
		}

		code := common.Hex2Bytes(account.Code)
		codeHash := crypto.Keccak256Hash(code)

		if len(account.Code) != 0 && !bytes.Equal(ethAcct.GetCodeHash().Bytes(), codeHash.Bytes()) {
			s := "the evm states code doesn't match with the codehash\n"
			panic(fmt.Sprintf("%s account: %s , evm states codehash: %v, ethAccount codehash: %v, evm states code: %s\n",
				s, account.Address, codeHash, ethAcct.GetCodeHash(), account.Code))
		}

		k.SetCode(ctx, codeHash.Bytes(), code)

		for _, storage := range []types.State(account.Storage) {
			k.SetState(ctx, address, common.HexToHash(storage.Key), common.HexToHash(storage.Value).Bytes())
		}
	}
}

// ExportGenesis returns the module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
