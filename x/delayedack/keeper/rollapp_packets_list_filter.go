package keeper

import commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"

type rollappPacketListFilter struct {
	prefixBytes      []byte
	filter           func(val commontypes.RollappPacket) bool
	stopOnFirstMatch bool
}

func AllRollappPackets() rollappPacketListFilter {
	return rollappPacketListFilter{}
}

func ByRollappIDAndStatus(rollappID string, status commontypes.Status) rollappPacketListFilter {
	return rollappPacketListFilter{
		prefixBytes: commontypes.RollappIDAndStatusPacketPrefix(rollappID, status),
	}
}

func ByRollappIDAndStatusAndMaxHeight(
	rollappID string,
	status commontypes.Status,
	maxProofHeight uint64,
	stopOnFirstMatch bool,
) rollappPacketListFilter {
	return rollappPacketListFilter{
		prefixBytes: commontypes.RollappIDAndStatusPacketPrefix(rollappID, status),
		filter: func(val commontypes.RollappPacket) bool {
			return val.ProofHeight <= maxProofHeight
		},
		stopOnFirstMatch: stopOnFirstMatch,
	}
}

func ByRollappID(rollappID string) rollappPacketListFilter {
	return rollappPacketListFilter{
		prefixBytes: []byte(rollappID),
	}
}

func ByStatus(status commontypes.Status) rollappPacketListFilter {
	return rollappPacketListFilter{
		filter: func(val commontypes.RollappPacket) bool {
			return val.Status == status
		},
	}
}

func ByNotStatus(notStatus commontypes.Status) rollappPacketListFilter {
	return rollappPacketListFilter{
		filter: func(val commontypes.RollappPacket) bool {
			return val.Status != notStatus
		},
	}
}
