package types

var (
	// ModuleName defines the module name.
	ModuleName = "streamer"

	// StoreKey defines the primary module store key.
	StoreKey = ModuleName

	// RouterKey is the message route for slashing.
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key.
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key.
	MemStoreKey = "mem_capability"

	// KeyPrefixTimestamp defines prefix key for timestamp iterator key.
	KeyPrefixTimestamp = []byte{0x01}

	// KeyLastStreamID defines key for setting last stream ID.
	KeyLastStreamID = []byte{0x02}

	// KeyPrefixPeriodStream defines prefix key for storing streams.
	KeyPrefixPeriodStream = []byte{0x03}

	// KeyPrefixStreams defines prefix key for storing reference key for all streams.
	KeyPrefixStreams = []byte{0x04}

	// KeyPrefixUpcomingStreams defines prefix key for storing reference key for upcoming streams.
	KeyPrefixUpcomingStreams = []byte{0x04, 0x00}

	// KeyPrefixActiveStreams defines prefix key for storing reference key for active streams.
	KeyPrefixActiveStreams = []byte{0x04, 0x01}

	// KeyPrefixFinishedStreams defines prefix key for storing reference key for finished streams.
	KeyPrefixFinishedStreams = []byte{0x04, 0x02}

	// KeyIndexSeparator defines key for merging bytes.
	KeyIndexSeparator = []byte{0x07}

	// KeyPrefixEpochPointers defines a prefix key holding EpochPointer objects.
	KeyPrefixEpochPointers = []byte{0x08}
)
