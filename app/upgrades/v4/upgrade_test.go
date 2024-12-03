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
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/v3/app"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	v4 "github.com/dymensionxyz/dymension/v3/app/upgrades/v4"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	delayedacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	streamertypes "github.com/dymensionxyz/dymension/v3/x/streamer/types"
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
	dummyUpgradeHeight                      int64 = 5
	expectDelayedackDeletePacketsEpochLimit int32 = 1000_000
	expectDelayedackEpochIdentifier               = "hour"

	expectLivenessSlashInterval = rollapptypes.DefaultLivenessSlashInterval
	expectLivenessSlashBlock    = rollapptypes.DefaultLivenessSlashBlocks
	expectDisputePeriodInBlocks = 3
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

				s.seedPendingRollappPackets()

				s.seedRollappFinalizationQueue()

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

				// Check rollapp gauges
				if err = s.validateRollappGaugesMigration(); err != nil {
					return
				}

				// Check rollapp packets
				if err = s.validateDelayedAckIndexMigration(); err != nil {
					return
				}

				// Check rollapp gauges
				if err = s.validateRollappGaugesMigration(); err != nil {
					return
				}

				// Check rollapp finalization queue
				s.validateRollappFinalizationQueue()

				s.validateNonFinalizedStateInfos()

				s.validateStreamerMigration()

				s.validateModulePermissions()

				return
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

func (s *UpgradeTestSuite) validateModulePermissions() {
	// a bit hacky : just check at least one concrete example of a permission upgrade
	acc := s.App.AccountKeeper.GetModuleAccount(s.Ctx, rollapptypes.ModuleName)
	s.Require().True(acc.HasPermission(authtypes.Burner))
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
	cond := rollappParams.DisputePeriodInBlocks == expectDisputePeriodInBlocks

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
		rollappID := rollapptypes.MustNewChainID(rollapp.RollappId)
		// check that the rollapp can be retrieved by EIP155 key
		if _, ok := s.App.RollappKeeper.GetRollappByEIP155(s.Ctx, rollappID.GetEIP155ID()); !ok {
			return fmt.Errorf("rollapp by EIP155 not found")
		}
	}

	s.Require().Equal(expectLivenessSlashBlock, s.App.RollappKeeper.GetParams(s.Ctx).LivenessSlashBlocks)
	s.Require().Equal(expectLivenessSlashInterval, s.App.RollappKeeper.GetParams(s.Ctx).LivenessSlashInterval)

	if !reflect.DeepEqual(rollapps, expectRollapps) {
		s.T().Log("Expect rollapps", expectRollapps)
		s.T().Log("Actual rollapps", rollapps)
		return fmt.Errorf("rollapps do not match")
	}
	return nil
}

// validate rollapp gauges
func (s *UpgradeTestSuite) validateRollappGaugesMigration() error {
	rollappMap := make(map[string]bool) // Create a map to store rollappId<->gaugeCreated

	rollapps := s.App.RollappKeeper.GetAllRollapps(s.Ctx)
	for _, rollapp := range rollapps {
		rollappMap[rollapp.RollappId] = false // false until gauge is validated
	}

	gauges := s.App.IncentivesKeeper.GetGauges(s.Ctx)
	if len(gauges) != len(rollapps) {
		return fmt.Errorf("rollapp gauges not created for all rollapps")
	}

	// Check that for each rollapp there exists a rollapp gauge
	for _, gauge := range gauges {
		if gauge.GetRollapp() != nil {
			gaugeExists, ok := rollappMap[gauge.GetRollapp().RollappId]
			if !ok {
				return fmt.Errorf("rollapp gauge for unknown rollapp %s", gauge.GetRollapp().RollappId)
			}

			if gaugeExists {
				return fmt.Errorf("rollapp gauge for rollapp %s already created", gauge.GetRollapp().RollappId)
			}

			rollappMap[gauge.GetRollapp().RollappId] = true
		}
	}

	return nil
}

func (s *UpgradeTestSuite) validateSequencersMigration(numSeq int) error {
	testSeqs := s.seedSequencers(numSeq)
	expectSequencers := make([]sequencertypes.Sequencer, len(testSeqs))
	for i, sequencer := range testSeqs {
		expectSequencers[i] = v4.ConvertOldSequencerToNew(sequencer)
	}
	sequencers := s.App.SequencerKeeper.AllSequencers(s.Ctx)
	s.Require().Len(sequencers, len(expectSequencers))

	sort.Slice(sequencers, func(i, j int) bool {
		return sequencers[i].Address < sequencers[j].Address
	})

	sort.Slice(expectSequencers, func(i, j int) bool {
		return expectSequencers[i].Address < expectSequencers[j].Address
	})

	for i, sequencer := range sequencers {
		// check that the sequencer can be retrieved by address
		_, err := s.App.SequencerKeeper.RealSequencer(s.Ctx, sequencer.Address)
		if err != nil {
			return err
		}

		seq := s.App.AppCodec().MustMarshalJSON(&sequencer)
		nSeq := s.App.AppCodec().MustMarshalJSON(&expectSequencers[i])

		s.Require().True(sequencer.OptedIn)
		s.Require().JSONEq(string(seq), string(nSeq))

		byDymintAddr, err := s.App.SequencerKeeper.SequencerByDymintAddr(s.Ctx, expectSequencers[i].MustProposerAddr())
		s.Require().NoError(err)
		s.Require().Equal(sequencer.Address, byDymintAddr.Address)
	}

	// check proposer
	for _, rollapp := range s.App.RollappKeeper.GetAllRollapps(s.Ctx) {
		p := s.App.SequencerKeeper.GetProposer(s.Ctx, rollapp.RollappId)
		s.Require().False(p.Sentinel())
	}
	s.Require().Equal(sequencertypes.DefaultNoticePeriod, s.App.SequencerKeeper.GetParams(s.Ctx).NoticePeriod)
	s.Require().Equal(sequencertypes.DefaultKickThreshold, s.App.SequencerKeeper.GetParams(s.Ctx).KickThreshold)
	s.Require().Equal(sequencertypes.DefaultLivenessSlashMultiplier, s.App.SequencerKeeper.GetParams(s.Ctx).LivenessSlashMinMultiplier)
	s.Require().Equal(sequencertypes.DefaultLivenessSlashMinAbsolute, s.App.SequencerKeeper.GetParams(s.Ctx).LivenessSlashMinAbsolute)

	return nil
}

func (s *UpgradeTestSuite) validateStreamerMigration() {
	epochInfos := s.App.EpochsKeeper.AllEpochInfos(s.Ctx)

	pointers, err := s.App.StreamerKeeper.GetAllEpochPointers(s.Ctx)
	s.Require().NoError(err)

	var expected []streamertypes.EpochPointer
	for _, info := range epochInfos {
		expected = append(expected, streamertypes.NewEpochPointer(info.Identifier, info.Duration))
	}

	// Equal also checks the order of pointers
	s.Require().Equal(expected, pointers)
}

func (s *UpgradeTestSuite) validateDelayedAckIndexMigration() error {
	packets := s.App.DelayedAckKeeper.ListRollappPackets(s.Ctx, delayedacktypes.ByStatus(commontypes.Status_PENDING))
	actual, err := s.App.DelayedAckKeeper.GetPendingPacketsByAddress(s.Ctx, apptesting.TestPacketReceiver)
	s.Require().NoError(err)
	s.Require().Equal(len(packets), len(actual))
	return nil
}

func (s *UpgradeTestSuite) validateRollappFinalizationQueue() {
	queue, err := s.App.RollappKeeper.GetEntireFinalizationQueue(s.Ctx)
	s.Require().NoError(err)

	s.Require().Equal([]rollapptypes.BlockHeightToFinalizationQueue{
		{
			CreationHeight: 1,
			FinalizationQueue: []rollapptypes.StateInfoIndex{
				{RollappId: rollappIDFromIdx(1), Index: 1},
				{RollappId: rollappIDFromIdx(1), Index: 2},
			},
			RollappId: rollappIDFromIdx(1),
		},
		{
			CreationHeight: 1,
			FinalizationQueue: []rollapptypes.StateInfoIndex{
				{RollappId: rollappIDFromIdx(2), Index: 1},
				{RollappId: rollappIDFromIdx(2), Index: 2},
			},
			RollappId: rollappIDFromIdx(2),
		},
		{
			CreationHeight: 1,
			FinalizationQueue: []rollapptypes.StateInfoIndex{
				{RollappId: rollappIDFromIdx(3), Index: 1},
			},
			RollappId: rollappIDFromIdx(3),
		},
		{
			CreationHeight: 2,
			FinalizationQueue: []rollapptypes.StateInfoIndex{
				{RollappId: rollappIDFromIdx(1), Index: 3},
			},
			RollappId: rollappIDFromIdx(1),
		},
		{
			CreationHeight: 2,
			FinalizationQueue: []rollapptypes.StateInfoIndex{
				{RollappId: rollappIDFromIdx(3), Index: 2},
			},
			RollappId: rollappIDFromIdx(3),
		},
		{
			CreationHeight: 3,
			FinalizationQueue: []rollapptypes.StateInfoIndex{
				{RollappId: rollappIDFromIdx(3), Index: 3},
				{RollappId: rollappIDFromIdx(3), Index: 4},
			},
			RollappId: rollappIDFromIdx(3),
		},
	}, queue)
}

func (s *UpgradeTestSuite) validateNonFinalizedStateInfos() {
	queue, err := s.App.RollappKeeper.GetEntireFinalizationQueue(s.Ctx)
	s.Require().NoError(err)

	for _, q := range queue {
		proposer := s.App.SequencerKeeper.GetProposer(s.Ctx, q.RollappId)
		for _, stateInfoIndex := range q.FinalizationQueue {
			stateInfo, found := s.App.RollappKeeper.GetStateInfo(s.Ctx, stateInfoIndex.RollappId, stateInfoIndex.Index)
			s.Require().True(found)

			// Verify that all non-finalized state infos contain the correct proposer (the same that's set in x/sequencer)
			s.Require().Equal(proposer.Address, stateInfo.NextProposer)
		}
	}
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
			RollappId:    rollappID,
			Owner:        sample.AccAddressFromSecret(rollappID),
			GenesisState: rollapptypes.RollappGenesisState{},
			ChannelId:    fmt.Sprintf("channel-%d", i),
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

func (s *UpgradeTestSuite) seedSequencers(numRollapps int) []sequencertypes.Sequencer {
	numSeqPerRollapp := numRollapps
	sequencers := make([]sequencertypes.Sequencer, 0, numSeqPerRollapp*numRollapps)
	for i := 0; i < numRollapps; i++ {
		rollappID := rollappIDFromIdx(i)

		for j := 0; j < numSeqPerRollapp; j++ {
			uuid := fmt.Sprintf("sequencer-%d-%d", i, j)
			pk := ed25519.GenPrivKeyFromSecret([]byte(uuid)).PubKey()
			pkAny, _ := codectypes.NewAnyWithValue(pk)
			sequencer := sequencertypes.Sequencer{
				Address:      sdk.AccAddress(pk.Address()).String(),
				DymintPubKey: pkAny,
				RollappId:    rollappID,
				Metadata: sequencertypes.SequencerMetadata{
					Moniker: uuid,
					Details: fmt.Sprintf("Additional details about the %s", uuid),
				},
				Status:   sequencertypes.Bonded,
				Tokens:   sdk.NewCoins(sdk.NewInt64Coin("dym", 100)),
				Proposer: j == 0, // first sequencer is proposer
			}
			sequencers = append(sequencers, sequencer)
		}
	}
	return sequencers
}

func rollappIDFromIdx(idx int) string {
	return fmt.Sprintf("roll%spp_123%d-1", string(rune(idx+'a')), idx+1)
}

func (s *UpgradeTestSuite) seedPendingRollappPackets() {
	packets := apptesting.GenerateRollappPackets(s.T(), "testrollappid_1-1", 20)
	for _, packet := range packets {
		s.App.DelayedAckKeeper.SetRollappPacket(s.Ctx, packet)
	}
}

func (s *UpgradeTestSuite) seedRollappFinalizationQueue() {
	q1 := rollapptypes.BlockHeightToFinalizationQueue{
		CreationHeight: 1,
		FinalizationQueue: []rollapptypes.StateInfoIndex{
			{RollappId: rollappIDFromIdx(1), Index: 1},
			{RollappId: rollappIDFromIdx(1), Index: 2},
			{RollappId: rollappIDFromIdx(2), Index: 1},
			{RollappId: rollappIDFromIdx(2), Index: 2},
			{RollappId: rollappIDFromIdx(3), Index: 1},
		},
		RollappId: "",
	}
	q2 := rollapptypes.BlockHeightToFinalizationQueue{
		CreationHeight: 2,
		FinalizationQueue: []rollapptypes.StateInfoIndex{
			{RollappId: rollappIDFromIdx(1), Index: 3},
			{RollappId: rollappIDFromIdx(3), Index: 2},
		},
		RollappId: "",
	}
	q3 := rollapptypes.BlockHeightToFinalizationQueue{
		CreationHeight: 3,
		FinalizationQueue: []rollapptypes.StateInfoIndex{
			{RollappId: rollappIDFromIdx(3), Index: 3},
			{RollappId: rollappIDFromIdx(3), Index: 4},
		},
		RollappId: "",
	}

	s.App.RollappKeeper.SetBlockHeightToFinalizationQueue(s.Ctx, q1)
	s.App.RollappKeeper.SetBlockHeightToFinalizationQueue(s.Ctx, q2)
	s.App.RollappKeeper.SetBlockHeightToFinalizationQueue(s.Ctx, q3)

	stateInfos := []rollapptypes.StateInfo{
		generateStateInfo(1, 1),
		generateStateInfo(1, 2),
		generateStateInfo(1, 3),
		generateStateInfo(2, 1),
		generateStateInfo(2, 2),
		generateStateInfo(3, 1),
		generateStateInfo(3, 2),
		generateStateInfo(3, 3),
		generateStateInfo(3, 4),
	}

	for _, stateInfo := range stateInfos {
		s.App.RollappKeeper.SetStateInfo(s.Ctx, stateInfo)
	}
}

func generateStateInfo(rollappIdx, stateIdx int) rollapptypes.StateInfo {
	return rollapptypes.StateInfo{
		StateInfoIndex: rollapptypes.StateInfoIndex{
			RollappId: rollappIDFromIdx(rollappIdx),
			Index:     uint64(stateIdx),
		},
	}
}

func TestReformatFinalizationQueue(t *testing.T) {
	q := rollapptypes.BlockHeightToFinalizationQueue{
		CreationHeight: 1,
		FinalizationQueue: []rollapptypes.StateInfoIndex{
			{RollappId: rollappIDFromIdx(1), Index: 1},
			{RollappId: rollappIDFromIdx(1), Index: 2},
			{RollappId: rollappIDFromIdx(1), Index: 3},
			{RollappId: rollappIDFromIdx(2), Index: 1},
			{RollappId: rollappIDFromIdx(2), Index: 2},
			{RollappId: rollappIDFromIdx(3), Index: 1},
		},
		RollappId: "", // empty for old-style queues
	}

	newQueues := v4.ReformatFinalizationQueue(q)

	require.Equal(t, []rollapptypes.BlockHeightToFinalizationQueue{
		{
			CreationHeight: 1,
			FinalizationQueue: []rollapptypes.StateInfoIndex{
				{RollappId: rollappIDFromIdx(1), Index: 1},
				{RollappId: rollappIDFromIdx(1), Index: 2},
				{RollappId: rollappIDFromIdx(1), Index: 3},
			},
			RollappId: rollappIDFromIdx(1),
		},
		{
			CreationHeight: 1,
			FinalizationQueue: []rollapptypes.StateInfoIndex{
				{RollappId: rollappIDFromIdx(2), Index: 1},
				{RollappId: rollappIDFromIdx(2), Index: 2},
			},
			RollappId: rollappIDFromIdx(2),
		},
		{
			CreationHeight: 1,
			FinalizationQueue: []rollapptypes.StateInfoIndex{
				{RollappId: rollappIDFromIdx(3), Index: 1},
			},
			RollappId: rollappIDFromIdx(3),
		},
	}, newQueues)
}
