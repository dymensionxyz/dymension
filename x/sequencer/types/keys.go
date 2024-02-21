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

// SequencersByRollappKey returns the store key to retrieve a SequencersByRollapp from the index fields
func SequencerKey(sequencerAddress string) []byte {
	sequencerAddrBytes := []byte(sequencerAddress)
	return []byte(fmt.Sprintf("%s%s%s", SequencersKeyPrefix, KeySeparator, sequencerAddrBytes))
}

func SequencersKey() []byte {
	return []byte(SequencersKeyPrefix)
}

// SequencersByRollappKey returns the store key to retrieve a SequencersByRollapp from the index fields
func SequencerByRollappByStatusKey(rollappId, seqAddr string, status OperatingStatus) []byte {
	return append(SequencersByRollappByStatusKey(rollappId, status), SequencerKey(seqAddr)...)
}

// SequencersByRollappKey returns the store key to retrieve a SequencersByRollapp from the index fields
func SequencersByRollappKey(rollappId string) []byte {
	rollappIdBytes := []byte(rollappId)
	return []byte(fmt.Sprintf("%s%s%s", SequencersByRollappKeyPrefix, KeySeparator, rollappIdBytes))
}

// SequencersByRollappKey returns the store key to retrieve a SequencersByRollapp from the index fields
func SequencersByRollappByStatusKey(rollappId string, status OperatingStatus) []byte {
	rollappIdBytes := []byte(rollappId)

	// Get the relevant key prefix based on the packet status
	var prefix []byte
	switch status {
	case Bonded:
	case Proposer:
		prefix = BondedSequencersKeyPrefix
	case Unbonded:
		prefix = UnbondedSequencersKeyPrefix
	case Unbonding:
		prefix = UnbondingSequencersKeyPrefix
	default:
		panic(fmt.Sprintf("invalid sequencer status: %s", status.String()))
	}

	return []byte(fmt.Sprintf("%s%s%s%s%s", SequencersByRollappKeyPrefix, KeySeparator, rollappIdBytes, KeySeparator, prefix))
}

func UnbondingQueueByTimeKey(endTime time.Time) []byte {
	return append(UnbondingQueueKey, sdk.FormatTimeBytes(endTime)...)
}

func UnbondingSequencerKey(sequencerAddress string, endTime time.Time) []byte {
	key := append(UnbondingQueueKey, sdk.FormatTimeBytes(endTime)...)
	key = append(key, KeySeparator...)
	key = append(key, []byte(sequencerAddress)...)
	return key
}
