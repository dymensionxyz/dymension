package keeper

import (
	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/internal/collcompat"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

var LPsByRollAppDenomPrefix = collections.NewPrefix(0)
var LPsByIDPrefix = collections.NewPrefix(1)
var LPsNextIDPrefix = collections.NewPrefix(2)

type LPs struct {
	// <rollapp,denom,id>
	byRollAppDenom collections.KeySet[collections.Triple[string, string, uint64]]
	// id -> lp
	byID   collections.Map[uint64, types.OnDemandLiquidity]
	nextID collections.Sequence
}

type OnDemandLiquidity struct {
	Id         uint64
	Spent      math.Int
	FundsAddr  string
	Rollapp    string
	Denom      string
	MaxPrice   math.Int
	MinFee     math.Int
	SpendLimit math.Int
}

func makeLPsStore(sb *collections.SchemaBuilder, cdc codec.BinaryCodec) LPs {
	return LPs{

		byRollAppDenom: collections.NewKeySet[collections.Triple[string, string, uint64]](
			sb, LPsByRollAppDenomPrefix, "byRollAppDenom",
			collections.TripleKeyCodec[string, string, uint64](
				collections.StringKey,
				collections.StringKey,
				collections.Uint64Key,
			)),
		byID: collections.NewMap[uint64, types.OnDemandLiquidity](
			sb, LPsByIDPrefix, "byID",
			collections.Uint64Key, collcompat.ProtoValue[types.OnDemandLiquidity](cdc),
		),
		nextID: collections.NewSequence(sb, LPsNextIDPrefix, "nextID"),
	}
}

func (s LPs) UpsertLP(ctx sdk.Context, lp *types.OnDemandLiquidity) (uint64, error) {
	id, err := s.nextID.Next(ctx)
	if err != nil {
		return 0, err
	}
	lp.Id = id
	err = s.byID.Set(ctx, id, *lp)
	if err != nil {
		return 0, err
	}
	err = s.byRollAppDenom.Set(ctx, collections.Join3(lp.Rollapp, lp.Denom, id))
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s LPs) FindLP(ctx sdk.Context, k Keeper, o *types.DemandOrder) (*types.OnDemandLiquidity, error) {

	rol := o.RollappId
	denom := o.Denom()
	rng := collections.NewSuperPrefixedTripleRange[string, string, uint64](rol, denom)
	iter, err := s.byRollAppDenom.Iterate(ctx, rng)
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		key, err := iter.Key()
		if err != nil {
			return nil, err
		}
		id := key.K3()

		lp, err := s.byID.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		if lp.Accepts(o) {
			// TODO: just direct fulfill here
			return &lp, nil
		}
	}
	return nil, nil
}
