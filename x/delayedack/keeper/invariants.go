package keeper

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/utils/uinv"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	rtypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

var invs = uinv.NamedFuncsList[Keeper]{
	{Name: "proof-height", Func: InvariantProofHeight},
}

func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	invs.RegisterInvariants(types.ModuleName, ir, k)
}

// DO NOT DELETE
func AllInvariants(k Keeper) sdk.Invariant {
	return invs.All(types.ModuleName, k)
}

// ensures packet not finalized before proof height is finalized
func InvariantProofHeight(k Keeper) uinv.Func {
	return uinv.AnyErrorIsBreaking(func(ctx sdk.Context) error {
		var errs []error
		for _, ra := range k.rollappKeeper.GetAllRollapps(ctx) {
			err := k.checkRollapp(ctx, ra)
			err = errorsmod.Wrapf(err, "rollapp: %s", ra.RollappId)
			errs = append(errs, err)
		}
		return errors.Join(errs...)
	})
}

func (k Keeper) checkRollapp(ctx sdk.Context, ra rtypes.Rollapp) error {
	// will stay 0 if no state is ok
	// but will still check packets
	var latestFinalizedHeight uint64

	latestFinalizedStateIndex, ok := k.rollappKeeper.GetLatestFinalizedStateIndex(ctx, ra.RollappId)
	if !ok {
		return nil
	}

	latestFinalizedStateInfo := k.rollappKeeper.MustGetStateInfo(ctx, ra.RollappId, latestFinalizedStateIndex.Index)
	latestFinalizedHeight = latestFinalizedStateInfo.GetLatestHeight()

	packets := k.ListRollappPackets(ctx, types.ByRollappID(ra.RollappId))
	var errs []error
	for _, p := range packets {
		err := k.checkPacket(p, latestFinalizedHeight)
		err = errorsmod.Wrapf(err, "packet: %s", p.RollappId)
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func (k Keeper) checkPacket(p commontypes.RollappPacket, latestFinalizedHeight uint64) error {
	finalizedTooEarly := latestFinalizedHeight < p.ProofHeight && p.Status == commontypes.Status_FINALIZED
	if finalizedTooEarly {
		return fmt.Errorf("finalized too early height=%d, rollapp=%s, status=%s\n",
			p.ProofHeight, p.RollappId, p.Status,
		)
	}
	return nil
}
