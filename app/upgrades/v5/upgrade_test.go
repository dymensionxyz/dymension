package v5_test

import (
	"fmt"
	"slices"
	"strconv"
	"testing"
	"time"

	"cosmossdk.io/core/header"
	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	abci "github.com/cometbft/cometbft/abci/types"
	cometbftproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/v3/app"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	v5 "github.com/dymensionxyz/dymension/v3/app/upgrades/v5"
	lockupmigration "github.com/dymensionxyz/dymension/v3/app/upgrades/v5/types/lockup"
	"github.com/dymensionxyz/dymension/v3/x/common/types"
	irotypes "github.com/dymensionxyz/dymension/v3/x/iro/types"
	lockuptypes "github.com/dymensionxyz/dymension/v3/x/lockup/types"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencerkeeper "github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// UpgradeTestSuite defines the structure for the upgrade test suite
type UpgradeTestSuite struct {
	suite.Suite
	Ctx sdk.Context
	App *app.App
}

// SetupTestCustom initializes the necessary items for each test
func (s *UpgradeTestSuite) SetupTestCustom(t *testing.T) {
	s.App = apptesting.Setup(t)
	s.Ctx = s.App.BaseApp.NewContext(false).WithBlockHeader(cometbftproto.Header{Height: 1, ChainID: "dymension_100-1", Time: time.Now().UTC()}).WithChainID("dymension_100-1")

	defParams := *apptesting.DefaultConsensusParams
	s.Ctx = s.Ctx.WithConsensusParams(defParams)
}

// TestUpgradeTestSuite runs the suite of tests for the upgrade handler
func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

const (
	dummyUpgradeHeight int64 = 5
)

var (
	expectLockupCreationFee                 = types.DYM.QuoRaw(20)
	expectLockupForceUnlockAllowedAddresses = []string{
		"dym19pas0pqwje540u5ptwnffjxeamdxc9tajmdrfa",
		"dym15saxgqw6kvhv6k5sg6r45kmdf4sf88kfw2adcw",
		"dym17g9cn4ss0h0dz5qhg2cg4zfnee6z3ftg3q6v58",
	}

	expectedEvidenceMaxAgeNumBlocks = apptesting.DefaultConsensusParams.Evidence.MaxAgeNumBlocks * v5.BlockSpeedup
)

// TestUpgrade is a method of UpgradeTestSuite to test the upgrade process.
func (s *UpgradeTestSuite) TestUpgrade() {
	testCases := []struct {
		msg         string
		preUpgrade  func() error
		upgrade     func()
		postUpgrade func() error
		expPass     bool
	}{
		{
			msg: "Test that upgrade does not panic and sets correct parameters",
			preUpgrade: func() error {
				s.setLockupParams()
				s.setIROParams()
				s.setGAMMParams()
				s.populateSequencers(s.Ctx, s.App.SequencerKeeper)
				s.populateLivenessEvents(s.Ctx, s.App.RollappKeeper)
				s.populateIBCChannels()
				return nil
			},
			upgrade: func() {
				// Run upgrade
				s.Ctx = s.Ctx.WithBlockHeight(dummyUpgradeHeight - 1)
				plan := upgradetypes.Plan{Name: v5.UpgradeName, Height: dummyUpgradeHeight}

				err := s.App.UpgradeKeeper.ScheduleUpgrade(s.Ctx, plan)
				s.Require().NoError(err)
				_, err = s.App.UpgradeKeeper.GetUpgradePlan(s.Ctx)
				s.Require().NoError(err)

				s.Ctx = s.Ctx.WithHeaderInfo(header.Info{Height: dummyUpgradeHeight, Time: s.Ctx.BlockTime().Add(time.Second)}).WithBlockHeight(dummyUpgradeHeight)
				// simulate the upgrade process not panic.
				s.Require().NotPanics(func() {
					defer func() {
						if r := recover(); r != nil {
							s.Fail("Upgrade panicked", r)
						}
					}()
					// simulate the upgrade process.
					_, err = s.App.PreBlocker(s.Ctx, &abci.RequestFinalizeBlock{})
					s.Require().NoError(err)
				})
			},
			postUpgrade: func() (err error) {
				// Post-update validation to ensure values are correctly set

				// Check Lockup parameters
				if err = s.validateLockupParamsMigration(); err != nil {
					return
				}

				// validate IRO params
				if err = s.validateIROParamsMigration(); err != nil {
					return
				}

				// validate gamm params
				if err = s.validateGammParamsMigration(); err != nil {
					return
				}

				if err = s.validateLivenessEventsMigration(s.Ctx, s.App.RollappKeeper); err != nil {
					return
				}

				s.validateSequencersMigration(s.Ctx, s.App.SequencerKeeper)

				s.validateIBCRateLimits()

				// validate consensus params
				s.validateConsensusParamsMigration()

				return
			},
			expPass: true,
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			s.SetupTestCustom(s.T()) // Reset for each case

			err := tc.preUpgrade()
			s.Require().NoError(err)
			tc.upgrade()
			err = tc.postUpgrade()
			if tc.expPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}
		})
	}
}

func (s *UpgradeTestSuite) setLockupParams() {
	params := lockupmigration.Params{
		ForceUnlockAllowedAddresses: expectLockupForceUnlockAllowedAddresses,
	}
	lockupSubspace := s.App.ParamsKeeper.Subspace(lockuptypes.ModuleName)
	lockupSubspace = lockupSubspace.WithKeyTable(lockupmigration.ParamKeyTable())
	lockupSubspace.SetParamSet(s.Ctx, &params)
}

func (s *UpgradeTestSuite) validateLockupParamsMigration() error {
	lockupParams := s.App.LockupKeeper.GetParams(s.Ctx)
	cond := slices.Equal(lockupParams.ForceUnlockAllowedAddresses, expectLockupForceUnlockAllowedAddresses) &&
		lockupParams.LockCreationFee.Equal(expectLockupCreationFee)

	if !cond {
		return fmt.Errorf("lockup parameters not set correctly")
	}
	return nil
}

func (s *UpgradeTestSuite) setIROParams() {
	oldParams := irotypes.DefaultParams()

	oldParams.MinLiquidityPart = math.LegacyDec{}
	oldParams.MinVestingDuration = 0
	oldParams.MinVestingStartTimeAfterSettlement = 0

	s.App.IROKeeper.SetParams(s.Ctx, oldParams)
}

func (s *UpgradeTestSuite) validateIROParamsMigration() error {
	params := s.App.IROKeeper.GetParams(s.Ctx)
	expected := irotypes.DefaultParams()

	if !params.MinLiquidityPart.Equal(expected.MinLiquidityPart) {
		return fmt.Errorf("min liquidity part not set correctly")
	}

	if params.MinVestingDuration != expected.MinVestingDuration || params.MinVestingStartTimeAfterSettlement != expected.MinVestingStartTimeAfterSettlement {
		return fmt.Errorf("min vesting duration or start time after settlement not set correctly")
	}

	return nil
}

func (s *UpgradeTestSuite) setGAMMParams() {
	params := s.App.GAMMKeeper.GetParams(s.Ctx)
	params.PoolCreationFee = sdk.NewCoins(sdk.NewCoin("adym", math.NewInt(100000000000000000)),
		sdk.NewCoin("ibc/B3504E092456BA618CC28AC671A71FB08C6CA0FD0BE7C8A5B5A3E2DD933CC9E4", math.NewInt(1)))
	s.App.GAMMKeeper.SetParams(s.Ctx, params)
}

// validate gamm params
func (s *UpgradeTestSuite) validateGammParamsMigration() error {
	params := s.App.GAMMKeeper.GetParams(s.Ctx)

	if len(params.PoolCreationFee) != 1 {
		return fmt.Errorf("pool creation fee not set correctly")
	}

	if params.PoolCreationFee[0].Denom != "adym" {
		return fmt.Errorf("pool creation fee not set correctly")
	}

	if !params.MinSwapAmount.Equal(math.NewInt(100000000000000000)) {
		return fmt.Errorf("min swap amount not set correctly")
	}

	return nil
}

var livenessEventsBlocks = []int64{0, 100, 200, 300}

func (s *UpgradeTestSuite) populateLivenessEvents(ctx sdk.Context, k *rollappkeeper.Keeper) {
	for i, h := range livenessEventsBlocks {
		k.PutLivenessEvent(ctx, rollapptypes.LivenessEvent{
			RollappId: strconv.Itoa(i),
			HubHeight: dummyUpgradeHeight + h,
		})
	}
}

func (s *UpgradeTestSuite) populateIBCChannels() {
	for _, path := range v5.IBCChannels {
		s.App.IBCKeeper.ChannelKeeper.SetChannel(s.Ctx, transfertypes.PortID, path.ChannelId, channeltypes.Channel{})
		err := s.App.BankKeeper.MintCoins(s.Ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin(path.Denom, math.NewInt(100))))
		s.Require().NoError(err)
	}
}

func (s *UpgradeTestSuite) validateIBCRateLimits() {
	for _, path := range v5.IBCChannels {
		_, found := s.App.RateLimitingKeeper.GetRateLimit(s.Ctx, path.Denom, path.ChannelId)
		s.Require().True(found)
	}
}

func (s *UpgradeTestSuite) validateLivenessEventsMigration(ctx sdk.Context, k *rollappkeeper.Keeper) error {
	evts := k.GetLivenessEvents(ctx, nil)
	s.Require().Equal(len(evts), len(livenessEventsBlocks))
	for i, e := range evts {
		s.Require().Equal(ctx.BlockHeight()+livenessEventsBlocks[i]*v5.BlockSpeedup, e.HubHeight)
	}

	return nil
}

func (s *UpgradeTestSuite) populateSequencers(ctx sdk.Context, k *sequencerkeeper.Keeper) {
	k.SetSequencer(ctx, sequencertypes.Sequencer{
		Address:  "dym19pas0pqwje540u5ptwnffjxeamdxc9tajmdrfa",
		Status:   sequencertypes.Bonded,
		Dishonor: v5.NewPenaltyKickThreshold + 1,
	})
}

func (s *UpgradeTestSuite) validateSequencersMigration(ctx sdk.Context, k *sequencerkeeper.Keeper) {
	sequencers := k.AllSequencers(ctx)
	s.Require().Equal(len(sequencers), 1)
	s.Require().Equal(v5.NewPenaltyKickThreshold, sequencers[0].GetPenalty())
}

func (s *UpgradeTestSuite) validateConsensusParamsMigration() {
	consensusParams, err := s.App.ConsensusParamsKeeper.Params(s.Ctx, nil)
	s.Require().NoError(err)
	s.Require().Equal(expectedEvidenceMaxAgeNumBlocks, consensusParams.Params.Evidence.MaxAgeNumBlocks)
}
