package keeper

import (
	"github.com/artela-network/artela-rollkit/x/evm/types"
)

var _ types.QueryServer = Keeper{}
