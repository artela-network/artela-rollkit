package app

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	crypto "github.com/artela-network/artela-rollkit/ethereum/crypto/codec"
	"github.com/artela-network/artela-rollkit/ethereum/types"
)

// RegisterInterfaces registers Interfaces from types, crypto, and SDK std.
func RegisterInterfaces(interfaceRegistry codectypes.InterfaceRegistry) {
	crypto.RegisterInterfaces(interfaceRegistry)
	types.RegisterInterfaces(interfaceRegistry)
}
