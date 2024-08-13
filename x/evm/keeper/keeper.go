package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/artela-network/artela-rollkit/x/evm/types"
)

type (
	Keeper struct {
		cdc          codec.BinaryCodec
		storeService store.KVStoreService
		logger       log.Logger

		// the address capable of executing a MsgUpdateParams message. Typically, this
		// should be the x/gov module account.
		authority string
	}
)

func (k Keeper) Account(ctx context.Context, request *types.QueryAccountRequest) (*types.QueryAccountResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (k Keeper) CosmosAccount(ctx context.Context, request *types.QueryCosmosAccountRequest) (*types.QueryCosmosAccountResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (k Keeper) ValidatorAccount(ctx context.Context, request *types.QueryValidatorAccountRequest) (*types.QueryValidatorAccountResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (k Keeper) Balance(ctx context.Context, request *types.QueryBalanceRequest) (*types.QueryBalanceResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (k Keeper) Storage(ctx context.Context, request *types.QueryStorageRequest) (*types.QueryStorageResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (k Keeper) Code(ctx context.Context, request *types.QueryCodeRequest) (*types.QueryCodeResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (k Keeper) EthCall(ctx context.Context, request *types.EthCallRequest) (*types.MsgEthereumTxResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (k Keeper) EstimateGas(ctx context.Context, request *types.EthCallRequest) (*types.EstimateGasResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (k Keeper) TraceTx(ctx context.Context, request *types.QueryTraceTxRequest) (*types.QueryTraceTxResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (k Keeper) TraceBlock(ctx context.Context, request *types.QueryTraceBlockRequest) (*types.QueryTraceBlockResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (k Keeper) BaseFee(ctx context.Context, request *types.QueryBaseFeeRequest) (*types.QueryBaseFeeResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (k Keeper) GetSender(ctx context.Context, tx *types.MsgEthereumTx) (*types.GetSenderResponse, error) {
	//TODO implement me
	panic("implement me")
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	logger log.Logger,
	authority string,

) Keeper {
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address: %s", authority))
	}

	return Keeper{
		cdc:          cdc,
		storeService: storeService,
		authority:    authority,
		logger:       logger,
	}
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger() log.Logger {
	return k.logger.With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
