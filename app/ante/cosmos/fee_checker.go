package cosmos

import (
	"fmt"
	"math"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	cosmos "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	anteutils "github.com/artela-network/artela-rollkit/app/ante/utils"
)

// DeductFeeDecorator deducts fees from the first signer of the tx.
// If the first signer does not have the funds to pay for the fees,
// and does not have enough unclaimed staking rewards, then return
// with InsufficientFunds error.
// The next AnteHandler is called if fees are successfully deducted.
//
// CONTRACT: Tx must implement FeeTx interface to use DeductFeeDecorator
type DeductFeeDecorator struct {
	accountKeeper      authante.AccountKeeper
	bankKeeper         BankKeeper
	distributionKeeper anteutils.DistributionKeeper
	feegrantKeeper     authante.FeegrantKeeper
	stakingKeeper      anteutils.StakingKeeper
	txFeeChecker       anteutils.TxFeeChecker
}

// NewDeductFeeDecorator returns a new DeductFeeDecorator.
func NewDeductFeeDecorator(
	ak authante.AccountKeeper,
	bk BankKeeper,
	dk anteutils.DistributionKeeper,
	fk authante.FeegrantKeeper,
	sk anteutils.StakingKeeper,
	tfc anteutils.TxFeeChecker,
) DeductFeeDecorator {
	if tfc == nil {
		tfc = checkTxFeeWithValidatorMinGasPrices
	}

	return DeductFeeDecorator{
		accountKeeper:      ak,
		bankKeeper:         bk,
		distributionKeeper: dk,
		feegrantKeeper:     fk,
		stakingKeeper:      sk,
		txFeeChecker:       tfc,
	}
}

// AnteHandle ensures that the transaction contains valid fee requirements and tries to deduct those
// from the account balance or unclaimed staking rewards, which the transaction sender might have.
func (dfd DeductFeeDecorator) AnteHandle(ctx cosmos.Context, tx cosmos.Tx, simulate bool, next cosmos.AnteHandler) (cosmos.Context, error) {
	feeTx, ok := tx.(cosmos.FeeTx)
	if !ok {
		return ctx, errorsmod.Wrap(errortypes.ErrTxDecode, "Tx must be a FeeTx")
	}

	if !simulate && ctx.BlockHeight() > 0 && feeTx.GetGas() <= 0 {
		return ctx, errorsmod.Wrap(errortypes.ErrInvalidGasLimit, "must provide positive gas")
	}

	var (
		priority int64
		err      error
	)

	fee := feeTx.GetFee()
	if !simulate {
		fee, priority, err = dfd.txFeeChecker(ctx, feeTx)
		if err != nil {
			return ctx, err
		}
	}

	feePayer := feeTx.FeePayer()
	feeGranter := feeTx.FeeGranter()

	if err = dfd.deductFee(ctx, tx, fee, feePayer, feeGranter); err != nil {
		return ctx, err
	}

	newCtx := ctx.WithPriority(priority)

	return next(newCtx, tx, simulate)
}

// deductFee checks if the fee payer has enough funds to pay for the fees and deducts them.
// If the spendable balance is not enough, it tries to claim enough staking rewards to cover the fees.
func (dfd DeductFeeDecorator) deductFee(ctx cosmos.Context, sdkTx cosmos.Tx, fees cosmos.Coins, feePayer, feeGranter cosmos.AccAddress) error {
	if fees.IsZero() {
		return nil
	}

	if addr := dfd.accountKeeper.GetModuleAddress(authtypes.FeeCollectorName); addr == nil {
		return fmt.Errorf("fee collector module account (%s) has not been set", authtypes.FeeCollectorName)
	}

	// by default, deduct fees from feePayer address
	deductFeesFrom := feePayer

	// if feegranter is set, then deduct the fee from the feegranter account.
	// this works only when feegrant is enabled.
	if feeGranter != nil {
		if dfd.feegrantKeeper == nil {
			return errortypes.ErrInvalidRequest.Wrap("fee grants are not enabled")
		}

		if !feeGranter.Equals(feePayer) {
			err := dfd.feegrantKeeper.UseGrantedFees(ctx, feeGranter, feePayer, fees, sdkTx.GetMsgs())
			if err != nil {
				return errorsmod.Wrapf(err, "%s does not not allow to pay fees for %s", feeGranter, feePayer)
			}
		}

		deductFeesFrom = feeGranter
	}

	deductFeesFromAcc := dfd.accountKeeper.GetAccount(ctx, deductFeesFrom)
	if deductFeesFromAcc == nil {
		return errortypes.ErrUnknownAddress.Wrapf("fee payer address: %s does not exist", deductFeesFrom)
	}

	// TODO mark deduct the fees
	// if err := deductFeesFromBalanceOrUnclaimedStakingRewards(ctx, dfd, deductFeesFromAcc, fees); err != nil {
	//	return fmt.Errorf("insufficient funds and failed to claim sufficient staking rewards to pay for fees: %w", err)
	// }

	events := cosmos.Events{
		cosmos.NewEvent(
			cosmos.EventTypeTx,
			cosmos.NewAttribute(cosmos.AttributeKeyFee, fees.String()),
			cosmos.NewAttribute(cosmos.AttributeKeyFeePayer, deductFeesFrom.String()),
		),
	}
	ctx.EventManager().EmitEvents(events)

	return nil
}

// deductFeesFromBalanceOrUnclaimedStakingRewards tries to deduct the fees from the account balance.
// If the account balance is not enough, it tries to claim enough staking rewards to cover the fees.
//
//nolint:unused
func deductFeesFromBalanceOrUnclaimedStakingRewards(
	ctx cosmos.Context, dfd DeductFeeDecorator, deductFeesFromAcc authtypes.AccountI, fees cosmos.Coins,
) error {
	if err := anteutils.ClaimStakingRewardsIfNecessary(
		ctx, dfd.bankKeeper, dfd.distributionKeeper, dfd.stakingKeeper, deductFeesFromAcc.GetAddress(), fees,
	); err != nil {
		return err
	}

	// TODO mark return authante.DeductFees(dfd.bankKeeper, ctx, deductFeesFromAcc, fees)
	return nil
}

// checkTxFeeWithValidatorMinGasPrices implements the default fee logic, where the minimum price per
// unit of gas is fixed and set by each validator, and the tx priority is computed from the gas price.
func checkTxFeeWithValidatorMinGasPrices(ctx cosmos.Context, feeTx cosmos.FeeTx) (cosmos.Coins, int64, error) {
	feeCoins := feeTx.GetFee()
	gas := feeTx.GetGas()

	// Ensure that the provided fees meets a minimum threshold for the validator,
	// if this is a CheckTx. This is only for local mempool purposes, and thus
	// is only ran on CheckTx.
	if ctx.IsCheckTx() {
		if err := checkFeeCoinsAgainstMinGasPrices(ctx, feeCoins, gas); err != nil {
			return nil, 0, err
		}
	}

	priority := getTxPriority(feeCoins, int64(gas)) // #nosec G701 -- gosec warning about integer overflow is not relevant here
	return feeCoins, priority, nil
}

// checkFeeCoinsAgainstMinGasPrices checks if the provided fee coins are greater than or equal to the
// required fees, that are based on the minimum gas prices and the gas. If not, it will return an error.
func checkFeeCoinsAgainstMinGasPrices(ctx cosmos.Context, feeCoins cosmos.Coins, gas uint64) error {
	minGasPrices := ctx.MinGasPrices()
	if minGasPrices.IsZero() {
		return nil
	}

	requiredFees := make(cosmos.Coins, len(minGasPrices))

	// Determine the required fees by multiplying each required minimum gas
	// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
	glDec := sdkmath.LegacyNewDec(int64(gas)) // #nosec G701 -- gosec warning about integer overflow is not relevant here
	for i, gp := range minGasPrices {
		fee := gp.Amount.Mul(glDec)
		requiredFees[i] = cosmos.NewCoin(gp.Denom, fee.Ceil().RoundInt())
	}

	if !feeCoins.IsAnyGTE(requiredFees) {
		return errorsmod.Wrapf(errortypes.ErrInsufficientFee, "insufficient fees; got: %s required: %s", feeCoins, requiredFees)
	}

	return nil
}

// getTxPriority returns a naive tx priority based on the amount of the smallest denomination of the gas price
// provided in a transaction.
// NOTE: This implementation should be used with a great consideration as it opens potential attack vectors
// where txs with multiple coins could not be prioritized as expected.
func getTxPriority(fees cosmos.Coins, gas int64) int64 {
	var priority int64
	for _, c := range fees {
		p := int64(math.MaxInt64)
		gasPrice := c.Amount.QuoRaw(gas)
		if gasPrice.IsInt64() {
			p = gasPrice.Int64()
		}
		if priority == 0 || p < priority {
			priority = p
		}
	}

	return priority
}
