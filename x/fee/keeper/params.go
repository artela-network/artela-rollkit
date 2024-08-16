package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/runtime"
	cosmos "github.com/cosmos/cosmos-sdk/types"

	"github.com/artela-network/artela-rollkit/x/fee/types"
)

// GetParams returns the total set of fee market parameters.
func (k Keeper) GetParams(ctx cosmos.Context) (params types.Params) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(types.ParamsKey)
	if len(bz) == 0 {
		panic("fee params are not set")
	}

	k.cdc.MustUnmarshal(bz, &params)
	return params
}

// SetParams sets the fee market params in a single key
func (k Keeper) SetParams(ctx cosmos.Context, params types.Params) error {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}

	store.Set(types.ParamsKey, bz)
	k.Logger().Debug("setState: SetBlockGasWanted",
		"key", string(types.ParamsKey),
		"params", fmt.Sprintf("%+v", params))

	return nil
}
