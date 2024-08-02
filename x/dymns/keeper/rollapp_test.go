package keeper_test

import (
	"testing"

	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/stretchr/testify/require"
)

func TestKeeper_IsRollAppId(t *testing.T) {
	dk, _, rk, ctx := testkeeper.DymNSKeeper(t)

	rk.SetRollapp(ctx, rollapptypes.Rollapp{
		RollappId: "rollapp_1-1",
		Creator:   testAddr(1).bech32(),
	})

	rk.SetRollapp(ctx, rollapptypes.Rollapp{
		RollappId: "rolling_2-2",
		Creator:   testAddr(2).bech32(),
	})

	tests := []struct {
		rollAppId     string
		wantIsRollApp bool
	}{
		{
			rollAppId:     "rollapp_1-1",
			wantIsRollApp: true,
		},
		{
			rollAppId:     "rolling_2-2",
			wantIsRollApp: true,
		},
		{
			rollAppId:     "rollapp_1-11",
			wantIsRollApp: false,
		},
		{
			rollAppId:     "rollapp_11-1",
			wantIsRollApp: false,
		},
		{
			rollAppId:     "rollapp_11-11",
			wantIsRollApp: false,
		},
		{
			rollAppId:     "rollapp_1-2",
			wantIsRollApp: false,
		},
		{
			rollAppId:     "rollapp_2-1",
			wantIsRollApp: false,
		},
		{
			rollAppId:     "rolling_1-1",
			wantIsRollApp: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.rollAppId, func(t *testing.T) {
			gotIsRollApp := dk.IsRollAppId(ctx, tt.rollAppId)
			require.Equal(t, tt.wantIsRollApp, gotIsRollApp)
		})
	}
}

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
			Creator:   testAddr(uint64(i)).bech32(),
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
}
