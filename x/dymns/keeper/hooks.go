package keeper

import (
	"errors"

	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"

	errorsmod "cosmossdk.io/errors"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	"github.com/osmosis-labs/osmosis/v15/osmoutils"

	sdk "github.com/cosmos/cosmos-sdk/types"

	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

/* -------------------------------------------------------------------------- */
/*                             x/rollapp hooks                                */
/* -------------------------------------------------------------------------- */

// GetRollAppHooks returns the RollApp hooks struct.
func (k Keeper) GetRollAppHooks() rollapptypes.RollappHooks {
	return rollappHooks{
		Keeper: k,
	}
}

type rollappHooks struct {
	Keeper
}

var _ rollapptypes.RollappHooks = rollappHooks{}

func (h rollappHooks) RollappCreated(ctx sdk.Context, rollappID, alias string, creatorAddr sdk.AccAddress) error {
	if alias == "" {
		return nil
	}

	// ensure RollApp record is set
	if !h.Keeper.IsRollAppId(ctx, rollappID) {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "not a RollApp chain-id: %s", rollappID)
	}

	if !dymnsutils.IsValidAlias(alias) {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid alias format: %s", alias)
	}

	if !h.Keeper.CanUseAliasForNewRegistration(ctx, alias) {
		return errorsmod.Wrapf(gerrc.ErrAlreadyExists, "alias already in use or preserved: %s", alias)
	}

	priceParams := h.Keeper.PriceParams(ctx)

	aliasCost := sdk.NewCoins(
		sdk.NewCoin(
			priceParams.PriceDenom, priceParams.GetAliasPrice(alias),
		),
	)

	err := h.Keeper.registerAliasForRollApp(ctx, rollappID, creatorAddr, alias, aliasCost)
	if err != nil {
		return errorsmod.Wrap(errors.Join(gerrc.ErrUnknown, err), "register alias for RollApp")
	}

	return nil
}

func (h rollappHooks) BeforeUpdateState(_ sdk.Context, _ string, _ string, _ bool) error {
	return nil
}

func (h rollappHooks) AfterUpdateState(ctx sdk.Context, stateInfo *rollapptypes.StateInfoMeta) error {
	return nil
}

func (h rollappHooks) AfterStateFinalized(_ sdk.Context, _ string, _ *rollapptypes.StateInfo) error {
	return nil
}

func (h rollappHooks) OnHardFork(_ sdk.Context, _ string, _ uint64) error { return nil }

func (h rollappHooks) AfterTransfersEnabled(_ sdk.Context, _, _ string) error {
	return nil
}

type FutureRollappHooks interface {
	// TODO: remove/deprecate - rollapp id cannot change
	OnRollAppIdChanged(ctx sdk.Context, previousRollAppId, newRollAppId string)
}

// TODO: Hooks should embed the noop base type, and only implement what they need, instead of repeating the whole interface.
var _ FutureRollappHooks = rollappHooks{}

func (k Keeper) GetFutureRollAppHooks() FutureRollappHooks {
	return rollappHooks{
		Keeper: k,
	}
}

// OnRollAppIdChanged implements FutureRollappHooks.
func (h rollappHooks) OnRollAppIdChanged(ctx sdk.Context, previousRollAppId, newRollAppId string) {
	logger := h.Logger(ctx).With(
		"old-rollapp-id", previousRollAppId, "new-rollapp-id", newRollAppId,
	)

	logger.Info("begin DymNS hook on RollApp ID changed.")

	// Due to the critical nature reason of the hook,
	// each step will be done in branched context and drop if error, to prevent any side effects.

	if err := osmoutils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
		aliasesLinkedToPreviousRollApp := h.GetAliasesOfRollAppId(ctx, previousRollAppId)
		if len(aliasesLinkedToPreviousRollApp) == 0 {
			return nil
		}

		for _, alias := range aliasesLinkedToPreviousRollApp {
			if err := h.MoveAliasToRollAppId(ctx, previousRollAppId, alias, newRollAppId); err != nil {
				return errorsmod.Wrapf(errors.Join(gerrc.ErrUnknown, err), "failed to migrate alias: %s", alias)
			}
		}

		// now priority the first alias from previous RollApp, because users are already familiar with it.
		return h.SetDefaultAliasForRollApp(ctx, newRollAppId, aliasesLinkedToPreviousRollApp[0])
	}); err != nil {
		logger.Error("aborted alias migration.", "error", err)
		return
	}

	if err := osmoutils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
		previousChainIdsToNewChainId := map[string]string{
			previousRollAppId: newRollAppId,
		}

		if err := h.migrateChainIdsInDymNames(ctx, previousChainIdsToNewChainId); err != nil {
			return errorsmod.Wrapf(errors.Join(gerrc.ErrUnknown, err), "failed to migrate chain-ids in Dym-Names")
		}

		return nil
	}); err != nil {
		logger.Error("aborted chain-id migration in Dym-Names configurations.", "error", err)
		return
	}

	logger.Info("finished DymNS hook on RollApp ID changed.")
}
