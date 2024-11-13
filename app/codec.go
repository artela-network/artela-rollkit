package app

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"

	crypto "github.com/artela-network/artela-rollkit/ethereum/crypto/codec"
	"github.com/artela-network/artela-rollkit/ethereum/types"
)

// RegisterInterfaces registers Interfaces from types, crypto, and SDK std.
func RegisterInterfaces(interfaceRegistry codectypes.InterfaceRegistry) {
	crypto.RegisterInterfaces(interfaceRegistry)
	types.RegisterInterfaces(interfaceRegistry)
}

// RegisterLegacyAminoCodec registers Interfaces from types, crypto, and SDK std.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	sdktypes.RegisterLegacyAminoCodec(cdc)
	crypto.RegisterCrypto(cdc)
	codec.RegisterEvidences(cdc)
}
