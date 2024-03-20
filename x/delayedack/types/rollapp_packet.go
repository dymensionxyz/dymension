package types

import rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"

func GetLatestFinalizedHeightFromStateInfo(stateInfo *rollapptypes.StateInfo) uint64 {
	return stateInfo.StartHeight + stateInfo.NumBlocks - 1
}
