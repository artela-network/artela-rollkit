package types

import (
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"
	"github.com/ethereum/go-ethereum/common"
)

// ----------------------------------------------------------------------------
// 							       State
// ----------------------------------------------------------------------------

// Validate performs a basic validation of the State fields.
// NOTE: states value can be empty
// State represents a single Storage key value pair item.
func (s State) Validate() error {
	if strings.TrimSpace(s.Key) == "" {
		return errorsmod.Wrap(ErrInvalidState, "states key hash cannot be blank")
	}

	return nil
}

// NewState creates a new State instance
func NewState(key, value common.Hash) State {
	return State{
		Key:   key.String(),
		Value: value.String(),
	}
}

// ----------------------------------------------------------------------------
// 						   State Array - Storage
// ----------------------------------------------------------------------------

// Storage represents the account Storage map as a slice of single key value
// State pairs.
type Storage []State

// Validate performs a basic validation of the Storage fields.
func (s Storage) Validate() error {
	seenStorage := make(map[string]bool)
	for i, state := range s {
		if seenStorage[state.Key] {
			return errorsmod.Wrapf(ErrInvalidState, "duplicate states key %d: %s", i, state.Key)
		}

		if err := state.Validate(); err != nil {
			return err
		}

		seenStorage[state.Key] = true
	}
	return nil
}

// String implements the stringer interface
func (s Storage) String() string {
	var str string
	for _, state := range s {
		str += fmt.Sprintf("%s\n", state.String())
	}
	return str
}

// Copy returns a copy of storage.
func (s Storage) Copy() Storage {
	cpy := make(Storage, len(s))
	copy(cpy, s)
	return cpy
}
