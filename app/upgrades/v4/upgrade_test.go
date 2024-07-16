package v4_test

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/dymensionxyz/dymension/v3/app"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	v4 "github.com/dymensionxyz/dymension/v3/app/upgrades/v4"
	"github.com/dymensionxyz/dymension/v3/app/upgrades/v4/types"
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
	s.Ctx = s.App.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: "dymension_100-1", Time: time.Now().UTC()})
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
	expectRegistrationFee       = "1000000000000000000adym"
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
				rollappIDs := s.createAndStoreOldRollapps(numRollapps)

				// Create and store old sequencers
				s.createAndStoreOldSequencers(rollappIDs)
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

				// Check Sequencers
				if err = s.validateSequencersMigration(); err != nil {
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
			return fmt.Errorf("rollapp not migrated correctly")
		}
	}

	if len(rollapps) != len(newRollapps) {
		return fmt.Errorf("rollapps not migrated correctly")
	}

	if !reflect.DeepEqual(rollapps, newRollapps) {
		return fmt.Errorf("rollapps not migrated correctly")
	}
	return nil
}

func (s *UpgradeTestSuite) validateSequencersMigration() error {
	sequencers := s.App.SequencerKeeper.GetAllSequencers(s.Ctx)
	s.Require().Len(sequencers, len(newSequencers))

	sort.Slice(sequencers, func(i, j int) bool {
		return sequencers[i].Address < sequencers[j].Address
	})

	for i, sequencer := range sequencers {
		// check that the sequencer can be retrieved by address
		_, ok := s.App.SequencerKeeper.GetSequencer(s.Ctx, sequencer.Address)
		if !ok {
			return fmt.Errorf("sequencer by address not migrated")
		}

		// check that the sequencer can be retrieved by rollapp and status
		sequencer, ok = s.App.SequencerKeeper.GetSequencerByRollappByStatus(s.Ctx, sequencer.RollappId, sequencer.Address, sequencer.Status)
		if !ok {
			return fmt.Errorf("sequencer by rollapp and status not migrated")
		}

		seq := s.App.AppCodec().MustMarshalJSON(&sequencer)
		nSeq := s.App.AppCodec().MustMarshalJSON(&newSequencers[i])

		s.Require().JSONEq(string(seq), string(nSeq))
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
		newRollapps[i] = v4.ConvertOldRollappToNew(oldRollapp)
	}
}

var (
	oldSequencers []types.Sequencer
	newSequencers []sequencertypes.Sequencer
)

func (s *UpgradeTestSuite) createAndStoreOldSequencers(rollappIDs []string) {
	storeKey := s.App.GetKey(sequencertypes.StoreKey)
	store := s.Ctx.KVStore(storeKey)
	createMockOldAndNewSequencers(rollappIDs)

	for _, sequencer := range oldSequencers {
		bz := s.App.AppCodec().MustMarshalJSON(&sequencer)
		store.Set(sequencertypes.SequencerKey(
			sequencer.SequencerAddress,
		), bz)

		seqByRollappKey := sequencertypes.SequencerByRollappByStatusKey(
			sequencer.RollappId,
			sequencer.SequencerAddress,
			sequencertypes.OperatingStatus(sequencer.Status),
		)
		store.Set(seqByRollappKey, bz)
	}
}

func createMockOldAndNewSequencers(rollappIDs []string) {
	numSeq := len(rollappIDs)
	oldSequencers = make([]types.Sequencer, numSeq)
	newSequencers = make([]sequencertypes.Sequencer, numSeq)

	for i := 0; i < numSeq; i++ {
		pk := ed25519.GenPrivKey().PubKey()
		pkAny, _ := codectypes.NewAnyWithValue(pk)
		oldSequencer := types.Sequencer{
			SequencerAddress: sample.AccAddress(),
			DymintPubKey:     pkAny,
			RollappId:        rollappIDs[i],
			Description: types.Description{
				Moniker:         "moniker",
				Identity:        "keybase:username",
				Website:         "http://example.com",
				SecurityContact: "security@example.com",
				Details:         "Additional details about the validator.",
			},
			Status:   types.Bonded,
			Proposer: true,
			Tokens:   sdk.NewCoins(sdk.NewInt64Coin("dym", 100)),
		}
		oldSequencers[i] = oldSequencer
		newSequencers[i] = v4.ConvertOldSequencerToNew(oldSequencer)
	}
	sort.Slice(oldSequencers, func(i, j int) bool {
		return oldSequencers[i].SequencerAddress < oldSequencers[j].SequencerAddress
	})
	sort.Slice(newSequencers, func(i, j int) bool {
		return newSequencers[i].Address < newSequencers[j].Address
	})
}
