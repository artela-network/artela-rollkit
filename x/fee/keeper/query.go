package keeper

import (
	"github.com/artela-network/artela-rollkit/x/fee/types"
)

var _ types.QueryServer = Keeper{}
