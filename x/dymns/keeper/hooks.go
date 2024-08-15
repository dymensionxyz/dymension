package keeper

import (
	"errors"
	"sort"

	"github.com/cometbft/cometbft/libs/log"

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
func (e epochHooks) BeforeEpochStart(_ sdk.Context, _ string, _ int64) error {
	return nil
}

// AfterEpochEnd is the epoch end hook.
func (e epochHooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	miscParams := e.MiscParams(ctx)

	if epochIdentifier != miscParams.EndEpochHookIdentifier {
		return nil
	}

	logger := e.Logger(ctx).With("hook", "After-Epoch-End", "epoch-number", epochNumber, "epoch-identifier", epochIdentifier)

	if miscParams.EnableTradingName {
		if err := e.processActiveDymNameSellOrders(ctx, logger); err != nil {
			return err
		}
	}

	if miscParams.EnableTradingAlias {
		if err := e.processActiveAliasSellOrders(ctx, logger); err != nil {
			return err
		}
	}

	return nil
}

// processActiveDymNameSellOrders process the finished Dym-Name Sell-Orders.
// Sell-Order will be deleted. If the Sell-Order has a winner, the Dym-Name ownership will be transferred.
func (e epochHooks) processActiveDymNameSellOrders(ctx sdk.Context, logger log.Logger) error {
	finishedSOs, aSoe := e.getFinishedSellOrders(ctx, dymnstypes.TypeName, logger)

	if len(finishedSOs) < 1 || len(aSoe.Records) < 1 {
		return nil
	}

	logger.Info("processing finished SOs", "count", len(finishedSOs))

	for _, so := range finishedSOs {
		// each order should be processed in a branched context, if error, discard the state change
		// and process next order, to prevent chain reaction when an individual order failed to process
		errApplyStateChange := osmoutils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
			if so.HighestBid == nil {
				e.DeleteSellOrder(ctx, so.AssetId, dymnstypes.TypeName)
				return nil
			}

			if err := e.CompleteDymNameSellOrder(ctx, so.AssetId); err != nil {
				logger.Error("failed to complete sell order", "asset-id", so.AssetId, "error", err)
				return err
			}

			return nil
		})

		if errApplyStateChange == nil {
			aSoe.Remove(so.AssetId)
		}
	}

	if err := e.SetActiveSellOrdersExpiration(ctx, aSoe, dymnstypes.TypeName); err != nil {
		e.Keeper.Logger(ctx).Error("failed to update active SO expiry", "error", err)
		return err
	}

	return nil
}

// processActiveAliasSellOrders process the finished Alias Sell-Orders.
// Sell-Order will be deleted. If the Sell-Order has a winner, the Alias linking will be updated.
func (e epochHooks) processActiveAliasSellOrders(ctx sdk.Context, logger log.Logger) error {
	finishedSOs, aSoe := e.getFinishedSellOrders(ctx, dymnstypes.TypeAlias, logger)

	if len(finishedSOs) < 1 || len(aSoe.Records) < 1 {
		return nil
	}

	logger.Info("processing finished SOs", "count", len(finishedSOs))

	prohibitedToTradeAliases := e.GetAllAliasAndChainIdInParams(ctx)

	for _, so := range finishedSOs {
		// each order should be processed in a branched context, if error, discard the state change
		// and process next order, to prevent chain reaction when an individual order failed to process
		errApplyStateChange := osmoutils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
			if so.HighestBid == nil {
				e.DeleteSellOrder(ctx, so.AssetId, dymnstypes.TypeAlias)
				return nil
			}

			if _, forceCancel := prohibitedToTradeAliases[so.AssetId]; forceCancel {
				// Sell-Order will be force cancelled and refund bids if any,
				// when the alias is prohibited to trade
				if err := e.RefundBid(ctx, *so.HighestBid, dymnstypes.TypeAlias); err != nil {
					logger.Error("failed to refund bid for a force-to-cancel sell order", "asset-id", so.AssetId, "error", err)
					return err
				}
				e.DeleteSellOrder(ctx, so.AssetId, dymnstypes.TypeAlias)
				return nil
			}

			if err := e.CompleteAliasSellOrder(ctx, so.AssetId); err != nil {
				logger.Error("failed to complete sell order", "asset-id", so.AssetId, "error", err)
				return err
			}

			return nil
		})

		if errApplyStateChange == nil {
			aSoe.Remove(so.AssetId)
		}
	}

	if err := e.SetActiveSellOrdersExpiration(ctx, aSoe, dymnstypes.TypeAlias); err != nil {
		e.Keeper.Logger(ctx).Error("failed to update active SO expiry", "error", err)
		return err
	}

	return nil
}

// getFinishedSellOrders returns the finished Sell-Orders for the asset type.
// Finished Sell-Orders are the Sell-Orders that have expired or completed by having a bid.
func (e epochHooks) getFinishedSellOrders(ctx sdk.Context, assetType dymnstypes.AssetType, logger log.Logger) (
	finishedSellOrders []dymnstypes.SellOrder,
	aSoe *dymnstypes.ActiveSellOrdersExpiration,
) {
	aSoe = e.Keeper.GetActiveSellOrdersExpiration(ctx, assetType)
	if len(aSoe.Records) < 1 {
		return
	}

	blockEpochUTC := ctx.BlockTime().Unix()
	logger = logger.With("asset-type", assetType.FriendlyString(), "time", blockEpochUTC)

	invalidRecordsToRemove := make([]string, 0)

	for i, record := range aSoe.Records {
		if record.ExpireAt > blockEpochUTC {
			// skip not expired ones
			continue
		}

		so := e.GetSellOrder(ctx, record.AssetId, assetType)

		if so == nil {
			// remove the invalid entry
			invalidRecordsToRemove = append(invalidRecordsToRemove, record.AssetId)
			continue
		}

		if !so.HasFinished(blockEpochUTC) {
			// invalid entry
			logger.Error("sell order has not finished", "asset-id", record.AssetId, "expiry", record.ExpireAt)

			aSoe.Records[i].ExpireAt = so.ExpireAt // correct it
			continue
		}

		finishedSellOrders = append(finishedSellOrders, *so)
	}

	for _, name := range invalidRecordsToRemove {
		aSoe.Remove(name)
	}

	// guarantee of the execution order
	sort.Slice(finishedSellOrders, func(i, j int) bool {
		return finishedSellOrders[i].AssetId < finishedSellOrders[j].AssetId
	})

	return
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
