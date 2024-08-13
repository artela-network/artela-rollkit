package api

import (
	"context"

	"github.com/pkg/errors"

	artelatypes "github.com/artela-network/aspect-core/types"

	"github.com/artela-network/artela-rollkit/x/evm/artela/types"
)

func GetStateDBHostInstance(ctx context.Context) (artelatypes.StateDBHostAPI, error) {
	aspectCtx, ok := ctx.(*types.AspectRuntimeContext)
	if !ok {
		return nil, errors.New("GetStateDBHostInstance: unwrap AspectRuntimeContext failed")
	}
	return aspectCtx.StateDb(), nil
}
