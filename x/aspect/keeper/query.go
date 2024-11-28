package keeper

import (
	"github.com/artela-network/artela-rollkit/x/aspect/types"
)

var _ types.QueryServer = Keeper{}
