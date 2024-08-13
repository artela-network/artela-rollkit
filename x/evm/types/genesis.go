package types

import (
	"fmt"

	"github.com/artela-network/artela-rollkit/ethereum/types"
)

// this line is used by starport scaffolding # genesis/types/import

// DefaultIndex is the default global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		// this line is used by starport scaffolding # genesis/types/default
		Params:   DefaultParams(),
		Accounts: []GenesisAccount{},
	}
}

// NewGenesisState creates a new genesis states.
func NewGenesisState(params Params, accounts []GenesisAccount) *GenesisState {
	return &GenesisState{
		Accounts: accounts,
		Params:   params,
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// this line is used by starport scaffolding # genesis/types/validate

	seenAccounts := make(map[string]bool)
	for _, acc := range gs.Accounts {
		if seenAccounts[acc.Address] {
			return fmt.Errorf("duplicated genesis account %s", acc.Address)
		}
		if err := acc.Validate(); err != nil {
			return fmt.Errorf("invalid genesis account %s: %w", acc.Address, err)
		}
		seenAccounts[acc.Address] = true
	}
	return gs.Params.Validate()
}

// ----------------------------------------------------------------------------
// 							 Genesis Account
// ----------------------------------------------------------------------------

// Validate performs a basic validation of a GenesisAccount fields.
func (ga GenesisAccount) Validate() error {
	if err := types.ValidateAddress(ga.Address); err != nil {
		return err
	}
	return ga.Storage.Validate()
}
