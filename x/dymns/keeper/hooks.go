package keeper

import (
	"sort"

	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
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
// Business logic is to prune historical sell orders and clearing preserved registration.
func (e epochHooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	params := e.GetParams(ctx)

	if epochIdentifier != params.Misc.BeginEpochHookIdentifier {
		return nil
	}

	e.Keeper.Logger(ctx).Info("DymNS hook Before-Epoch-Start: triggered", "epoch-number", epochNumber, "epoch-identifier", epochIdentifier)

	if err := e.processCleanupHistoricalSellOrders(ctx, epochIdentifier, epochNumber, params); err != nil {
		return err
	}

	var updatedParams bool
	params, updatedParams = e.processCleanupPreservedRegistration(ctx, epochIdentifier, epochNumber, params)
	if updatedParams {
		if err := e.SetParams(ctx, params); err != nil {
			return err
		}
	}

	return nil
}

// processCleanupHistoricalSellOrders prunes historical Sell-Orders records store when reservation date passed.
func (e epochHooks) processCleanupHistoricalSellOrders(ctx sdk.Context, epochIdentifier string, epochNumber int64, params dymnstypes.Params) error {
	dk := e.Keeper

	/**
	We use this method instead of iterating through all historical sell orders.
	It helps reduce number of IO needed to read all historical sell orders.
	*/
	minExpiryPerDymNameRecords := dk.GetMinExpiryOfAllHistoricalSellOrders(ctx)
	if len(minExpiryPerDymNameRecords) < 1 {
		return nil
	}

	cleanBeforeEpochUTC := ctx.BlockTime().Add(-1 * params.Misc.PreservedClosedSellOrderDuration).Unix()

	var cleanupHistoricalForDymNames []string
	for _, minExpiryPerDymName := range minExpiryPerDymNameRecords {
		if minExpiryPerDymName.MinExpiry > cleanBeforeEpochUTC {
			continue
		}

		cleanupHistoricalForDymNames = append(cleanupHistoricalForDymNames, minExpiryPerDymName.DymName)
	}
	if len(cleanupHistoricalForDymNames) < 1 {
		return nil
	}

	e.Keeper.Logger(ctx).Info(
		"DymNS hook Before-Epoch-Start: processing cleanup historical sell orders",
		"count", len(cleanupHistoricalForDymNames),
		"epoch-number", epochNumber, "epoch-identifier", epochIdentifier,
	)

	// ensure deterministic order, this action should be done regardless of the materials was sorted or not
	sort.Strings(cleanupHistoricalForDymNames)

	for _, dymName := range cleanupHistoricalForDymNames {
		list := dk.GetHistoricalSellOrders(ctx, dymName)
		if len(list) < 1 {
			dk.SetMinExpiryHistoricalSellOrder(ctx, dymName, 0)
			continue
		}

		var keepList []dymnstypes.SellOrder
		for _, hso := range list {
			if hso.ExpireAt > cleanBeforeEpochUTC {
				keepList = append(keepList, hso)
			}
		}

		if len(keepList) == 0 {
			dk.DeleteHistoricalSellOrders(ctx, dymName)
			dk.SetMinExpiryHistoricalSellOrder(ctx, dymName, 0)
			continue
		}

		newMinExpiry := keepList[0].ExpireAt
		for _, hso := range keepList {
			if hso.ExpireAt < newMinExpiry {
				newMinExpiry = hso.ExpireAt
			}
		}
		dk.SetMinExpiryHistoricalSellOrder(ctx, dymName, newMinExpiry)

		if len(keepList) != len(list) {
			hso := dymnstypes.HistoricalSellOrders{
				SellOrders: keepList,
			}
			dk.SetHistoricalSellOrders(ctx, dymName, hso)
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
// Business logic is to move expired Sell-Orders to historical
// and if Sell-Order has a winner, complete that SO.
func (e epochHooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	params := e.GetParams(ctx)

	if epochIdentifier != params.Misc.EndEpochHookIdentifier {
		return nil
	}

	e.Keeper.Logger(ctx).Info("DymNS hook After-Epoch-End: triggered", "epoch-number", epochNumber, "epoch-identifier", epochIdentifier)

	return e.processActiveSellOrders(ctx, epochIdentifier, epochNumber)
}

// processActiveSellOrders moves expired Sell-Orders to historical and completes Sell-Orders with winners.
func (e epochHooks) processActiveSellOrders(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	dk := e.Keeper

	aSoe := dk.GetActiveSellOrdersExpiration(ctx)
	nowEpochUTC := ctx.BlockTime().Unix()
	var finishedSOs []dymnstypes.SellOrder
	if len(aSoe.Records) > 0 {
		invalidRecordsToRemove := make([]string, 0)

		for i, record := range aSoe.Records {
			if record.ExpireAt > nowEpochUTC {
				// skip not expired ones
				continue
			}

			so := dk.GetSellOrder(ctx, record.GoodsId)

			if so == nil {
				// remove the invalid entry
				invalidRecordsToRemove = append(invalidRecordsToRemove, record.GoodsId)
				continue
			}

			if !so.HasFinished(nowEpochUTC) {
				// invalid entry
				dk.Logger(ctx).Error(
					"DymNS hook After-Epoch-End: sell order has not finished",
					"name", record.GoodsId, "expiry", record.ExpireAt, "now", nowEpochUTC,
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
	}

	if len(finishedSOs) < 1 {
		// skip updating store
		return nil
	}

	sort.Slice(finishedSOs, func(i, j int) bool {
		return finishedSOs[i].GoodsId < finishedSOs[j].GoodsId
	})

	dk.Logger(ctx).Info(
		"DymNS hook After-Epoch-End: processing finished SOs", "count", len(finishedSOs),
		"epoch-number", epochNumber, "epoch-identifier", epochIdentifier,
	)

	for _, so := range finishedSOs {
		aSoe.Remove(so.GoodsId)

		if so.HighestBid != nil {
			if err := dk.CompleteSellOrder(ctx, so.GoodsId); err != nil {
				dk.Logger(ctx).Error(
					"DymNS hook After-Epoch-End: failed to complete sell order",
					"name", so.GoodsId, "expiry", so.ExpireAt, "now", nowEpochUTC,
					"epoch-number", epochNumber, "epoch-identifier", epochIdentifier,
					"error", err,
				)
				return err
			}
			continue
		}

		// no bid placed, it just a normal expiry without winner,
		// in this case, just move to history
		if err := dk.MoveSellOrderToHistorical(ctx, so.GoodsId); err != nil {
			dk.Logger(ctx).Error(
				"DymNS hook After-Epoch-End: failed to move expired sell order to historical",
				"name", so.GoodsId, "expiry", so.ExpireAt, "now", nowEpochUTC,
				"epoch-number", epochNumber, "epoch-identifier", epochIdentifier,
				"error", err,
			)
			return err
		}
	}

	if err := dk.SetActiveSellOrdersExpiration(ctx, aSoe); err != nil {
		dk.Logger(ctx).Error(
			"DymNS hook After-Epoch-End: failed to update active SO expiry",
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

type RollAppHooks interface {
	// TODO DymNS: connect hooks and remove this interface

	RollappCreated(ctx sdk.Context, rollappID, alias string, creatorAddr sdk.AccAddress) error
}

func (k Keeper) GetRollAppHooks() RollAppHooks {
	return rollappHooks{
		Keeper: k,
	}
}

type rollappHooks struct {
	Keeper
}

var _ RollAppHooks = rollappHooks{}

func (h rollappHooks) RollappCreated(ctx sdk.Context, rollappID, alias string, creatorAddr sdk.AccAddress) error {
	// TODO DymNS: if alias is empty, do nothing

	if !h.Keeper.IsRollAppId(ctx, rollappID) {
		// ensure RollApp record is set
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "not a RollApp chain-id: %s", rollappID)
	}

	canUseAlias, err := h.CanUseAliasForNewRegistration(ctx, alias)
	if err != nil {
		return err
	}

	if !canUseAlias {
		return errorsmod.Wrapf(gerrc.ErrAlreadyExists, "alias already in use or preserved: %s", alias)
	}

	if err := h.Keeper.SetAliasForRollAppId(ctx, rollappID, alias); err != nil {
		return err
	}

	params := h.Keeper.GetParams(ctx)
	aliasCost := sdk.NewCoins(
		sdk.NewCoin(
			params.Price.PriceDenom, params.Price.GetAliasPrice(alias),
		),
	)

	if err := h.bankKeeper.SendCoinsFromAccountToModule(ctx,
		creatorAddr,
		dymnstypes.ModuleName,
		aliasCost,
	); err != nil {
		return err
	}

	if err := h.bankKeeper.BurnCoins(ctx, dymnstypes.ModuleName, aliasCost); err != nil {
		return err
	}

	return nil
}
