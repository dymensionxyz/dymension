package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// RegisterInvariants registers the bank module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "rollapp-finalized-packet", RollappFinalizedPackets(k))
	ir.RegisterRoute(types.ModuleName, "rollapp-reverted-packet", RollappRevertedPackets(k))
}

// AllInvariants runs all invariants of the X/bank module.
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		res, stop := RollappFinalizedPackets(k)(ctx)
		if stop {
			return res, stop
		}
		res, stop = RollappRevertedPackets(k)(ctx)
		if stop {
			return res, stop
		}
		return "", false
	}
}

// RollappFinalizedPackets checks that all rollapp packets stored for a rollapp finalized height are also finalized
func RollappFinalizedPackets(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			broken bool
			msg    string
		)

		packets := k.GetAllRollappPackets(ctx)

		for _, packet := range packets {
			latestFinalizedStateIndex, found := k.rollappKeeper.GetLatestFinalizedStateIndex(ctx, packet.RollappId)
			if !found {
				msg += fmt.Sprintf("unable to find latest finalized state index for rollapp %s\n", packet.RollappId)
				broken = true
			}
			latestFinalizedStateInfo, found := k.rollappKeeper.GetStateInfo(ctx, packet.RollappId, latestFinalizedStateIndex.Index)
			if !found {
				msg += fmt.Sprintf("unable to find latest finalized state info for rollapp %s\n", packet.RollappId)
				broken = true
			}
			latestFinalizedHeight := latestFinalizedStateInfo.StartHeight + latestFinalizedStateInfo.NumBlocks - 1
			if packet.ProofHeight <= latestFinalizedHeight && packet.Status != commontypes.Status_FINALIZED {
				msg += fmt.Sprintf("rollapp packet for height %d from rollapp %s should be in finalized status, but is in %s status\n", packet.ProofHeight, packet.RollappId, packet.Status)
				broken = true
			}
		}
		return msg, broken
	}
}

// RollappRevertedPackets checks that all rollapp packets stored for a rollapp reverted height are also reverted
func RollappRevertedPackets(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			broken bool
			msg    string
		)

		rollapps := k.rollappKeeper.GetAllRollapps(ctx)

		for _, rollapp := range rollapps {
			latestFinalizedStateIndex, found := k.rollappKeeper.GetLatestFinalizedStateIndex(ctx, rollapp.RollappId)
			if !found {
				continue
			}

			stateInfoIndex := latestFinalizedStateIndex.Index + 1
			//Checking that all packets after the latest finalized height, that belong to a reverted state info, are also in reverted state
			for {
				stateInfoToCheck, found := k.rollappKeeper.GetStateInfo(ctx, rollapp.RollappId, stateInfoIndex)
				if found {
					if stateInfoToCheck.Status == commontypes.Status_REVERTED {
						//TODO (srene) add GetRollappPacketByRollap to be more efficient
						for _, packet := range k.GetAllRollappPackets(ctx) {
							if packet.RollappId == rollapp.RollappId {
								if packet.ProofHeight >= stateInfoToCheck.StartHeight && packet.ProofHeight < stateInfoToCheck.StartHeight+stateInfoToCheck.NumBlocks && packet.Status != commontypes.Status_REVERTED {
									msg += fmt.Sprintf("rollapp packet for height %d from rollapp %s should be in reverted status, but is in %s status\n", packet.ProofHeight, packet.RollappId, packet.Status)
									broken = true
								}
							}
						}
					}
				} else {
					break
				}
				stateInfoIndex++
			}

		}
		return msg, broken
	}
}
