package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramsmodule "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/artela-network/artela-rollkit/x/fee/types"
)

type (
	Keeper struct {
		cdc                   codec.BinaryCodec
		storeService          store.KVStoreService
		transientStoreService store.TransientStoreService
		logger                log.Logger

		// the address capable of executing a MsgUpdateParams message. Typically, this
		// should be the x/gov module account.
		authority string

		// Legacy subspace
		ss paramsmodule.Subspace
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	ss paramsmodule.Subspace,
	storeService store.KVStoreService,
	transientStoreService store.TransientStoreService,
	logger log.Logger,
	authority string,

) Keeper {
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address: %s", authority))
	}

	return Keeper{
		cdc:                   cdc,
		ss:                    ss,
		storeService:          storeService,
		transientStoreService: transientStoreService,
		authority:             authority,
		logger:                logger,
	}
}

// ----------------------------------------------------------------------------
// Parent Block Gas Used
// Required by EIP1559 base fee calculation.
// ----------------------------------------------------------------------------

// SetBlockGasWanted sets the block gas wanted to the store.
// CONTRACT: this should be only called during EndBlock.
func (k Keeper) SetBlockGasWanted(ctx context.Context, gas uint64) {
	kvStore := k.storeService.OpenKVStore(ctx)
	gasBz := sdk.Uint64ToBigEndian(gas)
	_ = kvStore.Set(types.KeyPrefixBlockGasWanted, gasBz)

	k.Logger().Debug("setState: SetBlockGasWanted",
		"key", "KeyPrefixBlockGasWanted",
		"gas", fmt.Sprintf("%d", gas))
}

// GetBlockGasWanted returns the last block gas wanted value from the store.
func (k Keeper) GetBlockGasWanted(ctx context.Context) uint64 {
	kvStore := k.storeService.OpenKVStore(ctx)
	bz, _ := kvStore.Get(types.KeyPrefixBlockGasWanted)
	if len(bz) == 0 {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// GetTransientGasWanted returns the gas wanted in the current block from transient store.
func (k Keeper) GetTransientGasWanted(ctx context.Context) uint64 {
	transientStore := k.transientStoreService.OpenTransientStore(ctx)
	bz, _ := transientStore.Get(types.KeyPrefixTransientBlockGasWanted)
	if len(bz) == 0 {
		return 0
	}
	return sdk.BigEndianToUint64(bz)
}

// SetTransientBlockGasWanted sets the block gas wanted to the transient store.
func (k Keeper) SetTransientBlockGasWanted(ctx context.Context, gasWanted uint64) {
	transientStore := k.transientStoreService.OpenTransientStore(ctx)
	gasBz := sdk.Uint64ToBigEndian(gasWanted)
	_ = transientStore.Set(types.KeyPrefixTransientBlockGasWanted, gasBz)

	k.Logger().Debug("setState: SetTransientBlockGasWanted",
		"key", "KeyPrefixTransientBlockGasWanted",
		"gasWanted", fmt.Sprintf("%d", gasWanted))
}

// AddTransientGasWanted adds the cumulative gas wanted in the transient store
func (k Keeper) AddTransientGasWanted(ctx context.Context, gasWanted uint64) (uint64, error) {
	result := k.GetTransientGasWanted(ctx) + gasWanted
	k.SetTransientBlockGasWanted(ctx, result)
	return result, nil
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger() log.Logger {
	return k.logger.With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
