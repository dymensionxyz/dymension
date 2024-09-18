package keeper_test

import (
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	fpslices "github.com/dymensionxyz/dymension/v3/utils/fp/slices"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (s *RollappTestSuite) TestMarkVulnerableRollapps() {
	type rollapp struct {
		name       string
		drsVersion string
	}
	govModule := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	tests := []struct {
		name         string
		authority    string
		rollapps     []rollapp
		vulnVersions []string
		expError     error
	}{
		{
			name:      "happy path 1",
			authority: govModule,
			rollapps: []rollapp{
				{name: "rollappa_1-1", drsVersion: "drs_non_vuln_1"},
				{name: "rollappb_2-1", drsVersion: "drs_non_vuln_1"},
				{name: "rollappc_3-1", drsVersion: "drs_non_vuln_2"},
				{name: "rollappd_4-1", drsVersion: "drs_vuln_1"},
				{name: "rollappe_5-1", drsVersion: "drs_vuln_1"},
				{name: "rollappf_6-1", drsVersion: "drs_vuln_2"},
			},
			vulnVersions: []string{
				"drs_vuln_1",
				"drs_vuln_2",
			},
			expError: nil,
		},
		{
			name:      "happy path 2",
			authority: govModule,
			rollapps: []rollapp{
				{name: "rollappa_1-1", drsVersion: "drs_non_vuln_1"},
				{name: "rollappd_2-1", drsVersion: "drs_vuln_1"},
			},
			vulnVersions: []string{
				"drs_vuln_1",
			},
			expError: nil,
		},
		{
			name:      "some legacy rollapps don't have a DRS version",
			authority: govModule,
			rollapps: []rollapp{
				{name: "rollappa_1-1", drsVersion: "drs_non_vuln_1"},
				{name: "rollappb_2-1", drsVersion: ""},
				{name: "rollappc_3-1", drsVersion: "drs_non_vuln_2"},
				{name: "rollappd_4-1", drsVersion: "drs_vuln_1"},
				{name: "rollappe_5-1", drsVersion: ""},
				{name: "rollappf_6-1", drsVersion: "drs_vuln_2"},
			},
			vulnVersions: []string{
				"drs_vuln_1",
				"drs_vuln_2",
			},
			expError: nil,
		},
		{
			name:      "empty DRS version is also vulnerable",
			authority: govModule,
			rollapps: []rollapp{
				{name: "rollappa_1-1", drsVersion: "drs_non_vuln_1"},
				{name: "rollappb_2-1", drsVersion: ""},
				{name: "rollappc_3-1", drsVersion: "drs_non_vuln_2"},
				{name: "rollappd_4-1", drsVersion: "drs_vuln_1"},
				{name: "rollappe_5-1", drsVersion: ""},
				{name: "rollappf_6-1", drsVersion: "drs_vuln_2"},
			},
			vulnVersions: []string{
				"",
				"drs_vuln_1",
				"drs_vuln_2",
			},
			expError: nil,
		},
		{
			name:      "invalid authority",
			authority: apptesting.CreateRandomAccounts(1)[0].String(),
			rollapps: []rollapp{
				{name: "rollappa_1-1", drsVersion: "drs_non_vuln_1"},
				{name: "rollappe_2-1", drsVersion: "drs_vuln_1"},
			},
			vulnVersions: []string{
				"drs_vuln_1",
			},
			expError: gerrc.ErrInvalidArgument,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest()

			// prepare test data
			vulnVersions := fpslices.ToKeySet(tc.vulnVersions)
			// list of expected vulnerable rollapps
			expectedVulnRollappIDs := make([]string, 0, len(tc.vulnVersions))
			// list of expected non-vulnerable rollapps
			expectedNonVulnRollappIDs := make([]string, 0, len(tc.vulnVersions))
			// create rollapps for every rollapp record from the test case
			for _, ra := range tc.rollapps {
				// create a rollapp
				s.CreateRollappByName(ra.name)
				// create a sequencer
				proposer := s.CreateDefaultSequencer(s.Ctx, ra.name)

				// create a new update
				_, err := s.PostStateUpdateWithDRSVersion(s.Ctx, ra.name, proposer, 1, uint64(3), ra.drsVersion)
				s.Require().Nil(err)

				// fill lists with expectations
				if _, vuln := vulnVersions[ra.drsVersion]; vuln {
					expectedVulnRollappIDs = append(expectedVulnRollappIDs, ra.name)
				} else {
					expectedNonVulnRollappIDs = append(expectedNonVulnRollappIDs, ra.name)
				}
			}

			// trigger the message we want to test
			_, err := s.msgServer.MarkVulnerableRollapps(s.Ctx, &types.MsgMarkVulnerableRollapps{
				Authority:   tc.authority,
				DrsVersions: tc.vulnVersions,
			})

			// validate results
			if tc.expError != nil {
				s.Error(err)
				// TODO: try using errors.Is!
				s.ErrorContains(err, tc.expError.Error())

				// check the event is not emitted
				eventName := proto.MessageName(new(types.EventMarkVulnerableRollapps))
				s.AssertEventEmitted(s.Ctx, eventName, 0)

				// check non-vulnerable rollapps: all rollapps are still non-vulnerable
				nonVulnRa := s.App.RollappKeeper.FilterRollapps(s.Ctx, keeper.FilterActive)
				actualNonVulnRollappIDs := fpslices.Map(nonVulnRa, func(r types.Rollapp) string { return r.RollappId })
				allRollapps := fpslices.Merge(expectedVulnRollappIDs, expectedNonVulnRollappIDs)
				s.ElementsMatch(allRollapps, actualNonVulnRollappIDs)

				// check vulnerable rollapps: no vulnerable rollapps
				vulnRa := s.App.RollappKeeper.FilterRollapps(s.Ctx, FilterVulnerable)
				s.Empty(vulnRa)

				// check the vulnerable version set is empty
				actualVulnVersions, err := s.App.RollappKeeper.GetAllVulnerableDRSVersions(s.Ctx)
				s.Require().NoError(err)
				s.Require().Empty(actualVulnVersions)
			} else {
				s.NoError(err)

				// check the event is emitted
				eventName := proto.MessageName(new(types.EventMarkVulnerableRollapps))
				s.AssertEventEmitted(s.Ctx, eventName, 1)

				// check non-vulnerable rollapps
				nonVulnRa := s.App.RollappKeeper.FilterRollapps(s.Ctx, keeper.FilterActive)
				actualNonVulnRollappIDs := fpslices.Map(nonVulnRa, func(r types.Rollapp) string { return r.RollappId })
				s.ElementsMatch(expectedNonVulnRollappIDs, actualNonVulnRollappIDs)

				// check vulnerable rollapps
				vulnRa := s.App.RollappKeeper.FilterRollapps(s.Ctx, FilterVulnerable)
				actualVulnRollappIDs := fpslices.Map(vulnRa, func(r types.Rollapp) string { return r.RollappId })
				s.ElementsMatch(expectedVulnRollappIDs, actualVulnRollappIDs)

				// check the vulnerable version set
				actualVulnVersions, err := s.App.RollappKeeper.GetAllVulnerableDRSVersions(s.Ctx)
				s.Require().NoError(err)
				s.Require().ElementsMatch(tc.vulnVersions, actualVulnVersions)
			}
		})
	}
}

func FilterVulnerable(b types.Rollapp) bool {
	return b.Frozen
}
