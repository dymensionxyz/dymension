package v4_test

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	cometbftproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/v3/app"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	v4 "github.com/dymensionxyz/dymension/v3/app/upgrades/v4"
	"github.com/dymensionxyz/dymension/v3/app/upgrades/v4/types"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// UpgradeTestSuite defines the structure for the upgrade test suite
type UpgradeTestSuite struct {
	suite.Suite
	Ctx sdk.Context
	App *app.App
}

// SetupTest initializes the necessary items for each test
func (s *UpgradeTestSuite) SetupTest(t *testing.T) {
	s.App = apptesting.Setup(t, false)
	s.Ctx = s.App.BaseApp.NewContext(false, cometbftproto.Header{Height: 1, ChainID: "dymension_100-1", Time: time.Now().UTC()})
}

// TestUpgradeTestSuite runs the suite of tests for the upgrade handler
func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

const (
	dummyUpgradeHeight                      int64 = 5
	expectDelayedackDeletePacketsEpochLimit int32 = 1000_000
	expectDelayedackEpochIdentifier               = "hour"

	expectDisputePeriodInBlocks = 3
	expectRegistrationFee       = "10000000000000000000adym"
)

var expectDelayedackBridgingFee = sdk.NewDecWithPrec(1, 3)

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
			msg: "Test that upgrade does not panic and sets correct parameters and migrates rollapp module",
			preUpgrade: func() error {
				// Create and store old rollapps
				numRollapps := 5
				s.createAndStoreOldRollapps(numRollapps)
				return nil
			},
			upgrade: func() {
				// Run upgrade
				s.Ctx = s.Ctx.WithBlockHeight(dummyUpgradeHeight - 1)
				plan := upgradetypes.Plan{Name: "v4", Height: dummyUpgradeHeight}
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
			postUpgrade: func() (err error) {
				// Post-update validation to ensure values are correctly set

				// Check Delayedack parameters
				if err = s.validateDelayedAckParamsMigration(); err != nil {
					return
				}

				// Check Rollapp parameters
				if err = s.validateRollappParamsMigration(); err != nil {
					return
				}

				// Check Rollapps
				if err = s.validateRollappsMigration(); err != nil {
					return
				}

				return
			},
			expPass: true,
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			s.SetupTest(s.T()) // Reset for each case

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

func (s *UpgradeTestSuite) validateDelayedAckParamsMigration() error {
	delayedackParams := s.App.DelayedAckKeeper.GetParams(s.Ctx)
	cond := delayedackParams.DeletePacketsEpochLimit == expectDelayedackDeletePacketsEpochLimit &&
		delayedackParams.EpochIdentifier == expectDelayedackEpochIdentifier &&
		delayedackParams.BridgingFee.Equal(expectDelayedackBridgingFee)

	if !cond {
		return fmt.Errorf("delayedack parameters not set correctly")
	}
	return nil
}

func (s *UpgradeTestSuite) validateRollappParamsMigration() error {
	rollappParams := s.App.RollappKeeper.GetParams(s.Ctx)
	cond := rollappParams.DisputePeriodInBlocks == expectDisputePeriodInBlocks &&
		rollappParams.RegistrationFee.String() == expectRegistrationFee

	if !cond {
		return fmt.Errorf("rollapp parameters not set correctly")
	}
	return nil
}

func (s *UpgradeTestSuite) validateRollappsMigration() error {
	rollapps := s.App.RollappKeeper.GetAllRollapps(s.Ctx)
	s.Require().Len(rollapps, len(newRollapps))

	for _, rollapp := range rollapps {
		rollappID, _ := rollapptypes.NewChainID(rollapp.RollappId)
		// check that the rollapp can be retrieved by EIP155 key
		if _, ok := s.App.RollappKeeper.GetRollappByEIP155(s.Ctx, rollappID.GetEIP155ID()); !ok {
			return fmt.Errorf("rollapp by EIP155 not found")
		}
	}

	if len(rollapps) != len(newRollapps) {
		return fmt.Errorf("rollapps length not equal")
	}

	if !reflect.DeepEqual(rollapps, newRollapps) {
		return fmt.Errorf("rollapps do not match")
	}
	return nil
}

var (
	oldRollapps []types.Rollapp
	newRollapps []rollapptypes.Rollapp
)

func (s *UpgradeTestSuite) createAndStoreOldRollapps(numRollapps int) (ids []string) {
	// create 5 rollapps with the old proto version
	storeKey := s.App.GetKey(rollapptypes.StoreKey)
	store := prefix.NewStore(s.Ctx.KVStore(storeKey), rollapptypes.KeyPrefix(rollapptypes.RollappKeyPrefix))
	eip155Store := prefix.NewStore(s.Ctx.KVStore(storeKey), rollapptypes.KeyPrefix(rollapptypes.RollappByEIP155KeyPrefix))
	createMockOldAndNewRollapps(numRollapps)

	for _, rollapp := range oldRollapps {
		bz := s.App.AppCodec().MustMarshalJSON(&rollapp)
		store.Set(rollapptypes.RollappKey(rollapp.RollappId), bz)

		rollappID, _ := rollapptypes.NewChainID(rollapp.RollappId)
		if !rollappID.IsEIP155() {
			return
		}

		eip155Store.Set(rollapptypes.RollappByEIP155Key(
			rollappID.GetEIP155ID(),
		), []byte(rollapp.RollappId))
		ids = append(ids, rollapp.RollappId)
	}
	return
}

func createMockOldAndNewRollapps(nRollapps int) {
	oldRollapps = make([]types.Rollapp, nRollapps)
	newRollapps = make([]rollapptypes.Rollapp, nRollapps)

	for i := range nRollapps {
		oldRollapp := types.Rollapp{
			RollappId:     fmt.Sprintf("roll%spp_123%d-1", string(rune(i+97)), i+1),
			Creator:       sample.AccAddress(),
			Version:       0,
			MaxSequencers: 10,
			PermissionedAddresses: []string{
				sample.AccAddress(),
				sample.AccAddress(),
			},
			GenesisState:     types.RollappGenesisState{},
			ChannelId:        fmt.Sprintf("channel-%d", i),
			RegisteredDenoms: []string{"denom1", "denom2"},
		}
		oldRollapps[i] = oldRollapp
		newRollapps[i] = v4.ConvertOldRollappToNew(oldRollapp)
	}
}
