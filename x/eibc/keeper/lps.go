package keeper

import (
	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

var LPsByRollAppDenomPrefix = collections.NewPrefix(0)
var LPsByIDPrefix = collections.NewPrefix(1)
var LPsNextIDPrefix = collections.NewPrefix(2)

type LPs struct {
	// <rollapp,denom,id>
	byRollAppDenom collections.KeySet[collections.Triple[string, string, uint64]]
	// id -> lp
	byID   collections.Map[uint64, types.OnDemandLPRecord]
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
		byID: collections.NewMap[uint64, types.OnDemandLPRecord](
			sb, LPsByIDPrefix, "byID",
			collections.Uint64Key, codec.CollValue[types.OnDemandLPRecord](cdc),
		),
		nextID: collections.NewSequence(sb, LPsNextIDPrefix, "nextID"),
	}
}

func (s LPs) UpsertLP(ctx sdk.Context, lp *types.OnDemandLP) (uint64, error) {
	id, err := s.nextID.Next(ctx)
	if err != nil {
		return 0, err
	}
	record := types.OnDemandLPRecord{
		Id:    id,
		Lp:    lp,
		Spent: math.ZeroInt(),
	}
	err = s.byID.Set(ctx, id, record)
	if err != nil {
		return 0, err
	}
	err = s.byRollAppDenom.Set(ctx, collections.Join3(lp.Rollapp, lp.Denom, id))
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s LPs) FindLP(ctx sdk.Context, k Keeper, o *types.DemandOrder) (*types.OnDemandLPRecord, error) {

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

		lpr, err := s.byID.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		if lpr.Accepts(o) {
			// TODO: just direct fulfill here
			return &lpr, nil
		}
	}
	return nil, nil
}

func (k Keeper) FindOnDemandLP(ctx sdk.Context, order string) error {

}

func (k Keeper) CreateLP(ctx sdk.Context, lp *types.OnDemandLP) (uint64, error) {

}

func (k Keeper) DeleteLP(ctx sdk.Context, id uint64) error {

}
