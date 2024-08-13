package types

const (
	// ModuleName defines the module name
	ModuleName = "fee"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_fee"

	// TransientStoreKey defines the transient store key
	TransientStoreKey = "transient_fee"
)

// prefix bytes for the fee persistent store
const (
	prefixBlockGasWanted    = iota + 1
	deprecatedPrefixBaseFee // unused
)

const (
	prefixTransientBlockGasUsed = iota + 1
)

// KVStore key prefixes
var (
	KeyPrefixBlockGasWanted = []byte{prefixBlockGasWanted}
)

// Transient Store key prefixes
var (
	KeyPrefixTransientBlockGasWanted = []byte{prefixTransientBlockGasUsed}
)

// fee module events
const (
	EventTypeFee = "fee"

	AttributeKeyBaseFee = "base_fee"
)

var (
	ParamsKey = []byte("p_fee")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
