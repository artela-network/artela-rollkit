package keeper

import (
	"context"

	"cosmossdk.io/core/store"
)

func (k *Keeper) GetPrefixStore(ctx context.Context, prefix []byte) store.KVStore {
	kvStore := k.storeService.OpenKVStore(ctx)

}
