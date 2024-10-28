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
)

// RegisterInvariants registers the delayedack module invariants
func (k Keeper) RegisterInvariants(ir sdk.InvariantRegistry) {
	// INVARIANTS DISABLED SINCE LAZY FINALIZATION FEATURE
}

// PacketsFinalizationCorrespondsToFinalizationHeight checks that all rollapp packets stored are set to
// finalized status for all heights up to the latest height.
func PacketsFinalizationCorrespondsToFinalizationHeight(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			broken bool
			msg    string
		)

		for _, rollapp := range k.rollappKeeper.GetAllRollapps(ctx) {
			msg = k.checkRollapp(ctx, rollapp)
			if msg != "" {
				msg += fmt.Sprintf("rollapp: %s, msg: %s\n", rollapp.RollappId, msg)
				broken = true
			}
		}

		return sdk.FormatInvariant(types.ModuleName, routeFinalizedPacket, msg), broken
	}
}

func (k Keeper) checkRollapp(ctx sdk.Context, rollapp rtypes.Rollapp) (msg string) {
	// will stay 0 if no state is found
	// but will still check packets
	var latestFinalizedHeight uint64

	latestFinalizedStateIndex, found := k.rollappKeeper.GetLatestFinalizedStateIndex(ctx, rollapp.RollappId)
	if !found {
		return
	}

	latestFinalizedStateInfo := k.rollappKeeper.MustGetStateInfo(ctx, rollapp.RollappId, latestFinalizedStateIndex.Index)
	latestFinalizedHeight = latestFinalizedStateInfo.GetLatestHeight()

	packets := k.ListRollappPackets(ctx, types.ByRollappID(rollapp.RollappId))
	for _, packet := range packets {
		if packet.ProofHeight > latestFinalizedHeight && packet.Status == commontypes.Status_FINALIZED {
			return fmt.Sprintf("rollapp packet for the height should not be in finalized status. height=%d, rollapp=%s, status=%s\n",
				packet.ProofHeight, packet.RollappId, packet.Status)
		}
	}
	return
}
