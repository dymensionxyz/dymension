package v4_test

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	cometbftproto "github.com/cometbft/cometbft/proto/tendermint/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/v3/app"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	v4 "github.com/dymensionxyz/dymension/v3/app/upgrades/v4"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
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

				// Create and store sequencers
				s.seedAndStoreSequencers(numRollapps)
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
			postUpgrade: func(numRollapps int) (err error) {
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
				if err = s.validateRollappsMigration(numRollapps); err != nil {
					return
				}

				// Check Sequencers
				if err = s.validateSequencersMigration(numRollapps); err != nil {
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

func (s *UpgradeTestSuite) validateRollappsMigration(numRoll int) error {
	expectRollapps := make([]rollapptypes.Rollapp, numRoll)
	for i, rollapp := range s.seedRollapps(numRoll) {
		expectRollapps[i] = v4.ConvertOldRollappToNew(rollapp)
	}
	rollapps := s.App.RollappKeeper.GetAllRollapps(s.Ctx)
	s.Require().Len(rollapps, len(expectRollapps))

	for _, rollapp := range rollapps {
		rollappID, _ := rollapptypes.NewChainID(rollapp.RollappId)
		// check that the rollapp can be retrieved by EIP155 key
		if _, ok := s.App.RollappKeeper.GetRollappByEIP155(s.Ctx, rollappID.GetEIP155ID()); !ok {
			return fmt.Errorf("rollapp by EIP155 not found")
		}
	}

	if !reflect.DeepEqual(rollapps, expectRollapps) {
		return fmt.Errorf("rollapps do not match")
	}
	return nil
}

func (s *UpgradeTestSuite) validateSequencersMigration(numSeq int) error {
	expectSequencers := make([]sequencertypes.Sequencer, numSeq)
	for i, sequencer := range s.seedSequencers(numSeq) {
		expectSequencers[i] = v4.ConvertOldSequencerToNew(sequencer)
	}
	sequencers := s.App.SequencerKeeper.GetAllSequencers(s.Ctx)
	s.Require().Len(sequencers, len(expectSequencers))

	sort.Slice(sequencers, func(i, j int) bool {
		return sequencers[i].Address < sequencers[j].Address
	})

	sort.Slice(expectSequencers, func(i, j int) bool {
		return expectSequencers[i].Address < expectSequencers[j].Address
	})

	for i, sequencer := range sequencers {
		// check that the sequencer can be retrieved by address
		_, ok := s.App.SequencerKeeper.GetSequencer(s.Ctx, sequencer.Address)
		if !ok {
			return fmt.Errorf("sequencer by address not migrated")
		}

		seq := s.App.AppCodec().MustMarshalJSON(&sequencer)
		nSeq := s.App.AppCodec().MustMarshalJSON(&expectSequencers[i])

		s.Require().JSONEq(string(seq), string(nSeq))
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
	for i := range numRollapps {
		rollappID := rollappIDFromIdx(i)
		rollapp := rollapptypes.Rollapp{
			RollappId:        rollappID,
			Owner:            sample.AccAddressFromSecret(rollappID),
			GenesisState:     rollapptypes.RollappGenesisState{},
			ChannelId:        fmt.Sprintf("channel-%d", i),
			RegisteredDenoms: []string{"denom1", "denom2"},
		}
		rollapps[i] = rollapp
	}
	return rollapps
}

func (s *UpgradeTestSuite) seedAndStoreSequencers(numRollapps int) {
	for _, sequencer := range s.seedSequencers(numRollapps) {
		s.App.SequencerKeeper.SetSequencer(s.Ctx, sequencer)
	}
}

func (s *UpgradeTestSuite) seedSequencers(numSeq int) []sequencertypes.Sequencer {
	sequencers := make([]sequencertypes.Sequencer, numSeq)
	for i := 0; i < numSeq; i++ {
		rollappID := rollappIDFromIdx(i)
		pk := ed25519.GenPrivKeyFromSecret([]byte(rollappID)).PubKey()
		pkAny, _ := codectypes.NewAnyWithValue(pk)
		sequencer := sequencertypes.Sequencer{
			Address:      sdk.AccAddress(pk.Address()).String(),
			DymintPubKey: pkAny,
			RollappId:    rollappID,
			Metadata: sequencertypes.SequencerMetadata{
				Moniker: fmt.Sprintf("sequencer-%d", i),
				Details: fmt.Sprintf("Additional details about the sequencer-%d", i),
			},
			Status:   sequencertypes.Bonded,
			Proposer: true,
			Tokens:   sdk.NewCoins(sdk.NewInt64Coin("dym", 100)),
		}
		sequencers[i] = sequencer
	}
	return sequencers
}

func rollappIDFromIdx(idx int) string {
	return fmt.Sprintf("roll%spp_123%d-1", string(rune(idx+'a')), idx+1)
}
