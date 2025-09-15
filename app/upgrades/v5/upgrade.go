package v5

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/math"
	circuitkeeper "cosmossdk.io/x/circuit/keeper"
	circuittypes "cosmossdk.io/x/circuit/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	ratelimitkeeper "github.com/cosmos/ibc-apps/modules/rate-limiting/v8/keeper"
	ratelimittypes "github.com/cosmos/ibc-apps/modules/rate-limiting/v8/types"

	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
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
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	rollappmoduletypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencerkeeper "github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
	sponsorshipkeeper "github.com/dymensionxyz/dymension/v3/x/sponsorship/keeper"
	sponsorshiptypes "github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
	streamermoduletypes "github.com/dymensionxyz/dymension/v3/x/streamer/types"
	gammkeeper "github.com/osmosis-labs/osmosis/v15/x/gamm/keeper"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	txfeeskeeper "github.com/osmosis-labs/osmosis/v15/x/txfees/keeper"
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

		// Migrate locks: set creation_timestamp (UpdatedAt) if not set
		if err := migrateLockTimestamps(ctx, keepers.LockupKeeper, keepers.IncentivesKeeper); err != nil {
			return nil, err
		}

		// Migrate gauges: set min lock age for perpetual asset gauges
		if err := migrateGaugeLockAges(ctx, keepers.IncentivesKeeper); err != nil {
			return nil, err
		}

		// add authorized circuit breaker
		addAuthorizedCircuitBreaker(ctx, keepers.CircuitBreakKeeper, keepers.AccountKeeper)

		/* ----------------------------- params updates ----------------------------- */
		// new IRO params
		updateIROParams(ctx, keepers.IROKeeper)

		// new GAMM params
		updateGAMMParams(ctx, keepers.GAMMKeeper)

		// new x/gov params
		updateGovParams(ctx, keepers.GovKeeper)

		// update txfees params
		updateTxfeesParams(ctx, keepers.TxfeesKeeper)

		// update params to fast block speed
		updateParamsToFastBlockSpeed(ctx, keepers)

		// fix x/sequencer liveness slash params
		updateSequencerParams(ctx, keepers.SequencerKeeper)
		migrateSequencers(ctx, keepers.SequencerKeeper)

		// Set up rate limiting parameters for existing channels
		err = setupRateLimitingParams(ctx, keepers.RateLimitingKeeper)
		if err != nil {
			return nil, fmt.Errorf("setup rate limiting params: %w", err)
		}

		// Start running the module migrations
		logger.Debug("running module migrations ...")
		return migrations, nil
	}
}

// update params to support fast block speed (1s block time)
func updateParamsToFastBlockSpeed(ctx sdk.Context, keepers *upgrades.UpgradeKeepers) {
	updateRollappParams(ctx, keepers.RollappKeeper)
	updateMintParams(ctx, keepers.MintKeeper)
	updateSlashingParams(ctx, keepers.SlashingKeeper)
	updateConsensusParams(ctx, keepers.ConsensusKeeper)
}

func updateSlashingParams(ctx sdk.Context, k *slashingkeeper.Keeper) {
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}

	// Update signed_blocks_window to maintain similar time window
	// 10000 blocks * 6s = 60000s = 60000 blocks * 1s
	params.SignedBlocksWindow = params.SignedBlocksWindow * BlockSpeedup

	err = k.SetParams(ctx, params)
	if err != nil {
		panic(err)
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
		incentivestypes.DefaultMinLockAge,        // Default to 1 day
		incentivestypes.DefaultMinLockDuration,   // Default to 0
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
		time.Minute,                // same as incentives.LockableDurations
	)
	keepers.LockupKeeper.SetParams(ctx, newParams)
}

func updateGAMMParams(ctx sdk.Context, k *gammkeeper.Keeper) {
	params := k.GetParams(ctx)

	// add all existing denoms to the allowed pool creation denoms
	for _, coin := range params.PoolCreationFee {
		params.AllowedPoolCreationDenoms = append(params.AllowedPoolCreationDenoms, coin.Denom)
	}

	// leave only "adym" in the pool creation fee
	if params.PoolCreationFee[0].Denom != "adym" {
		panic("first pool creation fee denom expected to be adym")
	}
	params.PoolCreationFee = sdk.NewCoins(params.PoolCreationFee[0])

	// set min swap amount to 0.01 DYM
	params.MinSwapAmount = math.NewIntWithDecimal(1, 16) // 0.01 DYM
	k.SetParams(ctx, params)
}

func updateTxfeesParams(ctx sdk.Context, k *txfeeskeeper.Keeper) {
	params := k.GetParams(ctx)

	params.FeeExemptMsgs = []string{
		sdk.MsgTypeURL(&gammtypes.MsgSwapExactAmountIn{}),
		sdk.MsgTypeURL(&gammtypes.MsgSwapExactAmountOut{}),
	}
	k.SetParams(ctx, params)
}

// addAuthorizedCircuitBreaker
func addAuthorizedCircuitBreaker(ctx sdk.Context, k *circuitkeeper.Keeper, ak *authkeeper.AccountKeeper) {
	permissions := circuittypes.Permissions{
		Level: circuittypes.Permissions_LEVEL_SUPER_ADMIN,
	}

	for _, grantee := range CircuitBreakPermissioned {
		grantee, err := ak.AddressCodec().StringToBytes(grantee)
		if err != nil {
			panic(err)
		}

		// Append the account in the msg to the store's set of authorized super admins
		if err = k.Permissions.Set(ctx, grantee, permissions); err != nil {
			panic(err)
		}
	}
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

func updateMintParams(ctx sdk.Context, k *mintkeeper.Keeper) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		panic(err)
	}

	// Update blocks_per_year to account for 1s block time (previously 6s)
	params.BlocksPerYear = params.BlocksPerYear * BlockSpeedup

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
			endorserPosition := sponsorshiptypes.NewEndorserPosition(shares, endorsement.Accumulator, sdk.NewCoins())

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
		delayedackParams.DeletePacketsEpochLimit,
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
	var dymnsParams dymns.Params
	dymnsSubspace, ok := keepers.ParamsKeeper.GetSubspace(dymnstypes.ModuleName)
	if !ok {
		dymnsSubspace = keepers.ParamsKeeper.Subspace(dymns.ModuleName)
		dymnsSubspace = dymnsSubspace.WithKeyTable(dymns.ParamKeyTable())
	}
	dymnsSubspace.GetParamSetIfExists(ctx, &dymnsParams)
	err := keepers.DymNSKeeper.SetParams(ctx, dymnstypes.NewParams(
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
	if err != nil {
		panic(err)
	}

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
		newTeeConfig,
	))

	// Streamer module
	streamerSubspace := keepers.ParamsKeeper.Subspace(streamer.ModuleName)
	streamerSubspace = streamerSubspace.WithKeyTable(streamer.ParamKeyTable())
	var streamerParams streamer.Params
	streamerSubspace.GetParamSetIfExists(ctx, &streamerParams)
	keepers.StreamerKeeper.SetParams(ctx, streamermoduletypes.NewParams(
		streamerParams.MaxIterationsPerBlock,
	))

	// Incentives module params migration
	migrateAndUpdateIncentivesParams(ctx, keepers)

	// lockup module params migrations
	migrateAndUpdateLockupParams(ctx, keepers)
}

// TODO: use proper values
var newTeeConfig = rollappmoduletypes.TEEConfig{
	Enabled:         true,
	Verify:          false,
	PolicyValues:    "",
	PolicyQuery:     "",
	PolicyStructure: "",
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

// migrateLockTimestamps sets UpdatedAt on all locks if not set
func migrateLockTimestamps(ctx sdk.Context, lockupKeeper *lockupkeeper.Keeper, incentivesKeeper *incentiveskeeper.Keeper) error {
	locks, err := lockupKeeper.GetPeriodLocks(ctx)
	if err != nil {
		return fmt.Errorf("get period locks: %w", err)
	}

	// for each lock, set the updated_at to the min lock age eligible for distribution
	for _, lock := range locks {
		lock.UpdatedAt = ctx.BlockTime().Add(-incentivestypes.DefaultMinLockAge)
		err := lockupKeeper.SetLock(ctx, lock)
		if err != nil {
			return fmt.Errorf("set lock %d: %w", lock.ID, err)
		}
	}
	return nil
}

// migrateGaugeLockAges sets the min lock age for all perpetual asset gauges
func migrateGaugeLockAges(ctx sdk.Context, incentivesKeeper *incentiveskeeper.Keeper) error {
	minLockAge := incentivestypes.DefaultMinLockAge
	gauges := incentivesKeeper.GetGauges(ctx)
	for _, gauge := range gauges {
		if gauge.IsPerpetual && gauge.GetAsset() != nil {
			asset := gauge.GetAsset()
			asset.LockAge = minLockAge
			if err := incentivesKeeper.SetGauge(ctx, &gauge); err != nil {
				return fmt.Errorf("set gauge %d: %w", gauge.Id, err)
			}
		}
	}
	return nil
}

func updateConsensusParams(ctx sdk.Context, csk *consensusparamkeeper.Keeper) {
	// Get current consensus params
	consensusParamsRes, err := csk.Params(ctx, nil)
	if err != nil {
		panic(err)
	}
	consensusParams := consensusParamsRes.Params

	// Update MaxAgeNumBlocks to maintain similar time window
	// If it was set to 100000 blocks at 6s, it should be 600000 blocks at 1s
	consensusParams.Evidence.MaxAgeNumBlocks = consensusParams.Evidence.MaxAgeNumBlocks * BlockSpeedup

	// Create a new consensus params update message
	msg := &consensusparamtypes.MsgUpdateParams{
		Authority: csk.GetAuthority(),
		Block:     consensusParams.Block,
		Evidence:  consensusParams.Evidence,
		Validator: consensusParams.Validator,
	}

	// Update the consensus params
	_, err = csk.UpdateParams(ctx, msg)
	if err != nil {
		panic(err)
	}
}

// setupRateLimitingParams sets up the rate limiting parameters for Noble USDC and Kava USDT
func setupRateLimitingParams(ctx sdk.Context, k *ratelimitkeeper.Keeper) error {
	for _, path := range IBCChannels {
		// 1-Day Limit (15% send, no receive limit, 24h)
		err := k.AddRateLimit(ctx, &ratelimittypes.MsgAddRateLimit{
			Authority:      "", // is not necessary here
			Denom:          path.Denom,
			ChannelId:      path.ChannelId,
			MaxPercentSend: math.NewInt(15),  // 15%
			MaxPercentRecv: math.NewInt(100), // 100% is effectively no limit
			DurationHours:  24,
		})
		if err != nil {
			return fmt.Errorf("add rate limit: denom: %s, channelID: %s, error: %w", path.Denom, path.ChannelId, err)
		}
	}
	return nil
}
