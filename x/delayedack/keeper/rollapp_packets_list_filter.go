package keeper

import commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"

type rollappPacketListFilter struct {
	prefixes []prefix
}

type prefix struct {
	start []byte
	end   []byte
}

func PendingByRollappIDByMaxHeight(
	rollappID string,
	maxProofHeight uint64,
) rollappPacketListFilter {
	status := commontypes.Status_PENDING
	return rollappPacketListFilter{
		prefixes: []prefix{
			{
				start: commontypes.RollappPacketByStatusByRollappIDByProofHeightPrefix(rollappID, status, 0),
				end:   commontypes.RollappPacketByStatusByRollappIDByProofHeightPrefix(rollappID, status, maxProofHeight+1), // inclusive end
			},
		},
	}
}

func ByRollappIDByStatus(rollappID string, status ...commontypes.Status) rollappPacketListFilter {
	prefixes := make([]prefix, len(status))
	for i, s := range status {
		prefixes[i] = prefix{start: commontypes.RollappPacketByStatusByRollappIDPrefix(s, rollappID)}
	}
	return rollappPacketListFilter{
		prefixes: prefixes,
	}
}

func ByRollappID(rollappID string) rollappPacketListFilter {
	return ByRollappIDByStatus(rollappID,
		commontypes.Status_PENDING,
		commontypes.Status_FINALIZED,
		commontypes.Status_REVERTED,
	)
}

func ByStatus(status ...commontypes.Status) rollappPacketListFilter {
	prefixes := make([]prefix, len(status))
	for i, s := range status {
		prefixes[i] = prefix{start: commontypes.RollappPacketByStatusPrefix(s)}
	}
	return rollappPacketListFilter{
		prefixes: prefixes,
	}
}
