package keeper_test

import (
	"flag"
	"math"
	"testing"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/urand"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"pgregory.net/rapid"

	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func TestQuery(t *testing.T) {
	_ = flag.Set("rapid.checks", "50")
	_ = flag.Set("rapid.steps", "50")
	rapid.Check(t, func(r *rapid.T) {
		k, ctx := keepertest.RollappKeeper(t)
		ids := rapid.SampledFrom([]string{
			urand.RollappID(),
			urand.RollappID(),
			urand.RollappID(),
			urand.RollappID(),
			urand.RollappID(),
		})
		m := map[string]types.Rollapp{}
		r.Repeat(map[string]func(*rapid.T){
			"": func(t *rapid.T) {
			},
			"put": func(t *rapid.T) {
				ra := types.Rollapp{RollappId: ids.Draw(t, "id")}
				m[ra.RollappId] = ra
				k.SetRollapp(ctx, ra)
			},
			"del": func(t *rapid.T) {
				id := ids.Draw(t, "id")
				delete(m, id)
				k.RemoveRollapp(ctx, id)
			},
			"get one": func(t *rapid.T) {
				id := ids.Draw(t, "id")
				ra, ok := k.GetRollapp(ctx, id)
				_, okM := m[ra.RollappId]
				require.Equal(t, okM, ok)
				if okM {
					require.Equal(t, ra.RollappId, id)
				}
			},
			"get all": func(t *rapid.T) {
				got := k.GetAllRollapps(ctx)
				for _, ra := range got {
					_, ok := m[ra.RollappId]
					require.True(t, ok)
				}
				for _, ra := range m {
					require.True(t, slices.ContainsFunc(got, func(raGot types.Rollapp) bool {
						return raGot.RollappId == ra.RollappId
					}))
				}
			},
			"get one by name": func(t *rapid.T) {
				id := ids.Draw(t, "id")
				cid, _ := types.NewChainID(id)
				ra, ok := k.GetRollappByName(ctx, cid.GetName())
				_, okM := m[id]
				require.Equal(t, okM, ok)
				if okM {
					require.Equal(t, id, ra.RollappId)
				}
			},
			"query one": func(t *rapid.T) {
				id := ids.Draw(t, "id")
				res, err := k.Rollapp(sdk.WrapSDKContext(ctx), &types.QueryGetRollappRequest{RollappId: id})
				_, okM := m[id]
				if !okM {
					require.True(t, errorsmod.IsOf(err, gerrc.ErrNotFound))
				} else {
					require.Equal(t, id, res.Rollapp.RollappId)
				}
			},
			"query many": func(t *rapid.T) {
				var key []byte
				unseen := maps.Clone(m)
				q := &types.QueryAllRollappRequest{Pagination: &query.PageRequest{
					Key:   key,
					Limit: math.MaxUint64, // no need to rest pagination mechanism
				}}
				res, err := k.RollappAll(sdk.WrapSDKContext(ctx), q)
				require.NoError(t, err)
				for _, ra := range res.Rollapp {
					_, ok := unseen[ra.Rollapp.RollappId]
					require.True(t, ok)
					delete(unseen, ra.Rollapp.RollappId)
				}
				require.Empty(t, unseen)
			},
		})
	})
}
