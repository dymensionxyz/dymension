package keeper_test

import (
	"testing"

	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/stretchr/testify/require"
)

func (s *KeeperTestSuite) TestKeeper_GetSetAliasForRollAppId() {
	rollApp1 := *newRollApp("rollapp_1-1").WithAlias("al1")
	rollApp2 := *newRollApp("rolling_2-2").WithAlias("al2")
	rollApp3NotExists := *newRollApp("nah_0-0").WithAlias("al3")

	for i, ra := range []rollapp{rollApp1, rollApp2} {
		s.rollAppKeeper.SetRollapp(s.ctx, rollapptypes.Rollapp{
			RollappId: ra.rollAppId,
			Owner:     testAddr(uint64(i)).bech32(),
		})
	}

	s.Run("set - can set", func() {
		s.Require().True(s.dymNsKeeper.IsRollAppId(s.ctx, rollApp1.rollAppId), "must be a RollApp, just not set alias")

		err := s.dymNsKeeper.SetAliasForRollAppId(s.ctx, rollApp1.rollAppId, rollApp1.alias)
		s.Require().NoError(err)

		alias, found := s.dymNsKeeper.GetAliasByRollAppId(s.ctx, rollApp1.rollAppId)
		s.Equal(rollApp1.alias, alias)
		s.True(found)

		rollAppId, found := s.dymNsKeeper.GetRollAppIdByAlias(s.ctx, rollApp1.alias)
		s.Equal(rollApp1.rollAppId, rollAppId)
		s.True(found)
	})

	s.Run("set - can NOT set if alias is being in-used by another RollApp", func() {
		rollAppId, found := s.dymNsKeeper.GetRollAppIdByAlias(s.ctx, rollApp1.alias)
		s.Equal(rollApp1.rollAppId, rollAppId)
		s.True(found)

		err := s.dymNsKeeper.SetAliasForRollAppId(s.ctx, rollApp2.rollAppId, rollApp1.alias)
		s.requireErrorContains(err, "alias currently being in used by:")
	})

	s.Run("set - reject bad chain-id", func() {
		err := s.dymNsKeeper.SetAliasForRollAppId(s.ctx, "bad@", "alias")
		s.Error(err)
	})

	s.Run("set - reject bad alias", func() {
		s.Require().True(s.dymNsKeeper.IsRollAppId(s.ctx, rollApp2.rollAppId), "must be a RollApp")

		err := s.dymNsKeeper.SetAliasForRollAppId(s.ctx, rollApp2.rollAppId, "")
		s.requireErrorContains(err, "invalid alias")

		err = s.dymNsKeeper.SetAliasForRollAppId(s.ctx, rollApp2.rollAppId, "@")
		s.requireErrorContains(err, "invalid alias")
	})

	s.Run("get - of existing RollApp but no alias set", func() {
		s.Require().True(s.dymNsKeeper.IsRollAppId(s.ctx, rollApp2.rollAppId), "must be a RollApp, just not set alias")

		alias, found := s.dymNsKeeper.GetAliasByRollAppId(s.ctx, rollApp2.rollAppId)
		s.Empty(alias)
		s.False(found)

		rollAppId, found := s.dymNsKeeper.GetRollAppIdByAlias(s.ctx, rollApp2.alias)
		s.Empty(rollAppId)
		s.False(found)
	})

	s.Run("set - non-exists RollApp returns error", func() {
		s.Require().False(s.dymNsKeeper.IsRollAppId(s.ctx, rollApp3NotExists.rollAppId))

		err := s.dymNsKeeper.SetAliasForRollAppId(s.ctx, rollApp3NotExists.rollAppId, rollApp3NotExists.alias)
		s.requireErrorContains(err, "not a RollApp")
	})

	s.Run("get - non-exists RollApp returns empty", func() {
		s.Require().False(s.dymNsKeeper.IsRollAppId(s.ctx, rollApp3NotExists.rollAppId))

		alias, found := s.dymNsKeeper.GetAliasByRollAppId(s.ctx, rollApp3NotExists.rollAppId)
		s.Empty(alias)
		s.False(found)

		rollAppId, found := s.dymNsKeeper.GetRollAppIdByAlias(s.ctx, rollApp3NotExists.alias)
		s.Empty(rollAppId)
		s.False(found)
	})

	s.Run("set/get - can set multiple alias to a single Roll-App", func() {
		s.SetupTest()

		type testCase struct {
			rollAppId string
			aliases   []string
		}

		testcases := []testCase{
			{
				rollAppId: "rollapp_1-1",
				aliases:   []string{"one", "two", "three"},
			},
			{
				rollAppId: "rollapp_2-2",
				aliases:   []string{"four", "five"},
			},
		}

		for _, tc := range testcases {
			s.rollAppKeeper.SetRollapp(s.ctx, rollapptypes.Rollapp{
				RollappId: tc.rollAppId,
				Owner:     testAddr(0).bech32(),
			})
		}

		for _, tc := range testcases {
			for _, alias := range tc.aliases {
				err := s.dymNsKeeper.SetAliasForRollAppId(s.ctx, tc.rollAppId, alias)
				s.Require().NoError(err)
			}
		}

		for _, tc := range testcases {
			for _, alias := range tc.aliases {
				rollAppId, found := s.dymNsKeeper.GetRollAppIdByAlias(s.ctx, alias)
				s.Equal(tc.rollAppId, rollAppId)
				s.True(found)
			}

			alias, found := s.dymNsKeeper.GetAliasByRollAppId(s.ctx, tc.rollAppId)
			s.True(found)
			s.Contains(tc.aliases, alias)
			s.Equal(alias, tc.aliases[0], "should returns the first one added")
		}
	})
}

func (s *KeeperTestSuite) TestKeeper_GetAliasesOfRollAppId() {
	rollApp1 := *newRollApp("rollapp_1-1").WithAlias("one")
	rollApp2 := *newRollApp("rollapp_2-2").WithAlias("two")
	rollApp3NoAlias := *newRollApp("rollapp_3-3")

	s.persistRollApp(rollApp1)
	s.persistRollApp(rollApp2)
	s.persistRollApp(rollApp3NoAlias)

	err := s.dymNsKeeper.SetAliasForRollAppId(s.ctx, rollApp1.rollAppId, "more")
	s.Require().NoError(err)

	err = s.dymNsKeeper.SetAliasForRollAppId(s.ctx, rollApp1.rollAppId, "alias")
	s.Require().NoError(err)

	aliases := s.dymNsKeeper.GetAliasesOfRollAppId(s.ctx, rollApp1.rollAppId)
	s.Require().Equal([]string{rollApp1.alias, "more", "alias"}, aliases)

	aliases = s.dymNsKeeper.GetAliasesOfRollAppId(s.ctx, rollApp2.rollAppId)
	s.Require().Equal([]string{rollApp2.alias}, aliases)

	aliases = s.dymNsKeeper.GetAliasesOfRollAppId(s.ctx, rollApp3NoAlias.rollAppId)
	s.Require().Empty(aliases)
}

func (s *KeeperTestSuite) TestKeeper_RemoveAliasFromRollAppId() {
	rollApp1 := *newRollApp("rollapp_1-1").WithAlias("al1")
	rollApp2 := *newRollApp("rolling_2-2").WithAlias("al2")
	rollApp3 := *newRollApp("rollapp_3-3").WithAlias("al3")
	rollApp4NoAlias := *newRollApp("noa_4-4")
	rollApp5NotExists := *newRollApp("nah_5-5").WithAlias("al5")

	const aliasOne = "one"
	const aliasTwo = "two"
	const unusedAlias = "unused"

	tests := []struct {
		name            string
		addRollApps     []rollapp
		preRunFunc      func(*KeeperTestSuite)
		inputRollAppId  string
		inputAlias      string
		wantErr         bool
		wantErrContains string
		afterTestFunc   func(*KeeperTestSuite)
	}{
		{
			name:        "pass - can remove",
			addRollApps: []rollapp{rollApp1},
			preRunFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(rollApp1.rollAppId).HasAlias(rollApp1.alias)
			},
			inputRollAppId: rollApp1.rollAppId,
			inputAlias:     rollApp1.alias,
			wantErr:        false,
			afterTestFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(rollApp1.rollAppId).HasNoAlias()
				s.requireAlias(rollApp1.alias).NotInUse()
			},
		},
		{
			name:        "pass - can remove among multiple records",
			addRollApps: []rollapp{rollApp1, rollApp2, rollApp3},
			preRunFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(rollApp1.rollAppId).HasAlias(rollApp1.alias)
				s.requireRollApp(rollApp2.rollAppId).HasAlias(rollApp2.alias)
				s.requireRollApp(rollApp3.rollAppId).HasAlias(rollApp3.alias)
			},
			inputRollAppId: rollApp2.rollAppId,
			inputAlias:     rollApp2.alias,
			wantErr:        false,
			afterTestFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(rollApp2.rollAppId).HasNoAlias()
				s.requireAlias(rollApp2.alias).NotInUse()

				// other records remain unchanged
				s.requireRollApp(rollApp1.rollAppId).HasAlias(rollApp1.alias)
				s.requireRollApp(rollApp3.rollAppId).HasAlias(rollApp3.alias)
			},
		},
		{
			name:        "fail - reject if input RollApp ID is empty",
			addRollApps: []rollapp{rollApp1},
			preRunFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(rollApp1.rollAppId).HasAlias(rollApp1.alias)
			},
			inputRollAppId:  "",
			inputAlias:      rollApp1.alias,
			wantErr:         true,
			wantErrContains: "not a RollApp",
			afterTestFunc: func(s *KeeperTestSuite) {
				// record remains unchanged
				s.requireRollApp(rollApp1.rollAppId).HasAlias(rollApp1.alias)
			},
		},
		{
			name:        "fail - reject if input RollApp ID is not exists",
			addRollApps: []rollapp{rollApp1},
			preRunFunc: func(s *KeeperTestSuite) {
				s.Require().False(s.dymNsKeeper.IsRollAppId(s.ctx, rollApp5NotExists.rollAppId))

				s.requireRollApp(rollApp1.rollAppId).HasAlias(rollApp1.alias)
			},
			inputRollAppId:  rollApp5NotExists.rollAppId,
			inputAlias:      rollApp5NotExists.alias,
			wantErr:         true,
			wantErrContains: "not a RollApp",
			afterTestFunc: func(s *KeeperTestSuite) {
				// other records remain unchanged
				s.requireRollApp(rollApp1.rollAppId).HasAlias(rollApp1.alias)
			},
		},
		{
			name:        "fail - reject if input Alias is empty",
			addRollApps: []rollapp{rollApp1},
			preRunFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(rollApp1.rollAppId).HasAlias(rollApp1.alias)
			},
			inputRollAppId:  rollApp1.rollAppId,
			inputAlias:      "",
			wantErr:         true,
			wantErrContains: "invalid alias",
			afterTestFunc: func(s *KeeperTestSuite) {
				// record remains unchanged
				s.requireRollApp(rollApp1.rollAppId).HasAlias(rollApp1.alias)
			},
		},
		{
			name:        "fail - reject if input Alias is malformed",
			addRollApps: []rollapp{rollApp1},
			preRunFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(rollApp1.rollAppId).HasAlias(rollApp1.alias)
			},
			inputRollAppId:  rollApp1.rollAppId,
			inputAlias:      "@",
			wantErr:         true,
			wantErrContains: "invalid alias",
			afterTestFunc: func(s *KeeperTestSuite) {
				// record remains unchanged
				s.requireRollApp(rollApp1.rollAppId).HasAlias(rollApp1.alias)
			},
		},
		{
			name:        "fail - reject if Roll App has no alias linked",
			addRollApps: []rollapp{rollApp4NoAlias},
			preRunFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(rollApp4NoAlias.rollAppId).HasNoAlias()
			},
			inputRollAppId:  rollApp4NoAlias.rollAppId,
			inputAlias:      aliasOne,
			wantErr:         true,
			wantErrContains: "alias not found",
			afterTestFunc: func(s *KeeperTestSuite) {
				// record remains unchanged
				s.requireRollApp(rollApp4NoAlias.rollAppId).HasNoAlias()
			},
		},
		{
			name:        "fail - reject if Roll App has no alias linked and input alias linked to another Roll App",
			addRollApps: []rollapp{rollApp1, rollApp4NoAlias},
			preRunFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(rollApp1.rollAppId).HasAlias(rollApp1.alias)
				s.requireRollApp(rollApp4NoAlias.rollAppId).HasNoAlias()
			},
			inputRollAppId:  rollApp4NoAlias.rollAppId,
			inputAlias:      rollApp1.alias,
			wantErr:         true,
			wantErrContains: "alias currently being in used by:",
			afterTestFunc: func(s *KeeperTestSuite) {
				// records remain unchanged
				s.requireRollApp(rollApp1.rollAppId).HasAlias(rollApp1.alias)
				s.requireRollApp(rollApp4NoAlias.rollAppId).HasNoAlias()
			},
		},
		{
			name:        "fail - reject if remove alias linked to another Roll App",
			addRollApps: []rollapp{rollApp1, rollApp2},
			preRunFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(rollApp1.rollAppId).HasAlias(rollApp1.alias)
				s.requireRollApp(rollApp2.rollAppId).HasAlias(rollApp2.alias)
			},
			inputRollAppId:  rollApp1.rollAppId,
			inputAlias:      rollApp2.alias,
			wantErr:         true,
			wantErrContains: "alias currently being in used by",
			afterTestFunc: func(s *KeeperTestSuite) {
				// records remain unchanged
				s.requireRollApp(rollApp1.rollAppId).HasAlias(rollApp1.alias)
				s.requireRollApp(rollApp2.rollAppId).HasAlias(rollApp2.alias)
			},
		},
		{
			name:        "fail - reject if input alias does not link to any Roll App",
			addRollApps: []rollapp{rollApp1},
			preRunFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(rollApp1.rollAppId).HasAlias(rollApp1.alias)
			},
			inputRollAppId:  rollApp1.rollAppId,
			inputAlias:      unusedAlias,
			wantErr:         true,
			wantErrContains: "alias not found",
			afterTestFunc: func(s *KeeperTestSuite) {
				// records remain unchanged
				s.requireRollApp(rollApp1.rollAppId).HasAlias(rollApp1.alias)
			},
		},
		{
			name:        "pass - remove alias correctly among multiple aliases linked to a Roll App",
			addRollApps: []rollapp{rollApp1},
			preRunFunc: func(s *KeeperTestSuite) {
				s.Require().NoError(s.dymNsKeeper.SetAliasForRollAppId(s.ctx, rollApp1.rollAppId, aliasOne))
				s.Require().NoError(s.dymNsKeeper.SetAliasForRollAppId(s.ctx, rollApp1.rollAppId, aliasTwo))

				s.requireRollApp(rollApp1.rollAppId).HasAlias(rollApp1.alias, aliasOne, aliasTwo)
			},
			inputRollAppId: rollApp1.rollAppId,
			inputAlias:     aliasOne,
			wantErr:        false,
			afterTestFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(rollApp1.rollAppId).HasAlias(rollApp1.alias, aliasTwo)
				s.requireAlias(aliasOne).NotInUse()
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()

			for _, ra := range tt.addRollApps {
				s.persistRollApp(ra)
			}

			if tt.preRunFunc != nil {
				tt.preRunFunc(s)
			}

			err := s.dymNsKeeper.RemoveAliasFromRollAppId(s.ctx, tt.inputRollAppId, tt.inputAlias)

			defer func() {
				if s.T().Failed() {
					return
				}
				if tt.afterTestFunc != nil {
					tt.afterTestFunc(s)
				}
			}()

			if tt.wantErr {
				s.requireErrorContains(err, tt.wantErrContains)
				return
			}

			s.Require().NoError(err)
		})
	}
}

func TestKeeper_MoveAliasToRollAppId(t *testing.T) {
	type rollapp struct {
		rollAppId string
		alias     string
	}
	rollApp1 := rollapp{
		rollAppId: "rollapp_1-1",
		alias:     "al1",
	}
	rollApp2 := rollapp{
		rollAppId: "rolling_2-2",
		alias:     "al2",
	}
	rollApp3WithoutAlias := rollapp{
		rollAppId: "rollapp_3-3",
		alias:     "",
	}
	rollApp4WithoutAlias := rollapp{
		rollAppId: "rollapp_4-4",
		alias:     "",
	}

	tests := []struct {
		name            string
		rollapps        []rollapp
		srcRollAppId    string
		alias           string
		dstRollAppId    string
		preTestFunc     func(*testing.T, sdk.Context, dymnskeeper.Keeper, rollappkeeper.Keeper)
		wantErr         bool
		wantErrContains string
		afterTestFunc   func(*testing.T, sdk.Context, dymnskeeper.Keeper, rollappkeeper.Keeper)
	}{
		{
			name:         "pass - can move",
			rollapps:     []rollapp{rollApp1, rollApp2},
			srcRollAppId: rollApp1.rollAppId,
			alias:        rollApp1.alias,
			dstRollAppId: rollApp2.rollAppId,
			preTestFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				requireAssignedAliasPairs(rollApp1.rollAppId, rollApp1.alias, t, ctx, dk)
				requireAssignedAliasPairs(rollApp2.rollAppId, rollApp2.alias, t, ctx, dk)
			},
			wantErr: false,
			afterTestFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				requireRollAppHasNoAlias(rollApp1.rollAppId, t, ctx, dk)
				requireAliasLinkedToRollApp(rollApp1.alias, rollApp2.rollAppId, t, ctx, dk)
			},
		},
		{
			name:         "pass - can move to RollApp with existing Alias",
			rollapps:     []rollapp{rollApp1, rollApp2},
			srcRollAppId: rollApp1.rollAppId,
			alias:        rollApp1.alias,
			dstRollAppId: rollApp2.rollAppId,
			preTestFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				requireAssignedAliasPairs(rollApp1.rollAppId, rollApp1.alias, t, ctx, dk)
				requireAssignedAliasPairs(rollApp2.rollAppId, rollApp2.alias, t, ctx, dk)
			},
			wantErr: false,
			afterTestFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				requireRollAppHasNoAlias(rollApp1.rollAppId, t, ctx, dk)

				// now 2 aliases are linked to roll app 2
				requireAliasLinkedToRollApp(rollApp1.alias, rollApp2.rollAppId, t, ctx, dk)
				requireAliasLinkedToRollApp(rollApp2.alias, rollApp2.rollAppId, t, ctx, dk)
			},
		},
		{
			name:         "pass - can move to RollApp with existing multiple Aliases",
			rollapps:     []rollapp{rollApp1, rollApp2},
			srcRollAppId: rollApp1.rollAppId,
			alias:        rollApp1.alias,
			dstRollAppId: rollApp2.rollAppId,
			preTestFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				requireAssignedAliasPairs(rollApp1.rollAppId, rollApp1.alias, t, ctx, dk)
				requireAssignedAliasPairs(rollApp2.rollAppId, rollApp2.alias, t, ctx, dk)

				require.NoError(t, dk.SetAliasForRollAppId(ctx, rollApp2.rollAppId, "new"))
			},
			wantErr: false,
			afterTestFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				requireRollAppHasNoAlias(rollApp1.rollAppId, t, ctx, dk)
				// now 3 aliases are linked to roll app 2
				requireAliasLinkedToRollApp(rollApp1.alias, rollApp2.rollAppId, t, ctx, dk)
				requireAliasLinkedToRollApp(rollApp2.alias, rollApp2.rollAppId, t, ctx, dk)
				requireAliasLinkedToRollApp("new", rollApp2.rollAppId, t, ctx, dk)
			},
		},
		{
			name:         "pass - can move to RollApp without alias",
			rollapps:     []rollapp{rollApp1, rollApp3WithoutAlias},
			srcRollAppId: rollApp1.rollAppId,
			alias:        rollApp1.alias,
			dstRollAppId: rollApp3WithoutAlias.rollAppId,
			preTestFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				requireAssignedAliasPairs(rollApp1.rollAppId, rollApp1.alias, t, ctx, dk)
				requireRollAppHasNoAlias(rollApp3WithoutAlias.rollAppId, t, ctx, dk)
			},
			wantErr: false,
			afterTestFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				requireRollAppHasNoAlias(rollApp1.rollAppId, t, ctx, dk)
				requireAssignedAliasPairs(rollApp3WithoutAlias.rollAppId, rollApp1.alias, t, ctx, dk)
			},
		},
		{
			name:         "fail - source RollApp has no alias linked",
			rollapps:     []rollapp{rollApp3WithoutAlias, rollApp4WithoutAlias},
			srcRollAppId: rollApp3WithoutAlias.rollAppId,
			alias:        "alias",
			dstRollAppId: rollApp4WithoutAlias.rollAppId,
			preTestFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				requireRollAppHasNoAlias(rollApp3WithoutAlias.rollAppId, t, ctx, dk)
				requireRollAppHasNoAlias(rollApp4WithoutAlias.rollAppId, t, ctx, dk)
			},
			wantErr:         true,
			wantErrContains: "alias not found",
			afterTestFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				requireRollAppHasNoAlias(rollApp3WithoutAlias.rollAppId, t, ctx, dk)
				requireRollAppHasNoAlias(rollApp4WithoutAlias.rollAppId, t, ctx, dk)
			},
		},
		{
			name:         "fail - source RollApp has no alias linked, move alias of another",
			rollapps:     []rollapp{rollApp1, rollApp3WithoutAlias, rollApp4WithoutAlias},
			srcRollAppId: rollApp3WithoutAlias.rollAppId,
			alias:        rollApp1.alias,
			dstRollAppId: rollApp4WithoutAlias.rollAppId,
			preTestFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				requireAssignedAliasPairs(rollApp1.rollAppId, rollApp1.alias, t, ctx, dk)
				requireRollAppHasNoAlias(rollApp3WithoutAlias.rollAppId, t, ctx, dk)
				requireRollAppHasNoAlias(rollApp4WithoutAlias.rollAppId, t, ctx, dk)
			},
			wantErr:         true,
			wantErrContains: "permission denied",
			afterTestFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				requireAssignedAliasPairs(rollApp1.rollAppId, rollApp1.alias, t, ctx, dk)
				requireRollAppHasNoAlias(rollApp3WithoutAlias.rollAppId, t, ctx, dk)
				requireRollAppHasNoAlias(rollApp4WithoutAlias.rollAppId, t, ctx, dk)
			},
		},
		{
			name:         "fail - move alias in-used by another RollApp",
			rollapps:     []rollapp{rollApp1, rollApp2, rollApp3WithoutAlias},
			srcRollAppId: rollApp1.rollAppId,
			alias:        rollApp2.alias,
			dstRollAppId: rollApp3WithoutAlias.rollAppId,
			preTestFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				requireAssignedAliasPairs(rollApp2.rollAppId, rollApp2.alias, t, ctx, dk)
				requireRollAppHasNoAlias(rollApp3WithoutAlias.rollAppId, t, ctx, dk)
			},
			wantErr:         true,
			wantErrContains: "permission denied",
			afterTestFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				requireAssignedAliasPairs(rollApp2.rollAppId, rollApp2.alias, t, ctx, dk)
				requireRollAppHasNoAlias(rollApp3WithoutAlias.rollAppId, t, ctx, dk)
			},
		},
		{
			name:         "fail - source RollApp ID is malformed",
			rollapps:     []rollapp{rollApp3WithoutAlias},
			srcRollAppId: "@bad",
			alias:        "alias",
			dstRollAppId: rollApp3WithoutAlias.rollAppId,
			preTestFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				requireAliasNotInUse("alias", t, ctx, dk)
				requireRollAppHasNoAlias(rollApp3WithoutAlias.rollAppId, t, ctx, dk)
			},
			wantErr:         true,
			wantErrContains: "source RollApp does not exists",
			afterTestFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				requireAliasNotInUse("alias", t, ctx, dk)
				requireRollAppHasNoAlias(rollApp3WithoutAlias.rollAppId, t, ctx, dk)
			},
		},
		{
			name:         "fail - bad alias",
			rollapps:     []rollapp{rollApp1, rollApp2},
			srcRollAppId: rollApp1.rollAppId,
			alias:        "@bad",
			dstRollAppId: rollApp2.rollAppId,
			preTestFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				requireAssignedAliasPairs(rollApp1.rollAppId, rollApp1.alias, t, ctx, dk)
			},
			wantErr:         true,
			wantErrContains: "invalid alias",
			afterTestFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				requireAssignedAliasPairs(rollApp1.rollAppId, rollApp1.alias, t, ctx, dk)
			},
		},
		{
			name:         "fail - destination RollApp ID is malformed",
			rollapps:     []rollapp{rollApp1},
			srcRollAppId: rollApp1.rollAppId,
			alias:        rollApp1.alias,
			dstRollAppId: "@bad",
			preTestFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				requireAssignedAliasPairs(rollApp1.rollAppId, rollApp1.alias, t, ctx, dk)
			},
			wantErr:         true,
			wantErrContains: "destination RollApp does not exists",
			afterTestFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				requireAssignedAliasPairs(rollApp1.rollAppId, rollApp1.alias, t, ctx, dk)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dk, _, rk, ctx := testkeeper.DymNSKeeper(t)

			for _, ra := range tt.rollapps {
				registerRollApp(t, ctx, rk, dk, ra.rollAppId, "", ra.alias)
			}

			if tt.preTestFunc != nil {
				tt.preTestFunc(t, ctx, dk, rk)
			}

			err := dk.MoveAliasToRollAppId(ctx, tt.srcRollAppId, tt.alias, tt.dstRollAppId)

			defer func() {
				if t.Failed() {
					return
				}
				if tt.afterTestFunc != nil {
					tt.afterTestFunc(t, ctx, dk, rk)
				}
			}()

			if tt.wantErr {
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test case")
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
				return
			}

			require.NoError(t, err)
		})
	}
}

func requireAssignedAliasPairs(rollAppId, alias string, t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper) {
	gotAlias, found := dk.GetAliasByRollAppId(ctx, rollAppId)
	require.True(t, found)
	require.Equal(t, alias, gotAlias)

	gotRollAppId, found := dk.GetRollAppIdByAlias(ctx, alias)
	require.True(t, found)
	require.Equal(t, rollAppId, gotRollAppId)
}

func requireAliasLinkedToRollApp(alias, rollAppId string, t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper) {
	gotRollAppId, found := dk.GetRollAppIdByAlias(ctx, alias)
	require.True(t, found)
	require.Equal(t, rollAppId, gotRollAppId)
}

func requireRollAppHasNoAlias(rollAppId string, t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper) {
	gotAlias, found := dk.GetAliasByRollAppId(ctx, rollAppId)
	require.False(t, found)
	require.Empty(t, gotAlias)
}

func requireAliasNotInUse(alias string, t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper) {
	gotRollAppId, found := dk.GetRollAppIdByAlias(ctx, alias)
	require.False(t, found)
	require.Empty(t, gotRollAppId)
}

func (s *KeeperTestSuite) TestKeeper_IsAliasPresentsInParamsAsAliasOrChainId() {
	tests := []struct {
		name       string
		preRunFunc func(s *KeeperTestSuite)
		alias      string
		want       bool
	}{
		{
			name: "alias mapped in params",
			preRunFunc: func(s *KeeperTestSuite) {
				s.updateModuleParams(func(params dymnstypes.Params) dymnstypes.Params {
					params.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
						{
							ChainId: "dymension_100-1",
							Aliases: []string{"dym"},
						},
					}
					return params
				})
			},
			alias: "dym",
			want:  true,
		},
		{
			name: "alias as chain-id in params",
			preRunFunc: func(s *KeeperTestSuite) {
				s.updateModuleParams(func(params dymnstypes.Params) dymnstypes.Params {
					params.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
						{
							ChainId: "dymension",
							Aliases: []string{"dym"},
						},
					}
					return params
				})
			},
			alias: "dymension",
			want:  true,
		},
		{
			name: "alias not in params",
			preRunFunc: func(s *KeeperTestSuite) {
				s.updateModuleParams(func(params dymnstypes.Params) dymnstypes.Params {
					params.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
						{
							ChainId: "dymension",
							Aliases: []string{"dym"},
						},
					}
					return params
				})
			},
			alias: "alias",
			want:  false,
		},
		{
			name: "alias is used by RollApp",
			preRunFunc: func(s *KeeperTestSuite) {
				rollApp := newRollApp("rollapp_1-1").WithAlias("alias")
				s.persistRollApp(*rollApp)

				s.requireRollApp(rollApp.rollAppId).HasAlias("alias")
			},
			alias: "alias",
			want:  false,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.preRunFunc != nil {
				tt.preRunFunc(s)
			}

			got := s.dymNsKeeper.IsAliasPresentsInParamsAsAliasOrChainId(s.ctx, tt.alias)
			s.Require().Equal(tt.want, got)
		})
	}
}
