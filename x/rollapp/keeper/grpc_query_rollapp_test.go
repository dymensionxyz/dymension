package keeper_test

import (
	"flag"
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
	flag.Set("rapid.checks", "50")
	flag.Set("rapid.steps", "50")
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
			"query paginate key": func(t *rapid.T) {
				var key []byte
				unseen := maps.Clone(m)
				reverse := rapid.Bool().Draw(t, "reverse")
				for {
					limit := rapid.Uint64().Draw(t, "limit")
					q := &types.QueryAllRollappRequest{Pagination: &query.PageRequest{
						Key:     key,
						Limit:   limit,
						Reverse: reverse,
					}}
					res, err := k.RollappAll(sdk.WrapSDKContext(ctx), q)
					require.NoError(t, err)
					for _, ra := range res.Rollapp {
						_, ok := unseen[ra.Rollapp.RollappId]
						require.True(t, ok)
						delete(unseen, ra.Rollapp.RollappId)
					}
					key = res.Pagination.NextKey
					if key == nil {
						break
					}
				}
				require.Empty(t, unseen)
			},
			"query paginate offset": func(r *rapid.T) {
				unseen := maps.Clone(m)
				offset := uint64(0)
				reverse := rapid.Bool().Draw(r, "reverse")
				for {
					limit := rapid.Uint64().Draw(r, "limit")
					q := &types.QueryAllRollappRequest{Pagination: &query.PageRequest{
						Offset:  offset,
						Limit:   limit,
						Reverse: reverse,
					}}
					res, err := k.RollappAll(sdk.WrapSDKContext(ctx), q)
					require.NoError(t, err)
					for _, ra := range res.Rollapp {
						_, ok := unseen[ra.Rollapp.RollappId]
						require.True(t, ok, "id", ra.Rollapp.RollappId)
						delete(unseen, ra.Rollapp.RollappId)
					}
					if uint64(len(res.Rollapp)) < limit {
						break
					}
					offset += uint64(len(res.Rollapp))
				}
				require.Empty(t, unseen)
			},
		})
	})
}
