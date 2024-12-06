package aspect_test

import (
	"testing"

	keepertest "github.com/artela-network/artela-rollkit/testutil/keeper"
	"github.com/artela-network/artela-rollkit/testutil/nullify"
	aspect "github.com/artela-network/artela-rollkit/x/aspect/module"
	"github.com/artela-network/artela-rollkit/x/aspect/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.AspectKeeper(t)
	aspect.InitGenesis(ctx, k, genesisState)
	got := aspect.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
