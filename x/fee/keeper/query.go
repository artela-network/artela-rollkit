package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/artela-network/artela-rollkit/x/fee/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) BaseFee(c context.Context, request *types.QueryBaseFeeRequest) (*types.QueryBaseFeeResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	res := &types.QueryBaseFeeResponse{}
	baseFee := k.GetBaseFee(ctx)

	if baseFee != nil {
		aux := sdkmath.NewIntFromBigInt(baseFee)
		res.BaseFee = &aux
	}

	return res, nil
}

func (k Keeper) BlockGas(c context.Context, request *types.QueryBlockGasRequest) (*types.QueryBlockGasResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	gas := sdkmath.NewIntFromUint64(k.GetBlockGasWanted(ctx))

	if !gas.IsInt64() {
		return nil, errorsmod.Wrapf(sdk.ErrIntOverflowCoin, "block gas %s is higher than MaxInt64", gas)
	}

	return &types.QueryBlockGasResponse{
		Gas: gas.Int64(),
	}, nil
}
