package types

const (
	// ModuleName defines the module name
	ModuleName = "irc"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_irc"
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
