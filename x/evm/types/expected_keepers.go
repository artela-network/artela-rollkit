package types

import (
	"context"
	"math/big"

	"cosmossdk.io/core/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authmodule "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingmodule "github.com/cosmos/cosmos-sdk/x/staking/types"

	feemodule "github.com/artela-network/artela-rollkit/x/fee/types"
)

type BlockGetter func() int64

type ChainIDGetter func() string

// AccountKeeper defines the expected interface for the Account module.
type AccountKeeper interface {
	GetAccount(context.Context, sdk.AccAddress) sdk.AccountI // only used for simulation
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
	GetModuleAddress(moduleName string) sdk.AccAddress
	NewAccountWithAddress(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	GetParams(ctx context.Context) (params authmodule.Params)
	SetAccount(ctx context.Context, acc sdk.AccountI)
	AddressCodec() address.Codec
	RemoveAccount(ctx context.Context, acc sdk.AccountI)
	// Methods imported from account should be defined here
}

// BankKeeper defines the expected interface for the Bank module.
type BankKeeper interface {
	SpendableCoins(context.Context, sdk.AccAddress) sdk.Coins
	// Methods imported from bank should be defined here
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	IsSendEnabledCoins(ctx context.Context, coins ...sdk.Coin) error
	SendCoins(ctx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error
}

// StakingKeeper returns the historical headers kept in store.
type StakingKeeper interface {
	GetHistoricalInfo(ctx context.Context, height int64) (stakingmodule.HistoricalInfo, error)
	GetValidatorByConsAddr(ctx context.Context, consAddr sdk.ConsAddress) (validator stakingmodule.Validator, err error)
}

// FeeKeeper
type FeeKeeper interface {
	GetBaseFee(ctx context.Context) *big.Int
	GetParams(ctx context.Context) feemodule.Params
	AddTransientGasWanted(ctx context.Context, gasWanted uint64) (uint64, error)
}

// ParamSubspace defines the expected Subspace interface for parameters.
type ParamSubspace interface {
	Get(context.Context, []byte, interface{})
	Set(context.Context, []byte, interface{})
}
