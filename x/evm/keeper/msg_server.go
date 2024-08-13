package keeper

import (
	"context"

	"github.com/artela-network/artela-rollkit/x/evm/types"
)

type msgServer struct {
	Keeper
}

func (k msgServer) EthereumTx(ctx context.Context, tx *types.MsgEthereumTx) (*types.MsgEthereumTxResponse, error) {
	//TODO implement me
	panic("implement me")
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}
