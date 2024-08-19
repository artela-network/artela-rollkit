package cmd

import (
	"cosmossdk.io/math"
	cmtcfg "github.com/cometbft/cometbft/config"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/artela-network/artela-rollkit/app"
	"github.com/artela-network/artela-rollkit/ethereum/server/config"
	artelatypes "github.com/artela-network/artela-rollkit/ethereum/types"
)

func initSDKConfig() {
	// Set prefixes
	accountPubKeyPrefix := app.AccountAddressPrefix + "pub"
	validatorAddressPrefix := app.AccountAddressPrefix + "valoper"
	validatorPubKeyPrefix := app.AccountAddressPrefix + "valoperpub"
	consNodeAddressPrefix := app.AccountAddressPrefix + "valcons"
	consNodePubKeyPrefix := app.AccountAddressPrefix + "valconspub"

	// Set and seal config
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(app.AccountAddressPrefix, accountPubKeyPrefix)
	config.SetBech32PrefixForValidator(validatorAddressPrefix, validatorPubKeyPrefix)
	config.SetBech32PrefixForConsensusNode(consNodeAddressPrefix, consNodePubKeyPrefix)
	config.SetCoinType(artelatypes.Bip44CoinType)
	config.SetPurpose(sdk.Purpose)                        // Shared
	config.SetFullFundraiserPath(artelatypes.BIP44HDPath) //nolint: staticcheck
	config.Seal()

	registerDenoms()
}

// RegisterDenoms registers the base and display denominations to the SDK.
func registerDenoms() {
	if err := sdk.RegisterDenom(app.DisplayDenom, math.LegacyOneDec()); err != nil {
		panic(err)
	}

	if err := sdk.RegisterDenom(app.BaseDenom, math.LegacyNewDecWithPrec(1, artelatypes.BaseDenomUnit)); err != nil {
		panic(err)
	}
}

// initCometBFTConfig helps to override default CometBFT Config values.
// return cmtcfg.DefaultConfig if no custom configuration is required for the application.
func initCometBFTConfig() *cmtcfg.Config {
	cfg := cmtcfg.DefaultConfig()

	// these values put a higher strain on node memory
	// cfg.P2P.MaxNumInboundPeers = 100
	// cfg.P2P.MaxNumOutboundPeers = 40

	return cfg
}

// initAppConfig helps to override default appConfig template and configs.
// return "", nil if no custom configuration is required for the application.
func initAppConfig() (string, interface{}) {
	return config.AppConfig(artelatypes.AttoArtela)
}
