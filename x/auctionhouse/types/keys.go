package types

const (
	// ModuleName defines the module name
	ModuleName = "auctionhouse"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_auctionhouse"
)

var (
	// AuctionKeyPrefix is the prefix for auction keys
	AuctionKeyPrefix = []byte{0x01}

	// PurchaseKeyPrefix is the prefix for purchase keys
	PurchaseKeyPrefix = []byte{0x02}

	// NextAuctionIDKey is the key for the next auction ID
	NextAuctionIDKey = []byte{0x03}

	// ParamsKey is the key for module parameters
	ParamsKey = []byte{0x04}
)
