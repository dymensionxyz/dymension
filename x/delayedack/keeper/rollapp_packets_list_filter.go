package keeper

import commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"

type filterFunc func(val commontypes.RollappPacket) bool

type rollappPacketListFilter struct {
	prefixes        [][]byte
	filter          filterFunc
	breakOnMismatch bool
}

func AllRollappPackets() rollappPacketListFilter {
	return rollappPacketListFilter{prefixes: [][]byte{{}}}
}

func ByRollappIDAndStatus(rollappID string, status ...commontypes.Status) rollappPacketListFilter {
	return rollappPacketListFilter{
		prefixes: buildPrefixes(rollappID, status),
	}
}
func ByRollappIDAndStatusAndMaxHeight(
	rollappID string,
	maxProofHeight uint64,
	breakOnMismatch bool,
	status ...commontypes.Status,
) rollappPacketListFilter {
	filter := ByRollappIDAndStatus(rollappID, status...)
	filter.breakOnMismatch = breakOnMismatch
	filter.filter = func(val commontypes.RollappPacket) bool {
		return val.ProofHeight <= maxProofHeight // TODO: move into separate modifier
	}
	return filter
}

func ByRollappID(rollappID string) rollappPacketListFilter {
	return ByRollappIDAndStatus(rollappID,
		commontypes.Status_PENDING,
		commontypes.Status_FINALIZED,
		commontypes.Status_REVERTED,
	)
}

func ByStatus(status ...commontypes.Status) rollappPacketListFilter {
	return ByRollappIDAndStatus("", status...)
}

func buildPrefixes(rollappID string, status []commontypes.Status) [][]byte {
	prefixes := make([][]byte, len(status))
	for i, s := range status {
		packet := &commontypes.RollappPacket{
			RollappId: rollappID,
			Status:    s,
		}
		prefixes[i] = commontypes.RollappPacketStatusAndRollappIDKey(packet)
	}
	return prefixes
}
