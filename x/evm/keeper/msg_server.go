package keeper

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/cometbft/cometbft/libs/bytes"
	cometbft "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	cosmos "github.com/cosmos/cosmos-sdk/types"
	"github.com/hashicorp/go-metrics"

	"github.com/artela-network/artela-rollkit/x/evm/types"
)

type msgServer struct {
	*Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) EthereumTx(goCtx context.Context, msg *types.MsgEthereumTx) (*types.MsgEthereumTxResponse, error) {
	ctx := cosmos.UnwrapSDKContext(goCtx)

	sender := msg.From
	tx := msg.AsTransaction()
	txIndex := k.GetTxIndexTransient(ctx)

	labels := []metrics.Label{
		telemetry.NewLabel("tx_type", fmt.Sprintf("%d", tx.Type())),
	}
	if tx.To() == nil {
		labels = append(labels, telemetry.NewLabel("execution", "create"))
	} else {
		labels = append(labels, telemetry.NewLabel("execution", "call"))
	}

	response, err := k.ApplyTransaction(ctx, tx)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to apply txs")
	}

	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{"txs", "msg", "ethereum_tx", "total"},
			1,
			labels,
		)

		if response.GasUsed != 0 {
			telemetry.IncrCounterWithLabels(
				[]string{"txs", "msg", "ethereum_tx", "gas_used", "total"},
				float32(response.GasUsed),
				labels,
			)

			// Observe which users define a gas limit >> gas used. Note, that
			// gas_limit and gas_used are always > 0
			gasLimit := sdkmath.LegacyNewDec(int64(tx.Gas()))
			gasRatio, err := gasLimit.QuoInt64(int64(response.GasUsed)).Float64()
			if err == nil {
				telemetry.SetGaugeWithLabels(
					[]string{"txs", "msg", "ethereum_tx", "gas_limit", "per", "gas_used"},
					float32(gasRatio),
					labels,
				)
			}
		}
	}()

	attrs := []cosmos.Attribute{
		cosmos.NewAttribute(cosmos.AttributeKeyAmount, tx.Value().String()),
		// add event for ethereum txs hash format
		cosmos.NewAttribute(types.AttributeKeyEthereumTxHash, response.Hash),
		// add event for index of valid ethereum txs
		cosmos.NewAttribute(types.AttributeKeyTxIndex, strconv.FormatUint(txIndex, 10)),
		// add event for eth txs gas used, we can't get it from cosmos txs result when it contains multiple eth txs msgs.
		cosmos.NewAttribute(types.AttributeKeyTxGasUsed, strconv.FormatUint(response.GasUsed, 10)),
	}

	if len(ctx.TxBytes()) > 0 {
		// add event for tendermint txs hash format
		hash := bytes.HexBytes(cometbft.Tx(ctx.TxBytes()).Hash())
		attrs = append(attrs, cosmos.NewAttribute(types.AttributeKeyTxHash, hash.String()))
	}

	if to := tx.To(); to != nil {
		attrs = append(attrs, cosmos.NewAttribute(types.AttributeKeyRecipient, to.Hex()))
	}

	if response.Failed() {
		attrs = append(attrs, cosmos.NewAttribute(types.AttributeKeyEthereumTxFailed, response.VmError))
	}

	txLogAttrs := make([]cosmos.Attribute, len(response.Logs))
	for i, log := range response.Logs {
		value, err := json.Marshal(log)
		if err != nil {
			return nil, errorsmod.Wrap(err, "failed to encode log")
		}
		txLogAttrs[i] = cosmos.NewAttribute(types.AttributeKeyTxLog, string(value))
	}

	// emit events
	ctx.EventManager().EmitEvents(cosmos.Events{
		cosmos.NewEvent(
			types.EventTypeEthereumTx,
			attrs...,
		),
		cosmos.NewEvent(
			types.EventTypeTxLog,
			txLogAttrs...,
		),
		cosmos.NewEvent(
			cosmos.EventTypeMessage,
			cosmos.NewAttribute(cosmos.AttributeKeyModule, types.AttributeValueCategory),
			cosmos.NewAttribute(cosmos.AttributeKeySender, sender),
			cosmos.NewAttribute(types.AttributeKeyTxType, fmt.Sprintf("%d", tx.Type())),
		),
	})

	return response, nil
}
