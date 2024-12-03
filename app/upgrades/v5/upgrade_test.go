package v5_test

import (
	"fmt"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	cometbftproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/dymensionxyz/dymension/v3/app"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/sdk-utils/utils/urand"
	"github.com/stretchr/testify/suite"
)

// UpgradeTestSuite defines the structure for the upgrade test suite
type UpgradeTestSuite struct {
	suite.Suite
	Ctx sdk.Context
	App *app.App
}

// SetupTest initializes the necessary items for each test
func (s *UpgradeTestSuite) SetupTestCustom(t *testing.T) {
	s.App = apptesting.Setup(t)
	s.Ctx = s.App.BaseApp.NewContext(false, cometbftproto.Header{Height: 1, ChainID: "dymension_100-1", Time: time.Now().UTC()})
}

// TestUpgradeTestSuite runs the suite of tests for the upgrade handler
func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

const (
	dummyUpgradeHeight int64 = 10
)

func (s *UpgradeTestSuite) TestUpgrade() {
	testCases := []struct {
		msg         string
		numRollapps int
		preUpgrade  func(int) error
		upgrade     func()
		postUpgrade func(int) error
		expPass     bool
	}{
		{
			msg:         "Test that upgrade does not panic and sets correct parameters and migrates rollapp module",
			numRollapps: 5,
			preUpgrade: func(numRollapps int) error {
				// Create and store rollapps
				s.seedAndStoreRollapps(numRollapps)
				return nil
			},
			upgrade: func() {
				// Run upgrade
				s.Ctx = s.Ctx.WithBlockHeight(dummyUpgradeHeight - 1)
				plan := upgradetypes.Plan{Name: "v5", Height: dummyUpgradeHeight}

				err := s.App.UpgradeKeeper.ScheduleUpgrade(s.Ctx, plan)
				s.Require().NoError(err)
				_, exists := s.App.UpgradeKeeper.GetUpgradePlan(s.Ctx)
				s.Require().True(exists)

				s.Ctx = s.Ctx.WithBlockHeight(dummyUpgradeHeight)
				// simulate the upgrade process not panic.
				s.Require().NotPanics(func() {
					defer func() {
						if r := recover(); r != nil {
							s.Fail("Upgrade panicked", r)
						}
					}()
					// simulate the upgrade process.
					s.App.BeginBlocker(s.Ctx, abci.RequestBeginBlock{})
				})
			},
			postUpgrade: func(numRollapps int) error {
				// Post-upgrade validation to ensure values are correctly set

				// Check Rollapp parameters
				if err := s.validateRollappParamsMigration(); err != nil {
					return err
				}

				// Check Rollapps
				if err := s.validateRollappsMigration(numRollapps); err != nil {
					return err
				}

				return nil
			},
			expPass: true,
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			s.SetupTestCustom(s.T()) // Reset for each case

			err := tc.preUpgrade(tc.numRollapps)
			s.Require().NoError(err)
			tc.upgrade()
			err = tc.postUpgrade(tc.numRollapps)
			if tc.expPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}
		})
	}
}

func (s *UpgradeTestSuite) validateRollappParamsMigration() error {
	rollappParams := s.App.RollappKeeper.GetParams(s.Ctx)
	cond := rollappParams.MinSequencerBondGlobal.IsEqual(rollapptypes.DefaultMinSequencerBondGlobalCoin)

	if !cond {
		return fmt.Errorf("rollapp parameters not set correctly")
	}
	return nil
}

func (s *UpgradeTestSuite) validateRollappsMigration(numRoll int) error {
	rollapps := s.App.RollappKeeper.GetAllRollapps(s.Ctx)
	s.Require().Len(rollapps, numRoll)

	for _, ra := range rollapps {
		s.Require().NoError(rollapptypes.ValidateBasicMinSeqBondCoins(ra.MinSequencerBond))
		s.Require().Equal(ra.GenesisState.GetDeprecatedBridgeOpen(), ra.IsTransferEnabled())
		s.Require().Equal(sdk.NewCoins(rollapptypes.DefaultMinSequencerBondGlobalCoin), ra.MinSequencerBond)
	}

	return nil
}

func (s *UpgradeTestSuite) seedAndStoreRollapps(numRollapps int) {
	for _, rollapp := range s.seedRollapps(numRollapps) {
		s.App.RollappKeeper.SetRollapp(s.Ctx, rollapp)
	}
}

func (s *UpgradeTestSuite) seedRollapps(numRollapps int) []rollapptypes.Rollapp {
	rollapps := make([]rollapptypes.Rollapp, numRollapps)
	for i := range rollapps {
		rollappID := urand.RollappID()
		ra := rollapptypes.Rollapp{
			RollappId: rollappID,
			Owner:     fmt.Sprintf("owner_%d", i),
			GenesisState: rollapptypes.RollappGenesisState{
				DeprecatedBridgeOpen: i%2 == 0,
			},
			ChannelId: fmt.Sprintf("channel-%d", i),
		}
		rollapps[i] = ra
	}
	return rollapps
}
