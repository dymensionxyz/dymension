package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

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

func (k Keeper) MarkVulnerableRollapps(ctx sdk.Context, drsVersions []string) (int, error) {
	vulnerableVersions := make(map[string]struct{})
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

		// We only check first and last BD to avoid DoS attack related to iterating big number of BDs (taking into account a state update can be submitted with any numblock value)
		// It is assumed there cannot be two upgrades in the same state update (since it requires gov proposal), if this happens it will be a fraud caught by Rollapp validators.
		// Therefore checking first and last BD for deprecated DRS version should be enough.
		var bdsToCheck []*types.BlockDescriptor
		bdsToCheck = append(bdsToCheck, &info.BDs.BD[0])
		if info.NumBlocks > 1 {
			bdsToCheck = append(bdsToCheck, &info.BDs.BD[len(info.BDs.BD)-1])
		}
		for _, bd := range bdsToCheck {
			// TODO: this check may be deleted once empty DRS version is marked vulnerable
			//  https://github.com/dymensionxyz/dymension/issues/1233
			if bd.DrsVersion == "" {
				logger.With("rollapp_id", rollapp.RollappId).Info("no DRS version set for rollapp")
			}

			_, vulnerable := vulnerableVersions[bd.DrsVersion]
			if vulnerable {
				err := k.MarkRollappAsVulnerable(ctx, rollapp.RollappId)
				if err != nil {
					return 0, fmt.Errorf("freeze rollapp: %w", err)
				}
				vulnerableNum++
				break
			}
		}
	}

	return vulnerableNum, nil
}
