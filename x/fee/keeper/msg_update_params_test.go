package keeper_test

import (
	"testing"

	cosmos "github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/artela-network/artela-rollkit/x/fee/types"
)

func TestMsgUpdateParams(t *testing.T) {
	k, ms, ctx := setupMsgServer(t)
	params := types.DefaultParams()
	sdkCtx := cosmos.UnwrapSDKContext(ctx)
	require.NoError(t, k.SetParams(sdkCtx, params))
	wctx := sdk.UnwrapSDKContext(ctx)

	// default params
	testCases := []struct {
		name      string
		input     *types.MsgUpdateParams
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid authority",
			input: &types.MsgUpdateParams{
				Authority: "invalid",
				Params:    params,
			},
			expErr:    true,
			expErrMsg: "invalid authority",
		},
		{
			name: "send enabled param",
			input: &types.MsgUpdateParams{
				Authority: k.GetAuthority(),
				Params:    types.Params{},
			},
			expErr: false,
		},
		{
			name: "all good",
			input: &types.MsgUpdateParams{
				Authority: k.GetAuthority(),
				Params:    params,
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ms.UpdateParams(wctx, tc.input)

			if tc.expErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErrMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
