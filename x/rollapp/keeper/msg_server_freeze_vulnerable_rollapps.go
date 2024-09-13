package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

	fpslices "github.com/dymensionxyz/dymension/v3/utils/fp/slices"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k msgServer) FreezeVulnerableRollapps(goCtx context.Context, msg *types.MsgFreezeVulnerableRollapps) (*types.MsgFreezeVulnerableRollappsResponse, error) {
	err := msg.ValidateBasic()
	if err != nil {
		return nil, err
	}

	if msg.Authority != k.authority {
		return nil, sdkerrors.ErrorInvalidSigner.Wrapf("Only the gov module can freeze vulnerable rollapps")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	drsVersions := fpslices.Map(msg.DrsVersions, fpslices.StringToType[types.DRSVersion])
	vulnerableNum, err := k.Keeper.FreezeVulnerableRollapps(ctx, drsVersions)
	if err != nil {
		return nil, fmt.Errorf("freeze vulnerable rollapps: %w", err)
	}

	err = uevent.EmitTypedEvent(ctx, &types.EventFreezeVulnerableRollapps{
		RollappNum:  uint64(vulnerableNum),
		DrsVersions: msg.DrsVersions,
	})
	if err != nil {
		return nil, fmt.Errorf("emit event: %w", err)
	}

	return &types.MsgFreezeVulnerableRollappsResponse{}, nil
}

func (k Keeper) FreezeVulnerableRollapps(ctx sdk.Context, versions []types.DRSVersion) (int, error) {
	unfrozenRollapps, err := k.GetRollappsByFrozenStatus(ctx, false)
	if err != nil {
		return 0, err
	}

	vulnerableVersions := make(map[types.DRSVersion]struct{})
	for _, v := range versions {
		vulnerableVersions[v] = struct{}{}
		// this also saves in the state the vulnerable version
		err = k.vulnerableDRSVersions.Set(ctx, v)
		if err != nil {
			return 0, fmt.Errorf("set vulterable DRS version: %w", err)
		}
	}

	var vulnerableNum int
	for _, id := range unfrozenRollapps {
		version, found := k.GetRollappDRSVersion(ctx, id)
		if !found {
			// TODO: log and continue?
			return 0, fmt.Errorf("no DRS found for rollapp %s", id)
		}
		if version == "" {
			// TODO: log and continue? the version is not set yet, no MsgUpdateState was submitted after the software update
		}

		_, vulnerable := vulnerableVersions[version]
		if vulnerable {
			k.MustMarkRollappAsVulnerable(ctx, id)
			vulnerableNum++
		}
	}

	return vulnerableNum, nil
}
