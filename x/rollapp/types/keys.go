package types

import "cosmossdk.io/collections"

const (
	// ModuleName defines the module name
	ModuleName = "rollapp"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_rollapp"
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

const (
	ObsoleteDRSVersionsKeyPrefix = "obsoleteDRSVersions/value/"
	// KeyRegisteredDenomPrefix is the prefix to retrieve all RegisteredDenom
	KeyRegisteredDenomPrefix = "RegisteredDenom/value/"
)

var SeqToUnfinalizedHeightKeyPrefix = collections.NewPrefix("seqToFinalizeHeight/")
