package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/stretchr/testify/require"
)

func TestKeeper_GetSetAliasForRollAppId(t *testing.T) {
	type rollApp struct {
		id    string
		alias string
	}

	rollApp1 := rollApp{
		id:    "rollapp_1-1",
		alias: "al1",
	}

	rollApp2 := rollApp{
		id:    "rolling_2-2",
		alias: "al2",
	}

	rollApp3NotExists := rollApp{
		id:    "nah_2-2",
		alias: "al3",
	}

	dk, _, rk, ctx := testkeeper.DymNSKeeper(t)

	for i, ra := range []rollApp{rollApp1, rollApp2} {
		rk.SetRollapp(ctx, rollapptypes.Rollapp{
			RollappId: ra.id,
			Owner:     testAddr(uint64(i)).bech32(),
		})
	}

	t.Run("set - can set", func(t *testing.T) {
		require.True(t, dk.IsRollAppId(ctx, rollApp1.id), "must be a RollApp, just not set alias")

		err := dk.SetAliasForRollAppId(ctx, rollApp1.id, rollApp1.alias)
		require.NoError(t, err)

		alias, found := dk.GetAliasByRollAppId(ctx, rollApp1.id)
		require.Equal(t, rollApp1.alias, alias)
		require.True(t, found)

		rollAppId, found := dk.GetRollAppIdByAlias(ctx, rollApp1.alias)
		require.Equal(t, rollApp1.id, rollAppId)
		require.True(t, found)
	})

	t.Run("set - can NOT set if alias is being in-used by another RollApp", func(t *testing.T) {
		rollAppId, found := dk.GetRollAppIdByAlias(ctx, rollApp1.alias)
		require.Equal(t, rollApp1.id, rollAppId)
		require.True(t, found)

		err := dk.SetAliasForRollAppId(ctx, rollApp2.id, rollApp1.alias)
		require.Error(t, err)
		require.Contains(t, err.Error(), "alias currently being in used by:")
	})

	t.Run("set - reject chain-id", func(t *testing.T) {
		err := dk.SetAliasForRollAppId(ctx, "bad@", "alias")
		require.Error(t, err)
	})

	t.Run("set - reject bad alias", func(t *testing.T) {
		require.True(t, dk.IsRollAppId(ctx, rollApp2.id), "must be a RollApp")

		err := dk.SetAliasForRollAppId(ctx, rollApp2.id, "")
		require.Error(t, err)
		require.Contains(t, err.Error(), "alias can not be empty")

		err = dk.SetAliasForRollAppId(ctx, rollApp2.id, "@")
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid alias")
	})

	t.Run("get - of existing RollApp but no alias set", func(t *testing.T) {
		require.True(t, dk.IsRollAppId(ctx, rollApp2.id), "must be a RollApp, just not set alias")

		alias, found := dk.GetAliasByRollAppId(ctx, rollApp2.id)
		require.Empty(t, alias)
		require.False(t, found)

		rollAppId, found := dk.GetRollAppIdByAlias(ctx, rollApp2.alias)
		require.Empty(t, rollAppId)
		require.False(t, found)
	})

	t.Run("set - non-exists RollApp returns error", func(t *testing.T) {
		require.False(t, dk.IsRollAppId(ctx, rollApp3NotExists.id))

		err := dk.SetAliasForRollAppId(ctx, rollApp3NotExists.id, rollApp3NotExists.alias)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not a RollApp")
	})

	t.Run("get - non-exists RollApp returns empty", func(t *testing.T) {
		require.False(t, dk.IsRollAppId(ctx, rollApp3NotExists.id))

		alias, found := dk.GetAliasByRollAppId(ctx, rollApp3NotExists.id)
		require.Empty(t, alias)
		require.False(t, found)

		rollAppId, found := dk.GetRollAppIdByAlias(ctx, rollApp3NotExists.alias)
		require.Empty(t, rollAppId)
		require.False(t, found)
	})

	t.Run("set/get - can set multiple alias to a single Roll-App", func(t *testing.T) {
		dk, _, rk, ctx := testkeeper.DymNSKeeper(t)

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
			rk.SetRollapp(ctx, rollapptypes.Rollapp{
				RollappId: tc.rollAppId,
				Owner:     testAddr(0).bech32(),
			})
		}

		for _, tc := range testcases {
			for _, alias := range tc.aliases {
				err := dk.SetAliasForRollAppId(ctx, tc.rollAppId, alias)
				require.NoError(t, err)
			}
		}

		for _, tc := range testcases {
			for _, alias := range tc.aliases {
				rollAppId, found := dk.GetRollAppIdByAlias(ctx, alias)
				require.Equal(t, tc.rollAppId, rollAppId)
				require.True(t, found)
			}

			alias, found := dk.GetAliasByRollAppId(ctx, tc.rollAppId)
			require.True(t, found)
			require.Contains(t, tc.aliases, alias)
			require.Equal(t, alias, tc.aliases[0], "should returns the first one added")
		}
	})
}

func TestKeeper_RemoveAliasFromRollAppId(t *testing.T) {
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
	rollApp3 := rollapp{
		rollAppId: "rollapp_3-3",
		alias:     "al3",
	}
	rollApp4NoAlias := rollapp{
		rollAppId: "noa_4-4",
		alias:     "",
	}
	rollApp5NotExists := rollapp{
		rollAppId: "nah_5-5",
		alias:     "al5",
	}
	const aliasOne = "one"
	const aliasTwo = "two"
	const unusedAlias = "unused"

	tests := []struct {
		name            string
		addRollApps     []rollapp
		preRunFunc      func(*testing.T, sdk.Context, dymnskeeper.Keeper, rollappkeeper.Keeper)
		inputRollAppId  string
		inputAlias      string
		wantErr         bool
		wantErrContains string
		afterTestFunc   func(*testing.T, sdk.Context, dymnskeeper.Keeper, rollappkeeper.Keeper)
	}{
		{
			name:        "pass - can remove",
			addRollApps: []rollapp{rollApp1},
			preRunFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				requireAssignedAliasPairs(rollApp1.rollAppId, rollApp1.alias, t, ctx, dk)
			},
			inputRollAppId: rollApp1.rollAppId,
			inputAlias:     rollApp1.alias,
			wantErr:        false,
			afterTestFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				requireRollAppHasNoAlias(rollApp1.rollAppId, t, ctx, dk)
				requireAliasNotInUse(rollApp1.alias, t, ctx, dk)
			},
		},
		{
			name:        "pass - can remove among multiple records",
			addRollApps: []rollapp{rollApp1, rollApp2, rollApp3},
			preRunFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				requireAssignedAliasPairs(rollApp1.rollAppId, rollApp1.alias, t, ctx, dk)
				requireAssignedAliasPairs(rollApp2.rollAppId, rollApp2.alias, t, ctx, dk)
				requireAssignedAliasPairs(rollApp3.rollAppId, rollApp3.alias, t, ctx, dk)
			},
			inputRollAppId: rollApp2.rollAppId,
			inputAlias:     rollApp2.alias,
			wantErr:        false,
			afterTestFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				requireRollAppHasNoAlias(rollApp2.rollAppId, t, ctx, dk)
				requireAliasNotInUse(rollApp2.alias, t, ctx, dk)

				// other records remain unchanged
				requireAssignedAliasPairs(rollApp1.rollAppId, rollApp1.alias, t, ctx, dk)
				requireAssignedAliasPairs(rollApp3.rollAppId, rollApp3.alias, t, ctx, dk)
			},
		},
		{
			name:        "fail - reject if input RollApp ID is empty",
			addRollApps: []rollapp{rollApp1},
			preRunFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				requireAssignedAliasPairs(rollApp1.rollAppId, rollApp1.alias, t, ctx, dk)
			},
			inputRollAppId:  "",
			inputAlias:      rollApp1.alias,
			wantErr:         true,
			wantErrContains: "not a RollApp",
			afterTestFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				// record remains unchanged
				requireAssignedAliasPairs(rollApp1.rollAppId, rollApp1.alias, t, ctx, dk)
			},
		},
		{
			name:        "fail - reject if input RollApp ID is not exists",
			addRollApps: []rollapp{rollApp1},
			preRunFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				require.False(t, dk.IsRollAppId(ctx, rollApp5NotExists.rollAppId))

				requireAssignedAliasPairs(rollApp1.rollAppId, rollApp1.alias, t, ctx, dk)
			},
			inputRollAppId:  rollApp5NotExists.rollAppId,
			inputAlias:      rollApp5NotExists.alias,
			wantErr:         true,
			wantErrContains: "not a RollApp",
			afterTestFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				// other records remain unchanged
				requireAssignedAliasPairs(rollApp1.rollAppId, rollApp1.alias, t, ctx, dk)
			},
		},
		{
			name:        "fail - reject if input Alias is empty",
			addRollApps: []rollapp{rollApp1},
			preRunFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				requireAssignedAliasPairs(rollApp1.rollAppId, rollApp1.alias, t, ctx, dk)
			},
			inputRollAppId:  rollApp1.rollAppId,
			inputAlias:      "",
			wantErr:         true,
			wantErrContains: "alias can not be empty",
			afterTestFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				// record remains unchanged
				requireAssignedAliasPairs(rollApp1.rollAppId, rollApp1.alias, t, ctx, dk)
			},
		},
		{
			name:        "fail - reject if input Alias is malformed",
			addRollApps: []rollapp{rollApp1},
			preRunFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				requireAssignedAliasPairs(rollApp1.rollAppId, rollApp1.alias, t, ctx, dk)
			},
			inputRollAppId:  rollApp1.rollAppId,
			inputAlias:      "@",
			wantErr:         true,
			wantErrContains: "invalid alias",
			afterTestFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				// record remains unchanged
				requireAssignedAliasPairs(rollApp1.rollAppId, rollApp1.alias, t, ctx, dk)
			},
		},
		{
			name:        "fail - reject if Roll App has no alias linked",
			addRollApps: []rollapp{rollApp4NoAlias},
			preRunFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				requireRollAppHasNoAlias(rollApp4NoAlias.rollAppId, t, ctx, dk)
			},
			inputRollAppId:  rollApp4NoAlias.rollAppId,
			inputAlias:      aliasOne,
			wantErr:         true,
			wantErrContains: "alias not found",
			afterTestFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				// record remains unchanged
				requireRollAppHasNoAlias(rollApp4NoAlias.rollAppId, t, ctx, dk)
			},
		},
		{
			name:        "fail - reject if Roll App has no alias linked and input alias linked to another Roll App",
			addRollApps: []rollapp{rollApp1, rollApp4NoAlias},
			preRunFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				requireAssignedAliasPairs(rollApp1.rollAppId, rollApp1.alias, t, ctx, dk)
				requireRollAppHasNoAlias(rollApp4NoAlias.rollAppId, t, ctx, dk)
			},
			inputRollAppId:  rollApp4NoAlias.rollAppId,
			inputAlias:      rollApp1.alias,
			wantErr:         true,
			wantErrContains: "alias currently being in used by:",
			afterTestFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				// records remain unchanged
				requireAssignedAliasPairs(rollApp1.rollAppId, rollApp1.alias, t, ctx, dk)
				requireRollAppHasNoAlias(rollApp4NoAlias.rollAppId, t, ctx, dk)
			},
		},
		{
			name:        "fail - reject if remove alias linked to another Roll App",
			addRollApps: []rollapp{rollApp1, rollApp2},
			preRunFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				requireAssignedAliasPairs(rollApp1.rollAppId, rollApp1.alias, t, ctx, dk)
				requireAssignedAliasPairs(rollApp2.rollAppId, rollApp2.alias, t, ctx, dk)
			},
			inputRollAppId:  rollApp1.rollAppId,
			inputAlias:      rollApp2.alias,
			wantErr:         true,
			wantErrContains: "alias currently being in used by",
			afterTestFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				// records remain unchanged
				requireAssignedAliasPairs(rollApp1.rollAppId, rollApp1.alias, t, ctx, dk)
				requireAssignedAliasPairs(rollApp2.rollAppId, rollApp2.alias, t, ctx, dk)
			},
		},
		{
			name:        "fail - reject if input alias does not link to any Roll App",
			addRollApps: []rollapp{rollApp1},
			preRunFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				requireAssignedAliasPairs(rollApp1.rollAppId, rollApp1.alias, t, ctx, dk)
			},
			inputRollAppId:  rollApp1.rollAppId,
			inputAlias:      unusedAlias,
			wantErr:         true,
			wantErrContains: "alias not found",
			afterTestFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				// records remain unchanged
				requireAssignedAliasPairs(rollApp1.rollAppId, rollApp1.alias, t, ctx, dk)
			},
		},
		{
			name:        "pass - remove alias correctly among multiple aliases linked to a Roll App",
			addRollApps: []rollapp{rollApp1},
			preRunFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				require.NoError(t, dk.SetAliasForRollAppId(ctx, rollApp1.rollAppId, aliasOne))
				require.NoError(t, dk.SetAliasForRollAppId(ctx, rollApp1.rollAppId, aliasTwo))

				requireAssignedAliasPairs(rollApp1.rollAppId, rollApp1.alias, t, ctx, dk)
				requireAliasLinkedToRollApp(aliasOne, rollApp1.rollAppId, t, ctx, dk)
				requireAliasLinkedToRollApp(aliasTwo, rollApp1.rollAppId, t, ctx, dk)
			},
			inputRollAppId: rollApp1.rollAppId,
			inputAlias:     aliasOne,
			wantErr:        false,
			afterTestFunc: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				requireAssignedAliasPairs(rollApp1.rollAppId, rollApp1.alias, t, ctx, dk)
				requireAliasNotInUse(aliasOne, t, ctx, dk)
				requireAliasLinkedToRollApp(aliasTwo, rollApp1.rollAppId, t, ctx, dk)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dk, _, rk, ctx := testkeeper.DymNSKeeper(t)

			for _, ra := range tt.addRollApps {
				registerRollApp(t, ctx, rk, dk, ra.rollAppId, "", ra.alias)
			}

			if tt.preRunFunc != nil {
				tt.preRunFunc(t, ctx, dk, rk)
			}

			err := dk.RemoveAliasFromRollAppId(ctx, tt.inputRollAppId, tt.inputAlias)

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
