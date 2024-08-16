package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/runtime"
	cosmos "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/params"

	"github.com/artela-network/artela-rollkit/x/evm/types"
)

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx context.Context) (params types.Params) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(types.KeyPrefixParams)
	if bz == nil {
		panic("evm params are not set")
	}

	k.cdc.MustUnmarshal(bz, &params)
	return params
}

// SetParams set the params
func (k Keeper) SetParams(ctx context.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}
	store.Set(types.KeyPrefixParams, bz)

	return nil
}

func (k *Keeper) GetChainConfig(ctx cosmos.Context) *params.ChainConfig {
	chainParams := k.GetParams(ctx)
	ethCfg := chainParams.ChainConfig.EthereumConfig(ctx.BlockHeight(), k.ChainID())
	return ethCfg
}
