package keeper

import (
	"errors"

	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"

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
	activeSellOrdersExpiration := e.Keeper.GetActiveSellOrdersExpiration(ctx, dymnstypes.TypeName)

	finishedSOs := e.getFinishedSellOrders(ctx, activeSellOrdersExpiration, dymnstypes.TypeName, logger)

	if len(finishedSOs) < 1 {
		return nil
	}

	logger.Info("processing finished SOs.", "count", len(finishedSOs))

	for _, so := range finishedSOs {
		// each order should be processed in a branched context, if error, discard the state change
		// and process next order, to prevent chain reaction when an individual order failed to process
		errApplyStateChange := osmoutils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
			if so.HighestBid == nil {
				e.DeleteSellOrder(ctx, so.AssetId, dymnstypes.TypeName)
				return nil
			}

			if err := e.CompleteDymNameSellOrder(ctx, so.AssetId); err != nil {
				return err
			}

			return nil
		})

		if errApplyStateChange == nil {
			activeSellOrdersExpiration.Remove(so.AssetId)
		} else {
			logger.Error(
				"failed to process finished sell order.", "asset-id", so.AssetId,
				"bid", so.HighestBid != nil,
				"error", errApplyStateChange,
			)
		}
	}

	if err := e.SetActiveSellOrdersExpiration(ctx, activeSellOrdersExpiration, dymnstypes.TypeName); err != nil {
		return errorsmod.Wrap(errors.Join(gerrc.ErrInternal, err), "failed to update active SO expiry")
	}

	return nil
}

// processActiveAliasSellOrders process the finished Alias Sell-Orders.
// Sell-Order will be deleted.
// If the Sell-Order has a winner, the Alias linking will be updated.
// Sell-Orders for the aliases that are prohibited to trade will be force cancelled,
// please read the code body for more information about what it is.
func (e epochHooks) processActiveAliasSellOrders(ctx sdk.Context, logger log.Logger) error {
	activeSellOrdersExpiration := e.Keeper.GetActiveSellOrdersExpiration(ctx, dymnstypes.TypeAlias)

	finishedSOs := e.getFinishedSellOrders(ctx, activeSellOrdersExpiration, dymnstypes.TypeAlias, logger)

	if len(finishedSOs) < 1 {
		return nil
	}

	logger.Info("processing finished SOs.", "count", len(finishedSOs))

	prohibitedToTradeAliases := e.GetAllAliasAndChainIdInParams(ctx)

	for _, so := range finishedSOs {
		// each order should be processed in a branched context, if error, discard the state change
		// and process next order, to prevent chain reaction when an individual order failed to process
		errApplyStateChange := osmoutils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
			if so.HighestBid == nil {
				e.DeleteSellOrder(ctx, so.AssetId, dymnstypes.TypeAlias)
				return nil
			}

			/**
			For the Sell-Orders which the assets are prohibited to trade,
			the Sell-Order will be force cancelled and the bids will be refunded.

			Why some aliases are prohibited to trade? And what are they?
			In module params, there is a list of alias mapping for some external well-known chains.
			So those aliases are considered as reserved for the external chains,
			therefor trading is not allowed.

			Why can someone own a prohibited alias?
			An alias can be bought before the reservation was made.
			But when the alias becomes reserved for the external well-known chains,
			the alias will be prohibited to trade.

			Why can someone place a Sell-Order for the prohibited alias?
			When a Sell-Order created before the reservation was made.
			*/
			if _, forceCancel := prohibitedToTradeAliases[so.AssetId]; forceCancel {
				// Sell-Order will be force cancelled and refund bids if any,
				// when the alias is prohibited to trade
				if err := e.RefundBid(ctx, *so.HighestBid, dymnstypes.TypeAlias); err != nil {
					return err
				}
				e.DeleteSellOrder(ctx, so.AssetId, dymnstypes.TypeAlias)
				return nil
			}

			if err := e.CompleteAliasSellOrder(ctx, so.AssetId); err != nil {
				return err
			}

			return nil
		})

		if errApplyStateChange == nil {
			activeSellOrdersExpiration.Remove(so.AssetId)
		} else {
			_, forceCancel := prohibitedToTradeAliases[so.AssetId]
			logger.Error(
				"failed to process finished sell order.", "asset-id", so.AssetId,
				"bid", so.HighestBid != nil, "prohibited", forceCancel,
				"error", errApplyStateChange,
			)
		}
	}

	if err := e.SetActiveSellOrdersExpiration(ctx, activeSellOrdersExpiration, dymnstypes.TypeAlias); err != nil {
		return errorsmod.Wrap(errors.Join(gerrc.ErrInternal, err), "failed to update active SO expiry")
	}

	return nil
}

// getFinishedSellOrders returns the finished Sell-Orders for the asset type.
// Finished Sell-Orders are the Sell-Orders that have expired.
// Expired sell-orders can either have a bid or not. In both cases we consider them as `finished`.
func (e epochHooks) getFinishedSellOrders(
	ctx sdk.Context,
	activeSellOrdersExpiration *dymnstypes.ActiveSellOrdersExpiration, assetType dymnstypes.AssetType,
	logger log.Logger,
) (finishedSellOrders []dymnstypes.SellOrder) {
	blockEpochUTC := ctx.BlockTime().Unix()

	for _, record := range activeSellOrdersExpiration.Records {
		if record.ExpireAt > blockEpochUTC {
			// skip not expired ones
			continue
		}

		so := e.GetSellOrder(ctx, record.AssetId, assetType)

		if so == nil {
			logger.Error(
				"invalid entry on Active Sell Order Expiration records: Sell Order not found.",
				"asset-id", record.AssetId, "asset-type", assetType.PrettyName(),
			)
			// ignore the invalid entries for now, invariant will catch it
			continue
		}

		if !so.HasFinished(blockEpochUTC) {
			logger.Error(
				"invalid entry on Active Sell Order Expiration records: Sell Order not yet finished.",
				"asset-id", record.AssetId, "asset-type", assetType.PrettyName(),
				"record-expiry", record.ExpireAt, "actual-expiry", so.ExpireAt,
			)
			// ignore the invalid entries for now, invariant will catch it
			continue
		}

		finishedSellOrders = append(finishedSellOrders, *so)
	}

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

func (h rollappHooks) AfterStateFinalized(_ sdk.Context, _ string, _ *rollapptypes.StateInfo) error {
	return nil
}

func (h rollappHooks) FraudSubmitted(_ sdk.Context, _ string, _ uint64, _ string) error {
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
