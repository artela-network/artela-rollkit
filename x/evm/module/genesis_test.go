package evm_test

import (
	"testing"

	keepertest "github.com/artela-network/artela-rollkit/testutil/keeper"
	"github.com/artela-network/artela-rollkit/testutil/nullify"
	evm "github.com/artela-network/artela-rollkit/x/evm/module"
	"github.com/artela-network/artela-rollkit/x/evm/types"

	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.EvmKeeper(t)
	evm.InitGenesis(ctx, k, genesisState)
	got := evm.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
