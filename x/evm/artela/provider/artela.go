package provider

import (
	"context"
	"errors"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/ethereum/go-ethereum/common"

	asptypes "github.com/artela-network/aspect-core/types"

	"github.com/artela-network/artela-rollkit/x/evm/artela/contract"
	"github.com/artela-network/artela-rollkit/x/evm/artela/types"
)

var _ asptypes.AspectProvider = (*ArtelaProvider)(nil)

type ArtelaProvider struct {
	service      *contract.AspectService
	storeService store.KVStoreService
}

func NewArtelaProvider(storeService store.KVStoreService,
	getBlockHeight types.GetLastBlockHeight,
	logger log.Logger,
) *ArtelaProvider {
	service := contract.NewAspectService(storeService, getBlockHeight, logger)

	return &ArtelaProvider{service, storeService}
}

func (j *ArtelaProvider) GetTxBondAspects(ctx context.Context, address common.Address, point asptypes.PointCut) ([]*asptypes.AspectCode, error) {
	if ctx == nil {
		return nil, errors.New("invalid Context")
	}
	aspectCtx, ok := ctx.(*types.AspectRuntimeContext)
	if !ok {
		return nil, errors.New("failed to unwrap AspectRuntimeContext from context.Context")
	}
	return j.service.GetAspectsForJoinPoint(aspectCtx.CosmosContext(), address, point)
}

func (j *ArtelaProvider) GetAccountVerifiers(ctx context.Context, address common.Address) ([]*asptypes.AspectCode, error) {
	if ctx == nil {
		return nil, errors.New("invalid Context")
	}
	aspectCtx, ok := ctx.(*types.AspectRuntimeContext)
	if !ok {
		return nil, errors.New("failed to unwrap AspectRuntimeContext from context.Context")
	}
	return j.service.GetAccountVerifiers(aspectCtx.CosmosContext(), address)
}

func (j *ArtelaProvider) GetLatestBlock() int64 {
	return j.service.GetBlockHeight()
}
