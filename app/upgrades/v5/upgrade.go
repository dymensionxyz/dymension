package v5

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/math"
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
	lockuptypes "github.com/dymensionxyz/dymension/v3/x/lockup/types"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	rollappmoduletypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencerkeeper "github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
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

		/* ---------------------------- store migrations ---------------------------- */
		// move params from params keeper to each module's store
		migrateDeprecatedParamsKeeperSubspaces(ctx, keepers)
		// Initialize endorsements for existing rollapps
		err = migrateEndorsements(ctx, keepers.IncentivesKeeper, keepers.SponsorshipKeeper)
		if err != nil {
			return nil, fmt.Errorf("migrate endorsements: %w", err)
		}

		/* ----------------------------- params updates ----------------------------- */
		// Incentives module params migration
		migrateAndUpdateIncentivesParams(ctx, keepers)

		// lockup module params migrations
		migrateAndUpdateLockupParams(ctx, keepers)

		// IRO module params migration
		updateIROParams(ctx, keepers.IROKeeper)

		// GAMM module params migration
		updateGAMMParams(ctx, keepers.GAMMKeeper)

		// fix V50 x/gov params
		updateGovParams(ctx, keepers.GovKeeper)

		updateRollappParams(ctx, keepers.RollappKeeper)

		updateSequencerParams(ctx, keepers.SequencerKeeper)

		migrateSequencers(ctx, keepers.SequencerKeeper)

		// Start running the module migrations
		logger.Debug("running module migrations ...")
		return migrations, nil
	}
}

func migrateAndUpdateIncentivesParams(ctx sdk.Context, keepers *upgrades.UpgradeKeepers) {
	// Incentives module
	incentivesSubspace, ok := keepers.ParamsKeeper.GetSubspace(incentives.ModuleName)
	if !ok {
		incentivesSubspace = keepers.ParamsKeeper.Subspace(incentives.ModuleName).WithKeyTable(incentives.ParamKeyTable())
	}
	var incentivesParams incentives.Params
	incentivesSubspace.GetParamSetIfExists(ctx, &incentivesParams)

	newParams := incentivestypes.NewParams(
		incentivesParams.DistrEpochIdentifier,
		incentivesParams.CreateGaugeBaseFee,
		incentivesParams.AddToGaugeBaseFee,
		incentivesParams.AddDenomFee,
		/* ------------------------------- new params ------------------------------- */
		incentivestypes.DefaultMinValueForDistr,  // Default to 0.01 DYM
		incentivestypes.DefaultRollappGaugesMode, // Default to active rollapps only
	)

	keepers.IncentivesKeeper.SetParams(ctx, newParams)
}

// Lockup module
func migrateAndUpdateLockupParams(ctx sdk.Context, keepers *upgrades.UpgradeKeepers) {
	lockupSubspace, ok := keepers.ParamsKeeper.GetSubspace(lockup.ModuleName)
	if !ok {
		lockupSubspace = keepers.ParamsKeeper.Subspace(lockup.ModuleName).WithKeyTable(lockup.ParamKeyTable())
	}
	var lockupParams lockup.Params
	lockupSubspace.GetParamSetIfExists(ctx, &lockupParams)

	newParams := lockuptypes.NewParams(
		lockupParams.ForceUnlockAllowedAddresses,
		/* ------------------------------- new params ------------------------------- */
		lockuptypes.DefaultLockFee, // Default to 0.05 DYM
		24*time.Hour,               // Minimum lock duration is 24 hours
	)
	keepers.LockupKeeper.SetParams(ctx, newParams)
}

func updateGAMMParams(ctx sdk.Context, k *gammkeeper.Keeper) {
	params := k.GetParams(ctx)

	for _, coin := range params.PoolCreationFee {
		params.AllowedPoolCreationDenoms = append(params.AllowedPoolCreationDenoms, coin.Denom)
	}
	k.SetParams(ctx, params)
}

func updateIROParams(ctx sdk.Context, k *irokeeper.Keeper) {
	params := k.GetParams(ctx)
	defParams := irotypes.DefaultParams()

	params.MinLiquidityPart = defParams.MinLiquidityPart                                     // default: at least 40% goes to the liquidity pool
	params.MinVestingDuration = defParams.MinVestingDuration                                 // default: min 7 days
	params.MinVestingStartTimeAfterSettlement = defParams.MinVestingStartTimeAfterSettlement // default: no enforced minimum by default

	k.SetParams(ctx, params)
}

func updateGovParams(ctx sdk.Context, k *govkeeper.Keeper) {
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

// Create endorsment objects for existing rollapps and endorser positions for existing votes
// we iterate over rollapp gauges as we have one per rollapp
func migrateEndorsements(ctx sdk.Context, incentivesKeeper *incentiveskeeper.Keeper, sponsorshipKeeper *sponsorshipkeeper.Keeper) error {
	gauges := incentivesKeeper.GetGauges(ctx)
	distr, err := sponsorshipKeeper.GetDistribution(ctx)
	if err != nil {
		return fmt.Errorf("get distribution: %w", err)
	}

	// This is a temporary map for a fast lookup of gauge total voting power
	powerByGauge := make(map[uint64]math.Int, len(distr.Gauges))
	for _, gauge := range distr.Gauges {
		powerByGauge[gauge.GaugeId] = gauge.Power
	}

	// Map to store endorsement by gauge ID for efficient lookup
	endorsementsByGaugeId := make(map[uint64]sponsorshiptypes.Endorsement)

	// 1. Create Endorsement objects for existing RollapApps
	for _, gauge := range gauges {
		if rollappGauge := gauge.GetRollapp(); rollappGauge != nil {
			// Check if the gauge has any voting power. Total voting power is the initial
			// number of shares in the respective endorsement gauge.
			power, ok := powerByGauge[gauge.Id]
			if !ok {
				// If a RA gauge does not have any power, it's fine; use 0.
				// It means there is no voting power cast to this rollapp.
				power = math.ZeroInt()
			}

			// Create an endorsement for this rollapp gauge
			p := math.LegacyNewDecFromInt(power)
			endorsement := sponsorshiptypes.NewEndorsement(rollappGauge.RollappId, gauge.Id, p)

			// Store in map for efficient lookup later
			endorsementsByGaugeId[gauge.Id] = endorsement

			err := sponsorshipKeeper.SaveEndorsement(ctx, endorsement)
			if err != nil {
				return fmt.Errorf("failed to save endorsement: %w", err)
			}

			ctx.Logger().Info(fmt.Sprintf("Created endorsement for rollapp %s with gauge %d", rollappGauge.RollappId, gauge.Id))
		}
	}

	// 2. Create EndorserPosition objects for existing votes
	err = sponsorshipKeeper.IterateVotes(ctx, func(voter sdk.AccAddress, vote sponsorshiptypes.Vote) (stop bool, err error) {
		// Convert vote to distribution to get individual gauge power
		distribution := vote.ToDistribution()

		for _, gauge := range distribution.Gauges {
			// Check if this gauge is a rollapp gauge by looking up in our endorsements map
			endorsement, isRollappGauge := endorsementsByGaugeId[gauge.GaugeId]
			if !isRollappGauge {
				continue // Skip non-rollapp gauges
			}

			// Create endorser position with the voting power as shares
			shares := math.LegacyNewDecFromInt(gauge.Power)
			endorserPosition := sponsorshiptypes.EndorserPosition{
				Shares:              shares,
				LastSeenAccumulator: endorsement.Accumulator, // Start with current accumulator (zero for new endorsements)
			}

			// Save the endorser position
			err = sponsorshipKeeper.SaveEndorserPosition(ctx, voter, endorsement.RollappId, endorserPosition)
			if err != nil {
				return true, fmt.Errorf("save endorser position for user %s and rollapp %s: %w", voter.String(), endorsement.RollappId, err)
			}

			ctx.Logger().Info(fmt.Sprintf("Created endorser position for user %s and rollapp %s with %s shares",
				voter.String(), endorsement.RollappId, shares.String()))
		}

		return false, nil // Continue iteration
	})

	if err != nil {
		return fmt.Errorf("iterate votes: %w", err)
	}

	return nil
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
		eibcParams.TimeoutFee,
		eibcParams.ErrackFee,
	))

	// DymNS module
	dymnsSubspace := keepers.ParamsKeeper.Subspace(dymns.ModuleName)
	dymnsSubspace = dymnsSubspace.WithKeyTable(dymns.ParamKeyTable())
	var dymnsParams dymns.Params
	dymnsSubspace.GetParamSetIfExists(ctx, &dymnsParams)
	keepers.DymNSKeeper.SetParams(ctx, dymnstypes.NewParams(
		dymnstypes.PriceParams{
			NamePriceSteps:         dymnsParams.Price.NamePriceSteps,
			AliasPriceSteps:        dymnsParams.Price.AliasPriceSteps,
			PriceExtends:           dymnsParams.Price.PriceExtends,
			PriceDenom:             dymnsParams.Price.PriceDenom,
			MinOfferPrice:          dymnsParams.Price.MinOfferPrice,
			MinBidIncrementPercent: dymnsParams.Price.MinBidIncrementPercent,
		},
		dymnstypes.ChainsParams{
			AliasesOfChainIds: func() []dymnstypes.AliasesOfChainId {
				converted := make([]dymnstypes.AliasesOfChainId, len(dymnsParams.Chains.AliasesOfChainIds))
				for i, v := range dymnsParams.Chains.AliasesOfChainIds {
					converted[i] = dymnstypes.AliasesOfChainId{
						ChainId: v.ChainId,
						Aliases: v.Aliases,
					}
				}
				return converted
			}(),
		},
		dymnstypes.MiscParams{
			EndEpochHookIdentifier: dymnsParams.Misc.EndEpochHookIdentifier,
			GracePeriodDuration:    dymnsParams.Misc.GracePeriodDuration,
			SellOrderDuration:      dymnsParams.Misc.SellOrderDuration,
			EnableTradingName:      dymnsParams.Misc.EnableTradingName,
			EnableTradingAlias:     dymnsParams.Misc.EnableTradingAlias,
		},
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

	// Streamer module
	streamerSubspace := keepers.ParamsKeeper.Subspace(streamer.ModuleName)
	streamerSubspace = streamerSubspace.WithKeyTable(streamer.ParamKeyTable())
	var streamerParams streamer.Params
	streamerSubspace.GetParamSetIfExists(ctx, &streamerParams)
	keepers.StreamerKeeper.SetParams(ctx, streamermoduletypes.NewParams(
		streamerParams.MaxIterationsPerBlock,
	))
}

const (
	slowBlockDuration                    = 6
	fastBlockDuration                    = 1
	BlockSpeedup                         = slowBlockDuration / fastBlockDuration
	slowBlocksParamDisputePeriod         = 120960
	fastBlocksParamDisputePeriod         = slowBlocksParamDisputePeriod * BlockSpeedup
	slowBlocksParamLivenessSlashBlocks   = 7200 // 12 hrs
	fastBlocksParamLivenessSlashBlocks   = slowBlocksParamLivenessSlashBlocks * BlockSpeedup
	slowBlocksParamLivenessSlashInterval = 600 // 1hr
	slashIntervalMul                     = 6   // 1 -> 6 hrs
	fastBlocksParamLivenessSlashInterval = slowBlocksParamLivenessSlashInterval * BlockSpeedup * slashIntervalMul
)

var newLivenessSlashMinMultiplier = math.LegacyMustNewDecFromStr("0.02")

const (
	newPenaltyLiveness             = uint64(300)
	NewPenaltyKickThreshold        = uint64(600)
	newPenaltyReductionStateUpdate = uint64(150)
)

func updateRollappParams(ctx sdk.Context, k *rollappkeeper.Keeper) {
	// 1. params
	params := k.GetParams(ctx)
	params.DisputePeriodInBlocks = fastBlocksParamDisputePeriod
	params.LivenessSlashBlocks = fastBlocksParamLivenessSlashBlocks
	params.LivenessSlashInterval = fastBlocksParamLivenessSlashInterval
	k.SetParams(ctx, params)

	// 2. other state
	// (other migration for dispute not needed because finalization is computed based on stored creation height)
	migrateLivenessEvents(ctx, k)
}

func migrateLivenessEvents(ctx sdk.Context, k *rollappkeeper.Keeper) {
	events := k.GetLivenessEvents(ctx, nil)
	for _, e := range events {
		diff := e.HubHeight - ctx.BlockHeight()
		if diff < 0 {
			panic("assumed no liveness events in the past") // (zero is fine)
		}
		k.DelLivenessEvents(ctx, e.HubHeight, e.RollappId) // we can delete 'both' since there is only one kind currently
		e.HubHeight = ctx.BlockHeight() + diff*BlockSpeedup
		k.PutLivenessEvent(ctx, e)
	}
}

func updateSequencerParams(ctx sdk.Context, k *sequencerkeeper.Keeper) {
	params := k.GetParams(ctx)
	params.LivenessSlashMinMultiplier = newLivenessSlashMinMultiplier
	params.SetPenaltyLiveness(newPenaltyLiveness)
	params.SetPenaltyKickThreshold(NewPenaltyKickThreshold)
	params.SetPenaltyReductionStateUpdate(newPenaltyReductionStateUpdate)
	k.SetParams(ctx, params)
}

func migrateSequencers(ctx sdk.Context, k *sequencerkeeper.Keeper) {
	sequencers := k.AllSequencers(ctx)
	for _, s := range sequencers {
		if NewPenaltyKickThreshold < s.GetPenalty() {
			s.SetPenalty(NewPenaltyKickThreshold)
			k.SetSequencer(ctx, s)
		}
	}
}
