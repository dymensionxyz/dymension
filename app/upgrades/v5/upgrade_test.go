package v5_test

import (
	"fmt"
	"slices"
	"testing"
	"time"

	"cosmossdk.io/core/header"
	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	abci "github.com/cometbft/cometbft/abci/types"
	cometbftproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/v3/app"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	v5 "github.com/dymensionxyz/dymension/v3/app/upgrades/v5"
	"github.com/dymensionxyz/dymension/v3/x/common/types"
	irotypes "github.com/dymensionxyz/dymension/v3/x/iro/types"
	lockuptypes "github.com/dymensionxyz/dymension/v3/x/lockup/types"
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
	params := lockuptypes.Params{
		ForceUnlockAllowedAddresses: expectLockupForceUnlockAllowedAddresses,
	}
	s.App.LockupKeeper.SetParams(s.Ctx, params)
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
