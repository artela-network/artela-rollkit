package evm

import (
	"errors"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/baseapp"
	cosmos "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"

	"github.com/artela-network/artela-rollkit/app/interfaces"
	"github.com/artela-network/artela-rollkit/x/evm/types"
)

// EthSigVerificationDecorator validates an ethereum signatures
type EthSigVerificationDecorator struct {
	evmKeeper interfaces.EVMKeeper
	app       *baseapp.BaseApp
}

// NewEthSigVerificationDecorator creates a new EthSigVerificationDecorator
func NewEthSigVerificationDecorator(app *baseapp.BaseApp, ek interfaces.EVMKeeper) EthSigVerificationDecorator {
	return EthSigVerificationDecorator{
		evmKeeper: ek,
		app:       app,
	}
}

// AnteHandle validates checks that the registered chain id is the same as the one on the message, and
// that the signer address matches the one defined on the message.
// It's not skipped for RecheckTx, because it set `From` address which is critical from other ante handler to work.
// Failure in RecheckTx will prevent tx to be included into block, especially when CheckTx succeed, in which case user
// won't see the error message.
func (esvd EthSigVerificationDecorator) AnteHandle(ctx cosmos.Context, tx cosmos.Tx, simulate bool, next cosmos.AnteHandler) (newCtx cosmos.Context, err error) {
	for _, msg := range tx.GetMsgs() {
		msgEthTx, ok := msg.(*types.MsgEthereumTx)
		if !ok {
			return ctx, errorsmod.Wrapf(errortypes.ErrUnknownRequest, "invalid message type %T, expected %T", msg, (*types.MsgEthereumTx)(nil))
		}

		ethTx := msgEthTx.AsTransaction()
		sender, _, err := esvd.evmKeeper.VerifySig(ctx, ethTx)
		if err != nil {
			return ctx, err
		}

		// sender bytes should be equal with the one defined on the message
		if sender != common.Address(cosmos.MustAccAddressFromBech32(msgEthTx.From)) {
			return ctx, errors.New("sender address does not match the one defined on the message")
		}

		// Need to overwrite the From field with EVM address for future use
		msgEthTx.From = sender.Hex()
	}

	return next(ctx, tx, simulate)
}
