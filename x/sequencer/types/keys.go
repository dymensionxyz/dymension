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

/*
// GetUBDKey creates the key for an unbonding delegation by delegator and validator addr
// VALUE: staking/UnbondingDelegation
func GetUBDKey(delAddr sdk.AccAddress, valAddr sdk.ValAddress) []byte {
	return append(GetUBDsKey(delAddr.Bytes()), address.MustLengthPrefix(valAddr)...)
}

// GetUBDByValIndexKey creates the index-key for an unbonding delegation, stored by validator-index
// VALUE: none (key rearrangement used)
func GetUBDByValIndexKey(delAddr sdk.AccAddress, valAddr sdk.ValAddress) []byte {
	return append(GetUBDsByValIndexKey(valAddr), address.MustLengthPrefix(delAddr)...)
}

// GetUBDKeyFromValIndexKey rearranges the ValIndexKey to get the UBDKey
func GetUBDKeyFromValIndexKey(indexKey []byte) []byte {
	kv.AssertKeyAtLeastLength(indexKey, 2)
	addrs := indexKey[1:] // remove prefix bytes

	valAddrLen := addrs[0]
	kv.AssertKeyAtLeastLength(addrs, 2+int(valAddrLen))
	valAddr := addrs[1 : 1+valAddrLen]
	kv.AssertKeyAtLeastLength(addrs, 3+int(valAddrLen))
	delAddr := addrs[valAddrLen+2:]

	return GetUBDKey(delAddr, valAddr)
}

// GetUBDsKey creates the prefix for all unbonding delegations from a delegator
func GetUBDsKey(delAddr sdk.AccAddress) []byte {
	return append(UnbondingDelegationKey, address.MustLengthPrefix(delAddr)...)
}

// GetUBDsByValIndexKey creates the prefix keyspace for the indexes of unbonding delegations for a validator
func GetUBDsByValIndexKey(valAddr sdk.ValAddress) []byte {
	return append(UnbondingDelegationByValIndexKey, address.MustLengthPrefix(valAddr)...)
}

// GetUnbondingDelegationTimeKey creates the prefix for all unbonding delegations from a delegator
func GetUnbondingDelegationTimeKey(timestamp time.Time) []byte {
	bz := sdk.FormatTimeBytes(timestamp)
	return append(UnbondingQueueKey, bz...)
}
*/
