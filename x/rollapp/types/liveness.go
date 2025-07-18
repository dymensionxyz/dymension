package types

import (
	"encoding/binary"
)

var (
	LivenessEventQueueKeyPrefix = []byte("LivenessEventQueue")
	LivenessEventQueueSlash     = []byte("s")
)

func LivenessEventQueueKey(e LivenessEvent) []byte {
	kind := LivenessEventQueueSlash // there is only one kind now https://github.com/dymensionxyz/dymension/issues/1857

	ret := LivenessEventQueueIterHeightKey(e.HubHeight)
	ret = append(ret, []byte("/")...)
	ret = append(ret, kind...)
	ret = append(ret, []byte("/")...)
	ret = append(ret, e.RollappId...)
	return ret
}

// LivenessEventQueueIterHeightKey returns a key to iterate items
// If height is nil then all items
// Otherwise, only for heights greater than or equal to the passed height
func LivenessEventQueueIterHeightKey(height int64) []byte {
	ret := LivenessEventQueueKeyPrefix
	ret = append(ret, []byte("/")...)
	hBz := make([]byte, 8)
	binary.BigEndian.PutUint64(hBz, uint64(height)) //nolint:gosec
	ret = append(ret, hBz...)
	return ret
}

// LivenessEventQueueKeyToEvent converts store key to LivenessEvent
// Assumes the key is well-formed (contains both height and rollapp id)
func LivenessEventQueueKeyToEvent(k []byte) LivenessEvent {
	ret := LivenessEvent{}
	// key is like 'prefix/height/kind/rollapp'
	//                     i      j    l
	i := len(LivenessEventQueueKeyPrefix) + 1
	j := i + 8 + 1                                             // 8 is from big endian, 1 is from '/'
	l := j + 1 + 1                                             // kind is 1 character and the other 1 is from '/'
	ret.HubHeight = int64(binary.BigEndian.Uint64(k[i : i+8])) //nolint:gosec
	ret.RollappId = string(k[l:])
	return ret
}
