package v5

import (
	"context"
	"fmt"
	"time"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"

	"github.com/dymensionxyz/dymension/v3/app/upgrades"
	"github.com/dymensionxyz/dymension/v3/app/upgrades/v5/types/delayedack"
	"github.com/dymensionxyz/dymension/v3/app/upgrades/v5/types/dymns"
	"github.com/dymensionxyz/dymension/v3/app/upgrades/v5/types/eibc"
	"github.com/dymensionxyz/dymension/v3/app/upgrades/v5/types/incentives"
	"github.com/dymensionxyz/dymension/v3/app/upgrades/v5/types/lockup"
	"github.com/dymensionxyz/dymension/v3/app/upgrades/v5/types/rollapp"
	"github.com/dymensionxyz/dymension/v3/app/upgrades/v5/types/streamer"
	delayedacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	eibcmoduletypes "github.com/dymensionxyz/dymension/v3/x/eibc/types"
	incentiveskeeper "github.com/dymensionxyz/dymension/v3/x/incentives/keeper"
	incentivestypes "github.com/dymensionxyz/dymension/v3/x/incentives/types"
	irokeeper "github.com/dymensionxyz/dymension/v3/x/iro/keeper"
	irotypes "github.com/dymensionxyz/dymension/v3/x/iro/types"
	lockupkeeper "github.com/dymensionxyz/dymension/v3/x/lockup/keeper"
	lockuptypes "github.com/dymensionxyz/dymension/v3/x/lockup/types"
	rollappmoduletypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sponsorshipkeeper "github.com/dymensionxyz/dymension/v3/x/sponsorship/keeper"
	sponsorshiptypes "github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
	streamermoduletypes "github.com/dymensionxyz/dymension/v3/x/streamer/types"
	gammkeeper "github.com/osmosis-labs/osmosis/v15/x/gamm/keeper"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v5
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *upgrades.UpgradeKeepers,
) upgradetypes.UpgradeHandler {
	return func(goCtx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(goCtx)
		logger := ctx.Logger().With("upgrade", UpgradeName)

		// Run migrations before applying any other state changes.
		// NOTE: DO NOT PUT ANY STATE CHANGES BEFORE RunMigrations().
		// IRO store upgraded through module migrations
		// x/txfees and x/iro upgraded through module migrations
		migrations, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return nil, err
		}

		// instead of writing migrartion in each moudle, we do it here in a centralized place
		migrateDeprecatedParamsKeeperSubspaces(ctx, keepers)

		// Incentives module params migration
		migrateIncentivesParams(ctx, keepers.IncentivesKeeper)

		// lockup module params migrations
		migrateLockupParams(ctx, keepers.LockupKeeper)

		// IRO module params migration
		migrateIROParams(ctx, keepers.IROKeeper)

		// GAMM module params migration
		migrateGAMMParams(ctx, keepers.GAMMKeeper)

		// fix V50 x/gov params
		migrateGovParams(ctx, keepers.GovKeeper)

		// Initialize endorsements for existing rollapps
		migrateEndorsements(ctx, keepers.IncentivesKeeper, keepers.SponsorshipKeeper)

		// Start running the module migrations
		logger.Debug("running module migrations ...")
		return migrations, nil
	}
}

func migrateIncentivesParams(ctx sdk.Context, k *incentiveskeeper.Keeper) {
	params := k.GetParams(ctx)

	// default mode is active rollapps only
	params.RollappGaugesMode = incentivestypes.DefaultRollappGaugesMode

	// set default min value for distribution (0.01 DYM)
	params.MinValueForDistribution = incentivestypes.DefaultMinValueForDistr

	k.SetParams(ctx, params)
}

func migrateLockupParams(ctx sdk.Context, k *lockupkeeper.Keeper) {
	params := k.GetParams(ctx)

	params.LockCreationFee = lockuptypes.DefaultLockFee
	params.MinLockDuration = 24 * time.Hour

	k.SetParams(ctx, params)
}

func migrateGAMMParams(ctx sdk.Context, k *gammkeeper.Keeper) {
	params := k.GetParams(ctx)

	for _, coin := range params.PoolCreationFee {
		params.AllowedPoolCreationDenoms = append(params.AllowedPoolCreationDenoms, coin.Denom)
	}
	k.SetParams(ctx, params)
}

func migrateIROParams(ctx sdk.Context, k *irokeeper.Keeper) {
	params := k.GetParams(ctx)
	defParams := irotypes.DefaultParams()

	params.MinLiquidityPart = defParams.MinLiquidityPart
	params.MinVestingDuration = defParams.MinVestingDuration
	params.MinVestingStartTimeAfterSettlement = defParams.MinVestingStartTimeAfterSettlement

	k.SetParams(ctx, params)
}

func migrateGovParams(ctx sdk.Context, k *govkeeper.Keeper) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		panic(err)
	}

	// expedited min deposit is 5 times the min deposit
	params.ExpeditedMinDeposit = sdk.NewCoins(sdk.NewCoin(params.MinDeposit[0].Denom, params.MinDeposit[0].Amount.MulRaw(5)))

	err = k.Params.Set(ctx, params)
	if err != nil {
		panic(err)
	}
}

// Create endorsment objects for existing rollapps
// we iterate over rollapp gauges as we have one per rollapp
func migrateEndorsements(ctx sdk.Context, incentivesKeeper *incentiveskeeper.Keeper, sponsorshipKeeper *sponsorshipkeeper.Keeper) {
	gauges := incentivesKeeper.GetGauges(ctx)
	for _, gauge := range gauges {
		if rollappGauge := gauge.GetRollapp(); rollappGauge != nil {
			// Create endorsement for this rollapp gauge
			endorsement := sponsorshiptypes.NewEndorsement(rollappGauge.RollappId, gauge.Id)
			err := sponsorshipKeeper.SaveEndorsement(ctx, endorsement)
			if err != nil {
				panic(fmt.Errorf("failed to save endorsement: %w", err))
			}
			ctx.Logger().Info(fmt.Sprintf("Created endorsement for rollapp %s with gauge %d", rollappGauge.RollappId, gauge.Id))
		}
	}
}

// Get params from subspaces and set them using each keeper's SetParams method
func migrateDeprecatedParamsKeeperSubspaces(ctx sdk.Context, keepers *upgrades.UpgradeKeepers) {
	// DelayedAck module
	delayedackSubspace := keepers.ParamsKeeper.Subspace(delayedack.ModuleName)
	delayedackSubspace = delayedackSubspace.WithKeyTable(delayedack.ParamKeyTable())
	var delayedackParams delayedack.Params
	delayedackSubspace.GetParamSetIfExists(ctx, &delayedackParams)
	keepers.DelayedAckKeeper.SetParams(ctx, delayedacktypes.NewParams(
		delayedackParams.EpochIdentifier,
		delayedackParams.BridgingFee,
		int(delayedackParams.DeletePacketsEpochLimit),
	))

	// EIBC module
	eibcSubspace := keepers.ParamsKeeper.Subspace(eibc.ModuleName)
	eibcSubspace = eibcSubspace.WithKeyTable(eibc.ParamKeyTable())
	var eibcParams eibc.Params
	eibcSubspace.GetParamSetIfExists(ctx, &eibcParams)
	keepers.EIBCKeeper.SetParams(ctx, eibcmoduletypes.NewParams(
		eibcParams.EpochIdentifier,
		eibcParams.BridgeFee,
		eibcParams.BridgeFee,
	))

	// DymNS module
	dymnsSubspace := keepers.ParamsKeeper.Subspace(dymns.ModuleName)
	dymnsSubspace = dymnsSubspace.WithKeyTable(dymns.ParamKeyTable())
	var dymnsParams dymns.Params
	dymnsSubspace.GetParamSetIfExists(ctx, &dymnsParams)
	keepers.DymNSKeeper.SetParams(ctx, dymnstypes.NewParams(
		dymnstypes.PriceParams{
			PriceDenom: dymnsParams.Price.PriceDenom,
		},
		dymnstypes.ChainsParams{
			AliasesOfChainIds: make([]dymnstypes.AliasesOfChainId, len(dymnsParams.Chains.AliasesOfChainIds)),
		},
		dymnstypes.MiscParams{
			EndEpochHookIdentifier: dymnsParams.Misc.EndEpochHookIdentifier,
			GracePeriodDuration:    dymnsParams.Misc.GracePeriodDuration,
			SellOrderDuration:      dymnsParams.Misc.SellOrderDuration,
		},
	))

	// Incentives module
	incentivesSubspace := keepers.ParamsKeeper.Subspace(incentives.ModuleName)
	incentivesSubspace = incentivesSubspace.WithKeyTable(incentives.ParamKeyTable())
	var incentivesParams incentives.Params
	incentivesSubspace.GetParamSetIfExists(ctx, &incentivesParams)
	keepers.IncentivesKeeper.SetParams(ctx, incentivestypes.NewParams(
		incentivesParams.DistrEpochIdentifier,
		incentivesParams.CreateGaugeBaseFee,
		incentivesParams.AddToGaugeBaseFee,
		incentivesParams.AddDenomFee,
		incentivesParams.MinValueForDistribution,
		incentivestypes.Params_RollappGaugesModes(incentivesParams.RollappGaugesMode),
	))

	// Rollapp module
	rollappSubspace := keepers.ParamsKeeper.Subspace(rollapp.ModuleName)
	rollappSubspace = rollappSubspace.WithKeyTable(rollapp.ParamKeyTable())
	var rollappParams rollapp.Params
	rollappSubspace.GetParamSetIfExists(ctx, &rollappParams)
	keepers.RollappKeeper.SetParams(ctx, rollappmoduletypes.NewParams(
		rollappParams.DisputePeriodInBlocks,
		rollappParams.LivenessSlashBlocks,
		rollappParams.LivenessSlashInterval,
		rollappParams.AppRegistrationFee,
		rollappParams.MinSequencerBondGlobal,
	))

	// Lockup module
	lockupSubspace := keepers.ParamsKeeper.Subspace(lockup.ModuleName)
	lockupSubspace = lockupSubspace.WithKeyTable(lockup.ParamKeyTable())
	var lockupParams lockup.Params
	lockupSubspace.GetParamSetIfExists(ctx, &lockupParams)
	keepers.LockupKeeper.SetParams(ctx, lockuptypes.NewParams(
		lockupParams.ForceUnlockAllowedAddresses,
		lockupParams.LockCreationFee,
		lockupParams.MinLockDuration,
	))

	// Streamer module
	streamerSubspace := keepers.ParamsKeeper.Subspace(streamer.ModuleName)
	streamerSubspace = streamerSubspace.WithKeyTable(streamer.ParamKeyTable())
	var streamerParams streamer.Params
	streamerSubspace.GetParamSetIfExists(ctx, &streamerParams)
	keepers.StreamerKeeper.SetParams(ctx, streamermoduletypes.NewParams(
		streamerParams.MaxIterationsPerBlock,
	))
}

// func migrateParamsKeeper(ctx sdk.Context, keepers *upgrades.UpgradeKeepers) {
// app.GetSubspace(minttypes.ModuleName)

// // GetSubspace gets existing substore from keeper.
// func (a *AppKeepers) GetSubspace(moduleName string) paramstypes.Subspace {
// 	subspace, _ := a.ParamsKeeper.GetSubspace(moduleName)
// 	return subspace
// }
// }
