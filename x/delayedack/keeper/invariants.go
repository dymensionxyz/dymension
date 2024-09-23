package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	rtypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

const (
	routeFinalizedPacket = "rollapp-finalized-packet"
	routeRevertedPacket  = "rollapp-reverted-packet"
)

// RegisterInvariants registers the delayedack module invariants
func (k Keeper) RegisterInvariants(ir sdk.InvariantRegistry) {
	ir.RegisterRoute(types.ModuleName, routeFinalizedPacket, k.PacketsFinalizationCorrespondsToFinalizationHeight)
	ir.RegisterRoute(types.ModuleName, routeRevertedPacket, k.PacketsFromRevertedHeightsAreReverted)
}

// PacketsFinalizationCorrespondsToFinalizationHeight checks that all rollapp packets stored are set to
// finalized status for all heights up to the latest height.
// Skip the check if the rollapp is frozen
func (k Keeper) PacketsFinalizationCorrespondsToFinalizationHeight(ctx sdk.Context) (string, bool) {
	return k.packetsCorrespondsToStatusHeight(checkFinalizedPackets, false)(ctx)
}

// PacketsFromRevertedHeightsAreReverted checks that all rollapp packets stored are set to
// reverted status for all heights up to the latest height
// Check if the rollapp is frozen
func (k Keeper) PacketsFromRevertedHeightsAreReverted(ctx sdk.Context) (string, bool) {
	return k.packetsCorrespondsToStatusHeight(checkRevertedPackets, true)(ctx)
}

type checkPacketsFn func(packets []commontypes.RollappPacket, latestFinalizedHeight uint64) string

// packetsCorrespondsToStatusHeight checks that all rollapp packets stored are set to adequate status for all heights up to the latest height
func (k Keeper) packetsCorrespondsToStatusHeight(checkPackets checkPacketsFn, checkRollappFrozen bool) sdk.Invariant {
	return func(ctx sdk.Context) (msg string, stop bool) {
		for _, rollapp := range k.rollappKeeper.GetAllRollapps(ctx) {
			msg = k.checkRollapp(ctx, rollapp, checkPackets, checkRollappFrozen)
			if stop = msg != ""; stop {
				break
			}
		}

		return
	}
}

func (k Keeper) checkRollapp(ctx sdk.Context, rollapp rtypes.Rollapp, checkPackets checkPacketsFn, checkFrozen bool) (msg string) {
	if checkFrozen && !rollapp.Frozen {
		return
	}

	// will stay 0 if no state is found
	// but will still check packets
	var latestFinalizedHeight uint64

	defer func() {
		if msg == "" {
			packets := k.ListRollappPackets(ctx, types.ByRollappID(rollapp.RollappId))
			msg = checkPackets(packets, latestFinalizedHeight)
		}
	}()

	latestFinalizedStateIndex, found := k.rollappKeeper.GetLatestFinalizedStateIndex(ctx, rollapp.RollappId)
	if !found {
		return
	}

	latestFinalizedStateInfo, found := k.rollappKeeper.GetStateInfo(ctx, rollapp.RollappId, latestFinalizedStateIndex.Index)
	if !found {
		msg = fmt.Sprintf("latest finalized state info not found for rollapp: %s", rollapp.RollappId)
		return
	}

	latestFinalizedHeight = latestFinalizedStateInfo.GetLatestHeight()

	return
}

// checkFinalizedPackets checks that all rollapp packets stored are set to finalized status for all heights up to the latest height
func checkFinalizedPackets(packets []commontypes.RollappPacket, latestFinalizedHeight uint64) (_ string) {
	for _, packet := range packets {
		if packet.ProofHeight > latestFinalizedHeight && packet.Status == commontypes.Status_FINALIZED {
			return fmt.Sprintf("rollapp packet for the height should not be in finalized status. height=%d, rollapp=%s, status=%s\n",
				packet.ProofHeight, packet.RollappId, packet.Status)
		}
	}
	return
}

// checkRevertedPackets checks that all rollapp packets stored are set to reverted status for all heights up to the latest height
func checkRevertedPackets(packets []commontypes.RollappPacket, latestFinalizedHeight uint64) (_ string) {
	for _, packet := range packets {
		if packet.ProofHeight > latestFinalizedHeight && packet.Status != commontypes.Status_REVERTED {
			return fmt.Sprintf("packet should be reverted: rollapp: %s: height: %d: status: %s",
				packet.RollappId, packet.ProofHeight, packet.Status)
		}
	}

	return
}
