package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k msgServer) MarkObsoleteRollapps(goCtx context.Context, msg *types.MsgMarkObsoleteRollapps) (*types.MsgMarkObsoleteRollappsResponse, error) {
	err := msg.ValidateBasic()
	if err != nil {
		return nil, err
	}

	if msg.Authority != k.authority {
		return nil, errorsmod.Wrap(gerrc.ErrInvalidArgument, "only the gov module can mark obsolete rollapps")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	obsoleteNum, err := k.Keeper.MarkObsoleteRollapps(ctx, msg.DrsVersions)
	if err != nil {
		return nil, fmt.Errorf("mark obsolete rollapps: %w", err)
	}

	err = uevent.EmitTypedEvent(ctx, &types.EventMarkObsoleteRollapps{
		ObsoleteRollappNum: uint64(obsoleteNum),
		DrsVersions:        msg.DrsVersions,
	})
	if err != nil {
		return nil, fmt.Errorf("emit event: %w", err)
	}

	return &types.MsgMarkObsoleteRollappsResponse{}, nil
}

func (k Keeper) MarkObsoleteRollapps(ctx sdk.Context, drsVersions []uint32) (int, error) {
	obsoleteVersions := make(map[uint32]struct{})
	for _, v := range drsVersions {
		obsoleteVersions[v] = struct{}{}
		// this also saves in the state the obsolete version
		err := k.SetObsoleteDRSVersion(ctx, v)
		if err != nil {
			return 0, fmt.Errorf("set obsolete DRS version: %w", err)
		}
	}

	var (
		logger      = k.Logger(ctx)
		obsoleteNum int
	)
	for _, rollapp := range k.GetAllRollapps(ctx) {
		info, found := k.GetLatestStateInfo(ctx, rollapp.RollappId)
		if !found {
			logger.With("rollapp_id", rollapp.RollappId).Info("no latest state info for rollapp")
			continue
		}

		// check only last block descriptor DRS, since if that last is not obsolete it means the rollapp already upgraded and is not obsolete anymore
		bd := info.BDs.BD[len(info.BDs.BD)-1]

		_, obsolete := obsoleteVersions[bd.DrsVersion]
		if obsolete {
			// If this fails, no state change happens
			err := osmoutils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
				return k.HardForkToLatest(ctx, rollapp.RollappId)
			})
			if err != nil {
				// We do not want to fail if one rollapp cannot to be marked as obsolete
				k.Logger(ctx).With("rollapp_id", rollapp.RollappId, "drs_version", bd.DrsVersion, "error", err.Error()).
					Error("Failed to mark rollapp as obsolete")
			}
			obsoleteNum++
		}
	}

	return obsoleteNum, nil
}
