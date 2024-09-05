package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	artvmtype "github.com/artela-network/artela-rollkit/x/evm/artela/types"
)

func (k Keeper) JITSenderAspectByContext(ctx context.Context, userOpHash common.Hash) (common.Address, error) {
	return mustGetAspectCtx(ctx).JITManager().SenderAspect(userOpHash), nil
}

func (k Keeper) IsCommit(ctx context.Context) bool {
	return mustGetAspectCtx(ctx).EthTxContext().Commit()
}

func (k Keeper) GetAspectContext(ctx context.Context, address common.Address, key string) ([]byte, error) {
	return mustGetAspectCtx(ctx).AspectContext().Get(address, key), nil
}

func (k Keeper) SetAspectContext(ctx context.Context, address common.Address, key string, value []byte) error {
	mustGetAspectCtx(ctx).AspectContext().Add(address, key, value)
	return nil
}

func (k Keeper) GetBlockContext() *artvmtype.EthBlockContext {
	return k.BlockContext
}

func mustGetAspectCtx(ctx context.Context) *artvmtype.AspectRuntimeContext {
	aspectCtx, ok := ctx.(*artvmtype.AspectRuntimeContext)
	if ok {
		return aspectCtx
	}

	// unwrap as sdk ctx
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// get aspect context from sdk context, this will panic if aspect context is not found
	return sdkCtx.Value(artvmtype.AspectContextKey).(*artvmtype.AspectRuntimeContext)
}
