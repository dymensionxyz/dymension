package types

import (
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
)

type RollappPacketListFilter struct {
	Prefixes   []Prefix
	FilterFunc func(packet commontypes.RollappPacket) bool
}

type Prefix struct {
	Start []byte
	End   []byte
}

var bypassFilter = func(packet commontypes.RollappPacket) bool { return true }

func PendingByRollappIDByMaxHeight(
	rollappID string,
	maxProofHeight uint64,
) RollappPacketListFilter {
	status := commontypes.Status_PENDING
	return RollappPacketListFilter{
		Prefixes: []Prefix{
			{
				Start: commontypes.RollappPacketByStatusByRollappIDByProofHeightPrefix(rollappID, status, 0),
				End:   commontypes.RollappPacketByStatusByRollappIDByProofHeightPrefix(rollappID, status, maxProofHeight+1), // inclusive end
			},
		},
		FilterFunc: bypassFilter,
	}
}

func ByRollappIDByStatus(rollappID string, status ...commontypes.Status) RollappPacketListFilter {
	prefixes := make([]Prefix, len(status))
	for i, s := range status {
		prefixes[i] = Prefix{Start: commontypes.RollappPacketByStatusByRollappIDPrefix(s, rollappID)}
	}
	return RollappPacketListFilter{
		Prefixes:   prefixes,
		FilterFunc: bypassFilter,
	}
}

func ByRollappIDByTypeByStatus(rollappID string, packetType commontypes.Type, status ...commontypes.Status) RollappPacketListFilter {
	filter := ByRollappIDByStatus(rollappID, status...)
	if packetType != commontypes.Type_UNDEFINED {
		filter.FilterFunc = func(packet commontypes.RollappPacket) bool {
			return packet.Type == packetType
		}
	}
	return filter
}

func ByRollappID(rollappID string) RollappPacketListFilter {
	return ByRollappIDByStatus(rollappID,
		commontypes.Status_PENDING,
		commontypes.Status_FINALIZED,
		commontypes.Status_REVERTED,
	)
}

func ByStatus(status ...commontypes.Status) RollappPacketListFilter {
	prefixes := make([]Prefix, len(status))
	for i, s := range status {
		prefixes[i] = Prefix{Start: commontypes.RollappPacketByStatusPrefix(s)}
	}
	return RollappPacketListFilter{
		Prefixes:   prefixes,
		FilterFunc: bypassFilter,
	}
}

func ByTypeByStatus(packetType commontypes.Type, status ...commontypes.Status) RollappPacketListFilter {
	filter := ByStatus(status...)
	if packetType != commontypes.Type_UNDEFINED {
		filter.FilterFunc = func(packet commontypes.RollappPacket) bool {
			return packet.Type == packetType
		}
	}
	return filter
}
