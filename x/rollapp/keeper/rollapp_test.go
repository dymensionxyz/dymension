package keeper_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/testutil/nullify"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func TestRollappGet(t *testing.T) {
	k, ctx := keepertest.RollappKeeper(t)
	items, _ := createNRollapp(k, ctx, 10)
	for _, item := range items {
		item := item
		rst, found := k.GetRollapp(ctx,
			item.RollappId,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}

func TestRollappRemove(t *testing.T) {
	k, ctx := keepertest.RollappKeeper(t)
	items, _ := createNRollapp(k, ctx, 10)
	for _, item := range items {
		k.RemoveRollapp(ctx,
			item.RollappId,
		)
		_, found := k.GetRollapp(ctx,
			item.RollappId,
		)
		require.False(t, found)
	}
}

func TestRollappGetAll(t *testing.T) {
	k, ctx := keepertest.RollappKeeper(t)
	items, _ := createNRollapp(k, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(k.GetAllRollapps(ctx)),
	)
}

func TestKeeper_FindRollappByName(t *testing.T) {
	const rollappID = "rollapp1_1234-1"

	k, ctx := keepertest.RollappKeeper(t)
	k.SetRollapp(ctx, types.Rollapp{
		RollappId: rollappID,
	})

	tests := []struct {
		name             string
		queryRollappName string
		wantFound        bool
	}{
		{
			name:             "rollapp found: same name",
			queryRollappName: "rollapp1",
			wantFound:        true,
		}, {
			name:             "rollapp not found: different name",
			queryRollappName: "rollapp2",
			wantFound:        false,
		}, {
			name:             "rollapp not found: partial name match 1",
			queryRollappName: "rollapp",
			wantFound:        false,
		}, {
			name:             "rollapp not found: partial name match 2",
			queryRollappName: "rollapp12",
			wantFound:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, gotFound := k.FindRollappByName(ctx, tt.queryRollappName)
			require.Equal(t, tt.wantFound, gotFound)
		})
	}
}
