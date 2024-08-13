package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/artela-network/artela-rollkit/testutil/keeper"
	"github.com/artela-network/artela-rollkit/x/evm/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := keepertest.EvmKeeper(t)
	params := types.DefaultParams()

	require.NoError(t, k.SetParams(ctx, params))
	require.EqualValues(t, params, k.GetParams(ctx))
}
