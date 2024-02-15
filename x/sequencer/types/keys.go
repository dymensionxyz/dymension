package types

import (
	"encoding/binary"
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

// TODO: change keys to bytes
var (
	// SequencerKeyPrefix is the prefix to retrieve all Sequencer
	SequencerKeyPrefix = "Sequencer/value/"
	// SequencersByRollappKeyPrefix is the prefix to retrieve all SequencersByRollapp
	SequencersByRollappKeyPrefix = "SequencersByRollapp/value/"

	UnbondingSequencerKey          = []byte{0x32} // key for an unbonding-delegation
	UnbondingSequencerByRollappKey = []byte{0x33} // prefix for each key for an unbonding-delegation, by validator operator

	UnbondingQueueKey = []byte{0x41} // prefix for the timestamps in unbonding queue
)

// SequencerKey returns the store key to retrieve a Sequencer from the index fields
func SequencerKey(
	sequencerAddress string,
) []byte {
	var key []byte

	sequencerAddressBytes := []byte(sequencerAddress)
	key = append(key, sequencerAddressBytes...)
	key = append(key, []byte("/")...)

	return key
}

// SequencersByRollappKey returns the store key to retrieve a SequencersByRollapp from the index fields
func SequencersByRollappKey(
	rollappId string,
) []byte {
	var key []byte

	rollappIdBytes := []byte(rollappId)
	key = append(key, rollappIdBytes...)
	key = append(key, []byte("/")...)

	return key
}

func KeyPrefix(p string) []byte {
	return []byte(p)
}
