package types

const (
	// ModuleName defines the module name
	ModuleName = "aspect"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_aspect"
)

var (
	ParamsKey = []byte("p_aspect")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
