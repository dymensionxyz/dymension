package keeper

import (
	"errors"
	"sort"

	errorsmod "cosmossdk.io/errors"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	"github.com/osmosis-labs/osmosis/v15/osmoutils"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	epochstypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"
)

/* -------------------------------------------------------------------------- */
/*                              x/epochs hooks                                */
/* -------------------------------------------------------------------------- */

var _ epochstypes.EpochHooks = epochHooks{}

type epochHooks struct {
	Keeper
}

// GetEpochHooks returns the epoch hooks for the module.
func (k Keeper) GetEpochHooks() epochstypes.EpochHooks {
	return epochHooks{
		Keeper: k,
	}
}

// BeforeEpochStart is the epoch start hook.
// Business logic is to prune the expired preserved registration.
func (e epochHooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	params := e.GetParams(ctx)

	if epochIdentifier != params.Misc.BeginEpochHookIdentifier {
		return nil
	}

	e.Keeper.Logger(ctx).Info("DymNS hook Before-Epoch-Start: triggered", "epoch-number", epochNumber, "epoch-identifier", epochIdentifier)

	var updatedParams bool
	params, updatedParams = e.processCleanupPreservedRegistration(ctx, epochIdentifier, epochNumber, params)
	if updatedParams {
		if err := e.SetParams(ctx, params); err != nil {
			return err
		}
	}

	return nil
}

// processCleanupPreservedRegistration clears preserved registration if it's expired.
func (e epochHooks) processCleanupPreservedRegistration(ctx sdk.Context, epochIdentifier string, epochNumber int64, params dymnstypes.Params) (updatedParams dymnstypes.Params, updated bool) {
	if params.PreservedRegistration.ExpirationEpoch >= ctx.BlockTime().Unix() {
		return
	}

	// expired, clear it

	e.Keeper.Logger(ctx).Info(
		"DymNS hook Before-Epoch-Start: processing cleanup preserved registration",
		"epoch-number", epochNumber, "epoch-identifier", epochIdentifier,
	)

	updatedParams = params

	updatedParams.PreservedRegistration = dymnstypes.PreservedRegistrationParams{}
	updated = true

	return
}

// AfterEpochEnd is the epoch end hook.
func (e epochHooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	params := e.GetParams(ctx)

	if epochIdentifier != params.Misc.EndEpochHookIdentifier {
		return nil
	}

	e.Keeper.Logger(ctx).Info("DymNS hook After-Epoch-End: triggered", "epoch-number", epochNumber, "epoch-identifier", epochIdentifier)

	if params.Misc.EnableTradingName {
		if err := e.processActiveDymNameSellOrders(ctx, epochIdentifier, epochNumber); err != nil {
			return err
		}
	}

	if params.Misc.EnableTradingAlias {
		if err := e.processActiveAliasSellOrders(ctx, epochIdentifier, epochNumber, params); err != nil {
			return err
		}
	}

	return nil
}

// processActiveDymNameSellOrders process the finished Dym-Name Sell-Orders.
// Sell-Order will be deleted. If the Sell-Order has a winner, the Dym-Name ownership will be transferred.
func (e epochHooks) processActiveDymNameSellOrders(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	dk := e.Keeper

	aSoe := dk.GetActiveSellOrdersExpiration(ctx, dymnstypes.TypeName)
	if len(aSoe.Records) < 1 {
		return nil
	}

	nowEpochUTC := ctx.BlockTime().Unix()
	var finishedSOs []dymnstypes.SellOrder
	invalidRecordsToRemove := make([]string, 0)

	for i, record := range aSoe.Records {
		if record.ExpireAt > nowEpochUTC {
			// skip not expired ones
			continue
		}

		so := dk.GetSellOrder(ctx, record.AssetId, dymnstypes.TypeName)

		if so == nil {
			// remove the invalid entry
			invalidRecordsToRemove = append(invalidRecordsToRemove, record.AssetId)
			continue
		}

		if !so.HasFinished(nowEpochUTC) {
			// invalid entry
			dk.Logger(ctx).Error(
				"DymNS hook After-Epoch-End: sell order has not finished",
				"asset-id", record.AssetId, "asset-type", dymnstypes.TypeName.FriendlyString(),
				"expiry", record.ExpireAt, "now", nowEpochUTC,
				"epoch-number", epochNumber, "epoch-identifier", epochIdentifier,
			)

			aSoe.Records[i].ExpireAt = so.ExpireAt // correct it
			continue
		}

		finishedSOs = append(finishedSOs, *so)
	}

	for _, name := range invalidRecordsToRemove {
		aSoe.Remove(name)
	}

	if len(finishedSOs) < 1 {
		// skip updating store
		return nil
	}

	sort.Slice(finishedSOs, func(i, j int) bool {
		return finishedSOs[i].AssetId < finishedSOs[j].AssetId
	})

	dk.Logger(ctx).Info(
		"DymNS hook After-Epoch-End: processing finished SOs",
		"asset-type", dymnstypes.TypeName.FriendlyString(),
		"count", len(finishedSOs),
		"epoch-number", epochNumber, "epoch-identifier", epochIdentifier,
	)

	for _, so := range finishedSOs {
		var errApplyStateChange error
		// each order should be processed in a branched context, if error, discard the state change
		// and process next order, to prevent chain reaction when an individual order failed to process

		if so.HighestBid != nil {
			errApplyStateChange = osmoutils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
				return dk.CompleteDymNameSellOrder(ctx, so.AssetId)
			})
			if errApplyStateChange != nil {
				dk.Logger(ctx).Error(
					"DymNS hook After-Epoch-End: failed to complete sell order",
					"asset-id", so.AssetId, "asset-type", dymnstypes.TypeName.FriendlyString(),
					"expiry", so.ExpireAt, "now", nowEpochUTC,
					"epoch-number", epochNumber, "epoch-identifier", epochIdentifier,
					"error", errApplyStateChange,
				)
			}
		} else {
			dk.DeleteSellOrder(ctx, so.AssetId, dymnstypes.TypeName)
			errApplyStateChange = nil
		}

		if errApplyStateChange == nil {
			aSoe.Remove(so.AssetId)
		}
	}

	if err := dk.SetActiveSellOrdersExpiration(ctx, aSoe, dymnstypes.TypeName); err != nil {
		dk.Logger(ctx).Error(
			"DymNS hook After-Epoch-End: failed to update active SO expiry",
			"asset-type", dymnstypes.TypeName.FriendlyString(),
			"epoch-number", epochNumber, "epoch-identifier", epochIdentifier,
			"error", err,
		)
		return err
	}

	return nil
}

// processActiveAliasSellOrders process the finished Alias Sell-Orders.
// Sell-Order will be deleted. If the Sell-Order has a winner, the Alias linking will be updated.
func (e epochHooks) processActiveAliasSellOrders(ctx sdk.Context, epochIdentifier string, epochNumber int64, params dymnstypes.Params) error {
	dk := e.Keeper

	aSoe := dk.GetActiveSellOrdersExpiration(ctx, dymnstypes.TypeAlias)
	if len(aSoe.Records) < 1 {
		return nil
	}

	nowEpochUTC := ctx.BlockTime().Unix()
	var finishedSOs []dymnstypes.SellOrder
	invalidRecordsToRemove := make([]string, 0)

	for i, record := range aSoe.Records {
		if record.ExpireAt > nowEpochUTC {
			// skip not expired ones
			continue
		}

		so := dk.GetSellOrder(ctx, record.AssetId, dymnstypes.TypeAlias)

		if so == nil {
			// remove the invalid entry
			invalidRecordsToRemove = append(invalidRecordsToRemove, record.AssetId)
			continue
		}

		if !so.HasFinished(nowEpochUTC) {
			// invalid entry
			dk.Logger(ctx).Error(
				"DymNS hook After-Epoch-End: sell order has not finished",
				"asset-id", record.AssetId, "asset-type", dymnstypes.TypeAlias.FriendlyString(),
				"expiry", record.ExpireAt, "now", nowEpochUTC,
				"epoch-number", epochNumber, "epoch-identifier", epochIdentifier,
			)

			aSoe.Records[i].ExpireAt = so.ExpireAt // correct it
			continue
		}

		finishedSOs = append(finishedSOs, *so)
	}

	for _, name := range invalidRecordsToRemove {
		aSoe.Remove(name)
	}

	if len(finishedSOs) < 1 {
		// skip updating store
		return nil
	}

	sort.Slice(finishedSOs, func(i, j int) bool {
		return finishedSOs[i].AssetId < finishedSOs[j].AssetId
	})

	dk.Logger(ctx).Info(
		"DymNS hook After-Epoch-End: processing finished SOs",
		"asset-type", dymnstypes.TypeAlias.FriendlyString(),
		"count", len(finishedSOs),
		"epoch-number", epochNumber, "epoch-identifier", epochIdentifier,
	)

	prohibitedToTradeAliases := e.GetAllAliasAndChainIdInParams(ctx)

	for _, so := range finishedSOs {
		var errApplyStateChange error
		// each order should be processed in a branched context, if error, discard the state change
		// and process next order, to prevent chain reaction when an individual order failed to process

		if so.HighestBid == nil {
			// no bid placed, it just a normal expiry without winner
			dk.DeleteSellOrder(ctx, so.AssetId, dymnstypes.TypeAlias)
			errApplyStateChange = nil
		} else if _, forceCancel := prohibitedToTradeAliases[so.AssetId]; forceCancel {
			// Sell-Order will be force cancelled and refund bids if any,
			// when the alias is prohibited to trade

			errApplyStateChange = osmoutils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
				return dk.RefundBid(ctx, *so.HighestBid, dymnstypes.TypeAlias)
			})
			if errApplyStateChange != nil {
				dk.Logger(ctx).Error(
					"DymNS hook After-Epoch-End: failed to refund bid for a force-to-cancel sell order",
					"asset-id", so.AssetId, "asset-type", dymnstypes.TypeAlias.FriendlyString(),
					"expiry", so.ExpireAt, "now", nowEpochUTC,
					"epoch-number", epochNumber, "epoch-identifier", epochIdentifier,
					"error", errApplyStateChange,
				)
			} else {
				dk.DeleteSellOrder(ctx, so.AssetId, dymnstypes.TypeAlias)
			}
		} else {
			errApplyStateChange = osmoutils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
				return dk.CompleteAliasSellOrder(ctx, so.AssetId)
			})
			if errApplyStateChange != nil {
				dk.Logger(ctx).Error(
					"DymNS hook After-Epoch-End: failed to complete sell order",
					"asset-id", so.AssetId, "asset-type", dymnstypes.TypeAlias.FriendlyString(),
					"expiry", so.ExpireAt, "now", nowEpochUTC,
					"epoch-number", epochNumber, "epoch-identifier", epochIdentifier,
					"error", errApplyStateChange,
				)
			}
		}

		if errApplyStateChange == nil {
			aSoe.Remove(so.AssetId)
		}
	}

	if err := dk.SetActiveSellOrdersExpiration(ctx, aSoe, dymnstypes.TypeAlias); err != nil {
		dk.Logger(ctx).Error(
			"DymNS hook After-Epoch-End: failed to update active SO expiry",
			"asset-type", dymnstypes.TypeAlias.FriendlyString(),
			"epoch-number", epochNumber, "epoch-identifier", epochIdentifier,
			"error", err,
		)
		return err
	}

	return nil
}

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

	canUseAlias, err := h.Keeper.CanUseAliasForNewRegistration(ctx, alias)
	if err != nil {
		return errorsmod.Wrapf(errors.Join(gerrc.ErrInternal, err), "failed to check availability of alias: %s", alias)
	}

	if !canUseAlias {
		return errorsmod.Wrapf(gerrc.ErrAlreadyExists, "alias already in use or preserved: %s", alias)
	}

	moduleParams := h.Keeper.GetParams(ctx)

	aliasCost := sdk.NewCoins(
		sdk.NewCoin(
			moduleParams.Price.PriceDenom, moduleParams.Price.GetAliasPrice(alias),
		),
	)

	return h.Keeper.registerAliasForRollApp(ctx, rollappID, creatorAddr, alias, aliasCost)
}

func (h rollappHooks) BeforeUpdateState(_ sdk.Context, _ string, _ string) error {
	return nil
}

func (h rollappHooks) AfterStateFinalized(_ sdk.Context, _ string, _ *rollapptypes.StateInfo) error {
	return nil
}

func (h rollappHooks) FraudSubmitted(_ sdk.Context, _ string, _ uint64, _ string) error {
	return nil
}

// FutureRollappHooks is temporary added to handle future hooks that not available yet.
type FutureRollappHooks interface {
	OnRollAppIdChanged(ctx sdk.Context, previousRollAppId, newRollAppId string) error
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

func (h rollappHooks) OnRollAppIdChanged(ctx sdk.Context, previousRollAppId, newRollAppId string) error {
	// This can be call when the RollAppId is changed due to fraud submission,
	// so due to the critical nature of the reason,
	// the following execution will be done in branched context to prevent any side effects.
	// If any error occurs, the state change to this module will be discarded, no error returned to the caller.

	logger := h.Logger(ctx).With(
		"old-rollapp-id", previousRollAppId, "new-rollapp-id", newRollAppId,
	)

	logger.Info("begin DymNS hook on RollApp ID changed")
	defer func() {
		logger.Info("finished DymNS hook on RollApp ID changed")
	}()

	if err := osmoutils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
		aliasesLinkedToPreviousRollApp := h.GetAliasesOfRollAppId(ctx, previousRollAppId)
		if len(aliasesLinkedToPreviousRollApp) == 0 {
			return nil
		}

		for _, alias := range aliasesLinkedToPreviousRollApp {
			if err := h.MoveAliasToRollAppId(ctx, previousRollAppId, alias, newRollAppId); err != nil {
				logger.Error("failed to migrate alias", "alias", alias, "error", err)
				return err
			}
		}

		// now priority the first alias from previous RollApp, because users are already familiar with it.
		return h.SetDefaultAlias(ctx, newRollAppId, aliasesLinkedToPreviousRollApp[0])
	}); err != nil {
		logger.Error("aborted alias migration", "error", err)

		// do not return error, that might cause the caller to revert an important execution
		return nil
	}

	if err := osmoutils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
		previousChainIdsToNewChainId := map[string]string{
			previousRollAppId: newRollAppId,
		}

		if err := h.migrateChainIdsInDymNames(ctx, previousChainIdsToNewChainId); err != nil {
			logger.Error("failed to migrate chain-ids in Dym-Names", "error", err)
			return err
		}

		return nil
	}); err != nil {
		logger.Error("aborted chain-id migration in Dym-Names configurations", "error", err)

		// do not return error, that might cause the caller to revert an important execution
		return nil
	}

	return nil
}
