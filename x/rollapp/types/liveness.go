package types

import (
	"bytes"
	"encoding/binary"
	"time"
)

// LivenessCheckParams are the params needed to do a liveness check
// This is a utility struct to make dealing with these params which are commonly used together
// easier.
// TODO: do I really need this?
type LivenessCheckParams struct {
	// HubExpectedBlockTime	is how long it typically takes to produce a Hub block (e.g. 6 secs). Used to calculate length of time based on height.
	HubExpectedBlockTime time.Duration
	// SlashTime is the time a sequencer has to post a block, before he will be slashed
	SlashTime time.Duration
	// SlashInterval is the min gap between a sequence of slashes if the sequencer continues to be down
	SlashInterval time.Duration
	// JailTime	is the time a sequencer can be down after which he will be jailed rather than slashed
	JailTime time.Duration
}

var (
	LivenessEventQueueKeyPrefix = []byte("LivenessEventQueue/")
	LivenessEventQueueSlash     = []byte("s")
	LivenessEventQueueJail      = []byte("j")
)

func LivenessEventQueueKey(height int64, rollappID string) []byte {
	return createLivenessEventQueueKey(&height, &rollappID)
}

func LivenessEventQueueIterKey(height *int64) []byte {
	return createLivenessEventQueueKey(height, nil)
}

func createLivenessEventQueueKey(height *int64, rollappID *string) []byte {
	if rollappID != nil && height == nil {
		panic("must provide height with rollapp id")
	}
	var key []byte
	key = append(key, LivenessEventQueueKeyPrefix...)
	if height != nil {
		hBz := make([]byte, 8)
		binary.BigEndian.PutUint64(hBz, uint64(*height))
		key = append(key, hBz...)
	}
	if rollappID != nil {
		key = append(key, []byte("/")...)
		key = append(key, []byte(*rollappID)...)
	}
	return key
}

// LivenessEventQueueItemToEvent converts store key and value to an event
// Assumes the key is well-formed (contains both height and rollapp id)
func LivenessEventQueueItemToEvent(k, v []byte) LivenessEvent {
	ret := LivenessEvent{}
	ret.IsJail = bytes.Equal(v, LivenessEventQueueJail)
	// key is like 'foo/height/rollapp'
	//                  i      j
	i := len(LivenessEventQueueKeyPrefix) + 1
	j := i + 8 + 1
	ret.HubHeight = int64(binary.BigEndian.Uint64(k[i : j-1]))
	ret.RollappId = string(k[j:])
	return ret
}
