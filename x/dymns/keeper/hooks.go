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

	return h.Keeper.registerAliasForRollApp(ctx, rollappID, creatorAddr, alias, aliasCost)
}

func (h rollappHooks) BeforeUpdateState(_ sdk.Context, _ string, _ string, _ bool) error {
	return nil
}

func (h rollappHooks) AfterUpdateState(_ sdk.Context, _ string, _ *rollapptypes.StateInfo) error {
	return nil
}

func (h rollappHooks) AfterStateFinalized(_ sdk.Context, _ string, _ *rollapptypes.StateInfo) error {
	return nil
}

func (h rollappHooks) FraudSubmitted(_ sdk.Context, _ string, _ uint64, _ string) error {
	return nil
}

func (h rollappHooks) AfterTransfersEnabled(_ sdk.Context, _, _ string) error {
	return nil
}

// FutureRollappHooks is temporary added to handle future hooks that not available yet.
type FutureRollappHooks interface {
	// OnRollAppIdChanged is called when a RollApp's ID is changed, typically due to fraud submission.
	// It migrates all aliases and Dym-Names associated with the previous RollApp ID to the new one.
	// This function executes step by step in a branched context to prevent side effects, and any errors
	// during execution will result in the state changes being discarded.
	//
	// Parameters:
	//   - ctx: The SDK context
	//   - previousRollAppId: The original ID of the RollApp
	//   - newRollAppId: The new ID assigned to the RollApp
	OnRollAppIdChanged(ctx sdk.Context, previousRollAppId, newRollAppId string)
	// Just a pseudo method signature, the actual method signature might be different.

	// TODO DymNS: connect to the actual implementation when the hooks are available.
	//   The implementation of OnRollAppIdChanged assume that both of the RollApp records are exists in the x/rollapp store.
}

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
				return errorsmod.Wrapf(errors.Join(gerrc.ErrInternal, err), "failed to migrate alias: %s", alias)
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
			return errorsmod.Wrapf(errors.Join(gerrc.ErrInternal, err), "failed to migrate chain-ids in Dym-Names")
		}

		return nil
	}); err != nil {
		logger.Error("aborted chain-id migration in Dym-Names configurations.", "error", err)
		return
	}

	logger.Info("finished DymNS hook on RollApp ID changed.")
}
