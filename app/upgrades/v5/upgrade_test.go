package v5_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/dymensionxyz/dymension/v3/app"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	v5 "github.com/dymensionxyz/dymension/v3/app/upgrades/v5"
	"github.com/dymensionxyz/dymension/v3/app/upgrades/v5/types"
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
	s.Ctx = s.App.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: "dymension_100-1", Time: time.Now().UTC()})
}

// TestUpgradeTestSuite runs the suite of tests for the upgrade handler
func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

const (
	dummyUpgradeHeight          = 5
	expectDisputePeriodInBlocks = 3
	expectRegistrationFee       = "1000000000000000000adym"
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
				// create 5 rollapps with the old proto version
				storeKey := s.App.GetKey(rollapptypes.StoreKey)
				store := prefix.NewStore(s.Ctx.KVStore(storeKey), rollapptypes.KeyPrefix(rollapptypes.RollappKeyPrefix))
				eip155Store := prefix.NewStore(s.Ctx.KVStore(storeKey), rollapptypes.KeyPrefix(rollapptypes.RollappByEIP155KeyPrefix))
				createMockOldAndNewRollapps(1)

				for _, rollapp := range oldRollapps {
					bz := s.App.AppCodec().MustMarshalJSON(&rollapp)
					store.Set(rollapptypes.RollappKey(rollapp.RollappId), bz)

					rollappID, _ := rollapptypes.NewChainID(rollapp.RollappId)
					if !rollappID.IsEIP155() {
						return nil
					}

					eip155Store.Set(rollapptypes.RollappByEIP155Key(
						rollappID.GetEIP155ID(),
					), bz)
				}
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
					// simulate the upgrade process.
					s.App.BeginBlocker(s.Ctx, abci.RequestBeginBlock{})
				})
			},
			postUpgrade: func() error {
				// Post-update validation to ensure parameters are correctly set
				rollappParams := s.App.RollappKeeper.GetParams(s.Ctx)
				if rollappParams.DisputePeriodInBlocks != expectDisputePeriodInBlocks ||
					rollappParams.RegistrationFee.String() != expectRegistrationFee {
					return fmt.Errorf("rollapp parameters not set correctly")
				}

				// Check that the rollapps have been migrated correctly

				rollapps := s.App.RollappKeeper.GetAllRollapps(s.Ctx)
				s.Require().Len(rollapps, len(newRollapps))

				for _, rollapp := range rollapps {
					rollappID, _ := rollapptypes.NewChainID(rollapp.RollappId)
					// check that the rollapp can be retrieved by EIP155 key
					_, ok := s.App.RollappKeeper.GetRollappByEIP155(s.Ctx, rollappID.GetEIP155ID())
					if !ok {
						return fmt.Errorf("rollapp by EIP155 key not migrated")
					}
				}
				s.Require().ElementsMatch(rollapps, newRollapps)

				return nil
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

var (
	oldRollapps []types.Rollapp
	newRollapps []rollapptypes.Rollapp
)

func createMockOldAndNewRollapps(nRollapps int) {
	oldRollapps = make([]types.Rollapp, nRollapps)
	newRollapps = make([]rollapptypes.Rollapp, nRollapps)

	for i := 0; i < nRollapps; i++ {
		oldRollapp := types.Rollapp{
			RollappId:     fmt.Sprintf("rollapp_123%d-1", i+1),
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
		newRollapps[i] = v5.ConvertOldRollappToNew(oldRollapp)
	}
}
