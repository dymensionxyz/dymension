package types

import (
	"encoding/binary"
	fmt "fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ binary.ByteOrder

const (
	// ModuleName defines the module name
	ModuleName = "sequencer"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_sequencer"
)

var (
	// KeySeparator defines the separator for keys
	KeySeparator = "/"

	// SequencersKeyPrefix is the prefix to retrieve all Sequencers by their address
	SequencersKeyPrefix = []byte{0x00} // prefix/seqAddr

	// SequencersByRollappKeyPrefix is the prefix to retrieve all SequencersByRollapp
	SequencersByRollappKeyPrefix = []byte{0x01} // prefix/rollappId
	BondedSequencersKeyPrefix    = []byte{0xa1}
	UnbondedSequencersKeyPrefix  = []byte{0xa2}
	UnbondingSequencersKeyPrefix = []byte{0xa3}

	UnbondingQueueKey = []byte{0x41} // prefix for the timestamps in unbonding queue
)

/* --------------------- specific sequencer address keys -------------------- */
func SequencerKey(sequencerAddress string) []byte {
	sequencerAddrBytes := []byte(sequencerAddress)
	return []byte(fmt.Sprintf("%s%s%s", SequencersKeyPrefix, KeySeparator, sequencerAddrBytes))
}

// SequencersByRollappKey returns the store key to retrieve a SequencersByRollapp from the index fields
func SequencerByRollappByStatusKey(rollappId, seqAddr string, status OperatingStatus) []byte {
	return append(SequencersByRollappByStatusKey(rollappId, status), []byte(seqAddr)...)
}

/* ------------------------- multiple sequencers keys ------------------------ */
func SequencersKey() []byte {
	return []byte(SequencersKeyPrefix)
}

// SequencersByRollappKey returns the store key to retrieve a SequencersByRollapp from the index fields
func SequencersByRollappKey(rollappId string) []byte {
	rollappIdBytes := []byte(rollappId)
	return []byte(fmt.Sprintf("%s%s%s", SequencersByRollappKeyPrefix, KeySeparator, rollappIdBytes))
}

// SequencersByRollappByStatusKey returns the store key to retrieve a SequencersByRollappByStatus from the index fields
func SequencersByRollappByStatusKey(rollappId string, status OperatingStatus) []byte {
	// Get the relevant key prefix based on the packet status
	var prefix []byte
	switch status {
	case Bonded:
		prefix = BondedSequencersKeyPrefix
	case Unbonded:
		prefix = UnbondedSequencersKeyPrefix
	case Unbonding:
		prefix = UnbondingSequencersKeyPrefix
	}

	return []byte(fmt.Sprintf("%s%s%s", SequencersByRollappKey(rollappId), KeySeparator, prefix))
}

/* -------------------------- unbonding queue keys -------------------------- */
func UnbondingQueueByTimeKey(endTime time.Time) []byte {
	timeBz := sdk.FormatTimeBytes(endTime)
	prefixL := len(UnbondingQueueKey)

	bz := make([]byte, prefixL+len(timeBz))

	// copy the prefix
	copy(bz[:prefixL], UnbondingQueueKey)
	// copy the encoded time bytes
	copy(bz[prefixL:prefixL+len(timeBz)], timeBz)

	return bz
}

func UnbondingSequencerKey(sequencerAddress string, endTime time.Time) []byte {
	key := UnbondingQueueByTimeKey(endTime)
	key = append(key, KeySeparator...)
	key = append(key, []byte(sequencerAddress)...)
	return key
}
