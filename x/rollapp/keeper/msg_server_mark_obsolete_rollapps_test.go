package keeper_test

import (
	"slices"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uslice"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (s *RollappTestSuite) TestMarkObsoleteRollapps() {
	type rollapp struct {
		name       string
		drsVersion uint32
	}
	govModule := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	tests := []struct {
		name             string
		authority        string
		rollapps         []rollapp
		obsoleteVersions []uint32
		expError         error
	}{
		{
			name:      "happy path 1",
			authority: govModule,
			rollapps: []rollapp{
				{name: "rollappa_1-1", drsVersion: 3},
				{name: "rollappb_2-1", drsVersion: 3},
				{name: "rollappc_3-1", drsVersion: 4},
				{name: "rollappd_4-1", drsVersion: 1},
				{name: "rollappe_5-1", drsVersion: 1},
				{name: "rollappf_6-1", drsVersion: 2},
			},
			obsoleteVersions: []uint32{
				1,
				2,
			},
			expError: nil,
		},
		{
			name:      "happy path 2",
			authority: govModule,
			rollapps: []rollapp{
				{name: "rollappa_1-1", drsVersion: 2},
				{name: "rollappd_2-1", drsVersion: 1},
			},
			obsoleteVersions: []uint32{
				1,
			},
			expError: nil,
		},
		{
			name:      "some legacy rollapps don't have a DRS version",
			authority: govModule,
			rollapps: []rollapp{
				{name: "rollappa_1-1", drsVersion: 2},
				{name: "rollappb_2-1", drsVersion: 0},
				{name: "rollappc_3-1", drsVersion: 3},
				{name: "rollappd_4-1", drsVersion: 1},
				{name: "rollappe_5-1", drsVersion: 0},
				{name: "rollappf_6-1", drsVersion: 2},
			},
			obsoleteVersions: []uint32{
				1,
				2,
			},
			expError: nil,
		},
		{
			name:      "empty DRS version is also obsolete",
			authority: govModule,
			rollapps: []rollapp{
				{name: "rollappa_1-1", drsVersion: 3},
				{name: "rollappb_2-1", drsVersion: 0},
				{name: "rollappc_3-1", drsVersion: 4},
				{name: "rollappd_4-1", drsVersion: 1},
				{name: "rollappe_5-1", drsVersion: 0},
				{name: "rollappf_6-1", drsVersion: 1},
			},
			obsoleteVersions: []uint32{
				0,
				1,
				2,
			},
			expError: nil,
		},
		{
			name:      "invalid authority",
			authority: apptesting.CreateRandomAccounts(1)[0].String(),
			rollapps: []rollapp{
				{name: "rollappa_1-1", drsVersion: 2},
				{name: "rollappe_2-1", drsVersion: 1},
			},
			obsoleteVersions: []uint32{
				1,
			},
			expError: gerrc.ErrInvalidArgument,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest()
			s.App.RollappKeeper.SetHooks(nil) // disable hooks

			// prepare test data
			obsoleteVersions := uslice.ToKeySet(tc.obsoleteVersions)
			// list of expected obsolete rollapps
			expectForkedIDs := make([]string, 0, len(tc.rollapps))
			// list of expected non-obsolete rollapps
			expetNotForkedIDs := make([]string, 0, len(tc.rollapps))
			// create rollapps for every rollapp record from the test case
			for _, ra := range tc.rollapps {
				// create a rollapp
				s.CreateRollappByName(ra.name)
				// create a sequencer
				proposer := s.CreateDefaultSequencer(s.Ctx, ra.name)

				// create a new update
				_, err := s.PostStateUpdateWithDRSVersion(s.Ctx, ra.name, proposer, 1, uint64(3), ra.drsVersion)
				s.Require().NoError(err)

				// fill lists with expectations
				if _, obsolete := obsoleteVersions[ra.drsVersion]; obsolete {
					expectForkedIDs = append(expectForkedIDs, ra.name)
				} else {
					expetNotForkedIDs = append(expetNotForkedIDs, ra.name)
				}
			}

			// trigger the message we want to test
			_, err := s.msgServer.MarkObsoleteRollapps(s.Ctx, &types.MsgMarkObsoleteRollapps{
				Authority:   tc.authority,
				DrsVersions: tc.obsoleteVersions,
			})

			// validate results
			if tc.expError != nil {
				s.ErrorIs(err, tc.expError)

				// check the event is not emitted
				eventName := proto.MessageName(new(types.EventMarkObsoleteRollapps))
				s.AssertEventEmitted(s.Ctx, eventName, 0)

				nonForked := s.App.RollappKeeper.FilterRollapps(s.Ctx, FilterNotForked)
				actualNonForkedIDs := uslice.Map(nonForked, func(r types.Rollapp) string { return r.RollappId })
				allRollapps := slices.Concat(expectForkedIDs, expetNotForkedIDs)
				s.ElementsMatch(allRollapps, actualNonForkedIDs)

				// check obsolete rollapps: no obsolete rollapps
				obsoleteRa := s.App.RollappKeeper.FilterRollapps(s.Ctx, FilterForked)
				s.Empty(obsoleteRa)

				// check the obsolete version set is empty
				actualObsoleteVersions, err := s.App.RollappKeeper.GetAllObsoleteDRSVersions(s.Ctx)
				s.Require().NoError(err)
				s.Require().Empty(actualObsoleteVersions)
			} else {
				s.NoError(err)

				// check the event is emitted
				eventName := proto.MessageName(new(types.EventMarkObsoleteRollapps))
				s.AssertEventEmitted(s.Ctx, eventName, 1)

				// check non-obsolete rollapps
				notForked := s.App.RollappKeeper.FilterRollapps(s.Ctx, FilterNotForked)
				actualNotForkedIDs := uslice.Map(notForked, func(r types.Rollapp) string { return r.RollappId })
				s.ElementsMatch(expetNotForkedIDs, actualNotForkedIDs)

				// check obsolete rollapps
				forked := s.App.RollappKeeper.FilterRollapps(s.Ctx, FilterForked)
				actualForkedIDs := uslice.Map(forked, func(r types.Rollapp) string { return r.RollappId })
				s.ElementsMatch(expectForkedIDs, actualForkedIDs)

				// check the obsolete version set
				actualObsoleteVersions, err := s.App.RollappKeeper.GetAllObsoleteDRSVersions(s.Ctx)
				s.Require().NoError(err)
				s.Require().ElementsMatch(tc.obsoleteVersions, actualObsoleteVersions)
			}
		})
	}
}

func FilterForked(b types.Rollapp) bool {
	return !FilterNotForked(b)
}

func FilterNotForked(b types.Rollapp) bool {
	return b.RevisionNumber == 0
}
