package types

import commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"

type RollappPacketListFilter struct {
	Prefixes []Prefix
}

type Prefix struct {
	Start []byte
	End   []byte
}

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
	}
}

func ByRollappIDByStatus(rollappID string, status ...commontypes.Status) RollappPacketListFilter {
	prefixes := make([]Prefix, len(status))
	for i, s := range status {
		prefixes[i] = Prefix{Start: commontypes.RollappPacketByStatusByRollappIDPrefix(s, rollappID)}
	}
	return RollappPacketListFilter{
		Prefixes: prefixes,
	}
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
		Prefixes: prefixes,
	}
}
