package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

	"github.com/dymensionxyz/dymension/v3/x/agent/types"
	"github.com/dymensionxyz/dymension/v3/x/common/tee"
)

// fingerprintNotRevoked returns the policy's fingerprint, or
// ErrFailedPrecondition if it is in the revocation denylist.
func (k Keeper) fingerprintNotRevoked(ctx sdk.Context, p tee.Policy) (string, error) {
	fp, err := types.PolicyFingerprint(p)
	if err != nil {
		return "", errorsmod.Wrap(err, "policy fingerprint")
	}
	revoked, err := k.IsPolicyRevoked(ctx, fp)
	if err != nil {
		return "", errorsmod.Wrap(err, "is policy revoked")
	}
	if revoked {
		return "", gerrc.ErrFailedPrecondition.Wrapf("policy revoked: %s", fp)
	}
	return fp, nil
}

func (k Keeper) SetRevoked(ctx sdk.Context, fp string) error {
	return k.revokedPolicies.Set(ctx, fp)
}

func (k Keeper) DeleteRevoked(ctx sdk.Context, fp string) error {
	return k.revokedPolicies.Remove(ctx, fp)
}

func (k Keeper) IsPolicyRevoked(ctx sdk.Context, fp string) (bool, error) {
	return k.revokedPolicies.Has(ctx, fp)
}

// AllRevokedPolicies returns all revoked fingerprints in ascending order.
func (k Keeper) AllRevokedPolicies(ctx sdk.Context) ([]string, error) {
	var fps []string
	err := k.revokedPolicies.Walk(ctx, nil, func(fp string) (stop bool, err error) {
		fps = append(fps, fp)
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	return fps, nil
}

func (k msgServer) RevokePolicy(goCtx context.Context, msg *types.MsgRevokePolicy) (*types.MsgRevokePolicyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, errorsmod.Wrap(err, "validate basic")
	}
	if msg.Authority != k.authority {
		return nil, gerrc.ErrPermissionDenied.Wrapf("expected authority %s, got %s", k.authority, msg.Authority)
	}

	if err := k.SetRevoked(ctx, msg.Fingerprint); err != nil {
		return nil, errorsmod.Wrap(err, "set revoked")
	}

	if err := uevent.EmitTypedEvent(ctx, &types.EventPolicyRevoked{
		Fingerprint: msg.Fingerprint,
		Reason:      msg.Reason,
	}); err != nil {
		return nil, err
	}

	return &types.MsgRevokePolicyResponse{}, nil
}

func (k msgServer) UnrevokePolicy(goCtx context.Context, msg *types.MsgUnrevokePolicy) (*types.MsgUnrevokePolicyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, errorsmod.Wrap(err, "validate basic")
	}
	if msg.Authority != k.authority {
		return nil, gerrc.ErrPermissionDenied.Wrapf("expected authority %s, got %s", k.authority, msg.Authority)
	}

	if err := k.DeleteRevoked(ctx, msg.Fingerprint); err != nil {
		return nil, errorsmod.Wrap(err, "delete revoked")
	}

	if err := uevent.EmitTypedEvent(ctx, &types.EventPolicyUnrevoked{
		Fingerprint: msg.Fingerprint,
	}); err != nil {
		return nil, err
	}

	return &types.MsgUnrevokePolicyResponse{}, nil
}
