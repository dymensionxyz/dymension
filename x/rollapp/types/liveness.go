package types

import (
	"bytes"
	"encoding/binary"
)

var (
	LivenessEventQueueKeyPrefix = []byte("LivenessEventQueue")
	LivenessEventQueueSlash     = []byte("s")
	LivenessEventQueueJail      = []byte("j")
)

func LivenessEventQueueKey(e LivenessEvent) []byte {
	v := LivenessEventQueueSlash
	if e.IsJail {
		v = LivenessEventQueueJail
	}
	return createLivenessEventQueueKey(&e.HubHeight, v, &e.RollappId)
}

func LivenessEventQueueIterKey(height *int64) []byte {
	return createLivenessEventQueueKey(height, nil, nil)
}

// can be called with no arguments to retrieve all items
// can be called with only a height, to iterate all events for a height
// otherwise must have all three arguments, for put/del ops
func createLivenessEventQueueKey(height *int64, kind []byte, rollappID *string) []byte {
	if height == nil && (0 < len(kind) || rollappID != nil) {
		panic("must provide a height")
	}
	var key []byte
	key = append(key, LivenessEventQueueKeyPrefix...)
	if height != nil {
		key = append(key, []byte("/")...)
		hBz := make([]byte, 8)
		binary.BigEndian.PutUint64(hBz, uint64(*height))
		key = append(key, hBz...)
	}
	if len(kind) != 0 {
		key = append(key, []byte("/")...)
		key = append(key, kind...)
	}
	if rollappID != nil {
		key = append(key, []byte("/")...)
		key = append(key, []byte(*rollappID)...)
	}
	return key
}

// LivenessEventQueueKeyToEvent converts store key
// Assumes the key is well-formed (contains both height and rollapp id)
func LivenessEventQueueKeyToEvent(k []byte) LivenessEvent {
	ret := LivenessEvent{}
	// key is like 'prefix/height/kind/rollapp'
	//                     i      j    l
	i := len(LivenessEventQueueKeyPrefix) + 1
	j := i + 8 + 1
	l := j + 1 + 1
	ret.HubHeight = int64(binary.BigEndian.Uint64(k[i : i+8]))
	if bytes.Equal(k[j:j+1], LivenessEventQueueJail) {
		ret.IsJail = true
	}
	ret.RollappId = string(k[l:])
	return ret
}
