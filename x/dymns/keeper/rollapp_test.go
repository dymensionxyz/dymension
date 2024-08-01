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
