package types

var (
	// ModuleName defines the module name.
	ModuleName = "denommetadata"

	// StoreKey defines the primary module store key.
	StoreKey = ModuleName

	// RouterKey is the message route for slashing.
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key.
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key.
	MemStoreKey = "mem_capability"

	// KeyLastDenomMetadataID defines key for setting last denommetadata ID.
	KeyLastDenomMetadataID = []byte{0x02}

	// KeyPrefixPeriodDenomMetadata defines prefix key for storing denommetadata.
	KeyPrefixPeriodDenomMetadata = []byte{0x03}

	// KeyIndexSeparator defines key for merging bytes.
	KeyIndexSeparator = []byte{0x07}
)
