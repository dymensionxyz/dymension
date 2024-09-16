package keeper

import (
	"context"
	"fmt"

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
		return nil, gerrc.ErrInvalidArgument.Wrapf("only the gov module can mark vulnerable rollapps")
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
			return 0, fmt.Errorf("set vulterable DRS version: %w", err)
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
			logger.With("rollapp_id", rollapp.RollappId).Error("no latest state info for rollapp")
			continue
		}
		if info.DrsVersion == "" {
			logger.With("rollapp_id", rollapp.RollappId).Error("no DRS version set for rollapp")
		}

		_, vulnerable := vulnerableVersions[info.DrsVersion]
		if vulnerable {
			k.MustMarkRollappAsVulnerable(ctx, rollapp.RollappId)
			vulnerableNum++
		}
	}

	return vulnerableNum, nil
}
