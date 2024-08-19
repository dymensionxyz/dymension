package types

import (
	"encoding/binary"
	fmt "fmt"
	"time"

	"github.com/dymensionxyz/dymension/v3/utils"
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
)

var (
	// KeySeparator defines the separator for keys
	KeySeparator = "/"

	// SequencersKeyPrefix is the prefix to retrieve all Sequencers by their address
	SequencersKeyPrefix = []byte{0x00} // prefix/seqAddr

	// SequencersByRollappKeyPrefix is the prefix to retrieve all SequencersByRollapp
	SequencersByRollappKeyPrefix = []byte{0x01} // prefix/rollappId

	// ProposerKeyPrefix is the prefix to retrieve the proposer for a rollapp
	// This key is set when the rotation handshake is completed
	ProposerKeyPrefix = []byte{0x02} // prefix/rollappId
	// NextProposerKeyPrefix is the prefix to retrieve the next proposer for a rollapp
	// This key is set only when rotation handshake is started
	// It will be cleared after the rotation is completed
	NextProposerKeyPrefix = []byte{0x03} // prefix/rollappId

	// Prefixes for the different sequencer statuses
	BondedSequencersKeyPrefix    = []byte{0xa1}
	UnbondedSequencersKeyPrefix  = []byte{0xa2}
	UnbondingSequencersKeyPrefix = []byte{0xa3}

	UnbondingQueueKey      = []byte{0x41} // prefix for the timestamps in unbonding queue
	NoticePeriodQueueKey   = []byte{0x42} // prefix for the timestamps in notice period queue
	DecreasingBondQueueKey = []byte{0x43} // prefix for the timestamps in decreasing bond queue
)

/* --------------------- specific sequencer address keys -------------------- */
func SequencerKey(sequencerAddress string) []byte {
	sequencerAddrBytes := []byte(sequencerAddress)
	return []byte(fmt.Sprintf("%s%s%s", SequencersKeyPrefix, KeySeparator, sequencerAddrBytes))
}

// SequencerByRollappByStatusKey returns the store key to retrieve a SequencersByRollapp from the index fields
func SequencerByRollappByStatusKey(rollappId, seqAddr string, status OperatingStatus) []byte {
	return append(SequencersByRollappByStatusKey(rollappId, status), []byte(seqAddr)...)
}

/* ------------------------- multiple sequencers keys ------------------------ */
func SequencersKey() []byte {
	return SequencersKeyPrefix
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

/* --------------------------  queues keys -------------------------- */

func UnbondingQueueByTimeKey(endTime time.Time) []byte {
	return utils.EncodeTimeToKey(UnbondingQueueKey, endTime)
}

func NoticePeriodQueueByTimeKey(endTime time.Time) []byte {
	return utils.EncodeTimeToKey(NoticePeriodQueueKey, endTime)
}

func UnbondingSequencerKey(sequencerAddress string, endTime time.Time) []byte {
	key := UnbondingQueueByTimeKey(endTime)
	key = append(key, KeySeparator...)
	key = append(key, []byte(sequencerAddress)...)
	return key
}

func NoticePeriodSequencerKey(sequencerAddress string, endTime time.Time) []byte {
	key := NoticePeriodQueueByTimeKey(endTime)
	key = append(key, KeySeparator...)
	key = append(key, []byte(sequencerAddress)...)
	return key
}

/* -------------------------- decreasing bond queue keys -------------------------- */
func DecreasingBondQueueByTimeKey(endTime time.Time) []byte {
	return utils.EncodeTimeToKey(DecreasingBondQueueKey, endTime)
}

func GetDecreasingBondQueueKey(sequencerAddress string, endTime time.Time) []byte {
	key := DecreasingBondQueueByTimeKey(endTime)
	key = append(key, KeySeparator...)
	key = append(key, []byte(sequencerAddress)...)
	return key
}

/* --------------------- active and next sequencer keys --------------------- */
func ProposerByRollappKey(rollappId string) []byte {
	return []byte(fmt.Sprintf("%s%s%s", ProposerKeyPrefix, KeySeparator, []byte(rollappId)))
}

func NextProposerByRollappKey(rollappId string) []byte {
	return []byte(fmt.Sprintf("%s%s%s", NextProposerKeyPrefix, KeySeparator, []byte(rollappId)))
}
