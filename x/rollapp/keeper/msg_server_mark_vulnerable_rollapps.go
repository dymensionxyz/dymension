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

func (k msgServer) MarkVulnerableRollapps(goCtx context.Context, msg *types.MsgMarkVulnerableRollapps) (*types.MsgMarkVulnerableRollappsResponse, error) {
	err := msg.ValidateBasic()
	if err != nil {
		return nil, err
	}

	if msg.Authority != k.authority {
		return nil, errorsmod.Wrap(gerrc.ErrInvalidArgument, "only the gov module can mark vulnerable rollapps")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	vulnerableNum, err := k.Keeper.MarkVulnerableRollapps(ctx, msg.DrsVersions)
	if err != nil {
		return nil, fmt.Errorf("mark vulnerable rollapps: %w", err)
	}

	err = uevent.EmitTypedEvent(ctx, &types.EventMarkVulnerableRollapps{
		VulnerableRollappNum: uint64(vulnerableNum),
		DrsVersions:          msg.DrsVersions,
	})
	if err != nil {
		return nil, fmt.Errorf("emit event: %w", err)
	}

	return &types.MsgMarkVulnerableRollappsResponse{}, nil
}

func (k Keeper) MarkVulnerableRollapps(ctx sdk.Context, drsVersions []uint32) (int, error) {
	vulnerableVersions := make(map[uint32]struct{})
	for _, v := range drsVersions {
		vulnerableVersions[v] = struct{}{}
		// this also saves in the state the vulnerable version
		err := k.SetVulnerableDRSVersion(ctx, v)
		if err != nil {
			return 0, fmt.Errorf("set vulnerable DRS version: %w", err)
		}
	}

	var (
		logger        = k.Logger(ctx)
		nonVulnerable = k.FilterRollapps(ctx, FilterNonVulnerable)
		vulnerableNum int
	)
	for _, rollapp := range nonVulnerable {
		info, found := k.GetLatestStateInfo(ctx, rollapp.RollappId)
		if !found {
			logger.With("rollapp_id", rollapp.RollappId).Info("no latest state info for rollapp")
			continue
		}

		// check only last block descriptor DRS, since if that last is not vulnerable it means the rollapp already upgraded and is not vulnerable anymore
		bd := info.BDs.BD[len(info.BDs.BD)-1]

		_, vulnerable := vulnerableVersions[bd.DrsVersion]
		if vulnerable {
			// If this fails, no state change happens
			err := osmoutils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
				return k.MarkRollappAsVulnerable(ctx, rollapp.RollappId)
			})
			if err != nil {
				// We do not want to fail if one rollapp cannot to be marked as vulnerable
				k.Logger(ctx).With("rollapp_id", rollapp.RollappId, "drs_version", bd.DrsVersion, "error", err.Error()).
					Error("Failed to mark rollapp as vulnerable")
			}
			vulnerableNum++
		}
	}

	return vulnerableNum, nil
}
