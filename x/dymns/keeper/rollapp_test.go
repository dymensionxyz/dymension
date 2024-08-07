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
		Owner:     testAddr(1).bech32(),
	})

	rk.SetRollapp(ctx, rollapptypes.Rollapp{
		RollappId: "rolling_2-2",
		Owner:     testAddr(2).bech32(),
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

func TestKeeper_IsRollAppCreator(t *testing.T) {
	acc1 := testAddr(1)
	acc2 := testAddr(2)

	tests := []struct {
		name      string
		rollApp   *rollapptypes.Rollapp
		rollAppId string
		account   string
		want      bool
	}{
		{
			name: "pass - is creator",
			rollApp: &rollapptypes.Rollapp{
				RollappId: "rollapp_1-1",
				Owner:     acc1.bech32(),
			},
			rollAppId: "rollapp_1-1",
			account:   acc1.bech32(),
			want:      true,
		},
		{
			name: "fail - rollapp does not exists",
			rollApp: &rollapptypes.Rollapp{
				RollappId: "rollapp_1-1",
				Owner:     acc1.bech32(),
			},
			rollAppId: "nah_2-2",
			account:   acc1.bech32(),
			want:      false,
		},
		{
			name: "fail - is NOT creator",
			rollApp: &rollapptypes.Rollapp{
				RollappId: "rollapp_1-1",
				Owner:     acc1.bech32(),
			},
			rollAppId: "rollapp_1-1",
			account:   acc2.bech32(),
			want:      false,
		},
		{
			name: "fail - creator but in different bech32 format is not accepted",
			rollApp: &rollapptypes.Rollapp{
				RollappId: "rollapp_1-1",
				Owner:     acc1.bech32(),
			},
			rollAppId: "rollapp_1-1",
			account:   acc1.bech32C("nim"),
			want:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dk, _, rk, ctx := testkeeper.DymNSKeeper(t)

			if tt.rollApp != nil {
				rk.SetRollapp(ctx, *tt.rollApp)
			}

			got := dk.IsRollAppCreator(ctx, tt.rollAppId, tt.account)
			require.Equal(t, tt.want, got)
		})
	}

	t.Run("pass - can detect among multiple RollApps of same owned", func(t *testing.T) {
		dk, _, rk, ctx := testkeeper.DymNSKeeper(t)

		rollAppABy1 := rollapptypes.Rollapp{
			RollappId: "rollapp_1-1",
			Owner:     acc1.bech32(),
		}
		rollAppBBy1 := rollapptypes.Rollapp{
			RollappId: "rollapp_2-2",
			Owner:     acc1.bech32(),
		}
		rollAppCBy2 := rollapptypes.Rollapp{
			RollappId: "rollapp_3-3",
			Owner:     acc2.bech32(),
		}
		rollAppDBy2 := rollapptypes.Rollapp{
			RollappId: "rollapp_4-4",
			Owner:     acc2.bech32(),
		}

		rk.SetRollapp(ctx, rollAppABy1)
		rk.SetRollapp(ctx, rollAppBBy1)
		rk.SetRollapp(ctx, rollAppCBy2)
		rk.SetRollapp(ctx, rollAppDBy2)

		require.True(t, dk.IsRollAppCreator(ctx, rollAppABy1.RollappId, acc1.bech32()))
		require.True(t, dk.IsRollAppCreator(ctx, rollAppBBy1.RollappId, acc1.bech32()))
		require.True(t, dk.IsRollAppCreator(ctx, rollAppCBy2.RollappId, acc2.bech32()))
		require.True(t, dk.IsRollAppCreator(ctx, rollAppDBy2.RollappId, acc2.bech32()))

		require.False(t, dk.IsRollAppCreator(ctx, rollAppABy1.RollappId, acc2.bech32()))
		require.False(t, dk.IsRollAppCreator(ctx, rollAppBBy1.RollappId, acc2.bech32()))
		require.False(t, dk.IsRollAppCreator(ctx, rollAppCBy2.RollappId, acc1.bech32()))
		require.False(t, dk.IsRollAppCreator(ctx, rollAppDBy2.RollappId, acc1.bech32()))
	})
}

func TestKeeper_GetRollAppBech32Prefix(t *testing.T) {
	rollApp1 := rollapptypes.Rollapp{
		RollappId:    "rollapp_1-1",
		Owner:        testAddr(0).bech32(),
		Bech32Prefix: "one",
	}
	rollApp2 := rollapptypes.Rollapp{
		RollappId:    "rolling_2-2",
		Owner:        testAddr(0).bech32(),
		Bech32Prefix: "two",
	}
	rollApp3NonExists := rollapptypes.Rollapp{
		RollappId:    "nah_3-3",
		Owner:        testAddr(0).bech32(),
		Bech32Prefix: "nah",
	}

	dk, _, rk, ctx := testkeeper.DymNSKeeper(t)
	rk.SetRollapp(ctx, rollApp1)
	rk.SetRollapp(ctx, rollApp2)

	bech32, found := dk.GetRollAppBech32Prefix(ctx, rollApp1.RollappId)
	require.True(t, found)
	require.Equal(t, "one", bech32)

	bech32, found = dk.GetRollAppBech32Prefix(ctx, rollApp2.RollappId)
	require.True(t, found)
	require.Equal(t, "two", bech32)

	bech32, found = dk.GetRollAppBech32Prefix(ctx, rollApp3NonExists.RollappId)
	require.False(t, found)
	require.Empty(t, bech32)
}
