package keeper

import (
	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/internal/collcompat"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

var LPsByRollAppDenomPrefix = collections.NewPrefix(0)
var LPsByIDPrefix = collections.NewPrefix(1)

//var LPsPrefix = collections.NewPrefix(0)
//var LPsIndexesAccNumberPrefix = collections.NewPrefix(1)

type LPs struct {
	//Accounts *collections.IndexedMap[sdk.AccAddress, authtypes.BaseAccount, LPsIndexes]

	/*
			TODO: where I'm at:
			Trying to come up with the data scheme that will satisfy all data patterns
			Ideas
				1. Keyset(rollapp,denom,id)
					Allows quickly finding viable set of ids
				2. Map(id -> full struct)
					Allows looks up which we will inevitably want for updates/debugging
				3. Expiry -> id
					Allows efficient expiration

				Thus
					find(order):
						scan keyset for rollapp,denom to get set of ids
						read full structs by id, iterate through to find match
		                update the spent total
						write back in map
						(*can optionally constant optimize by having price, min fee, proof height as keyset value)
		            expire
						scan the expiry range and delete
					Revoke/update
						look up full by id then lookup in keyset/expiry

			After simplications (omitting fields for mvp):
				Exactly the same


	*/
	//LPs collections.IndexedMap[
	//collections.Triple[rollapp,denom,]
	//byRollAppDenom collections.Map[collections.Triple[string,string,string], uint64]
	// <rollapp,denom,id>
	byRollAppDenom collections.KeySet[collections.Triple[string, string, uint64]]
	// id -> lp
	byID collections.Map[uint64, types.OnDemandLiquidity]
	//M collections.Map[uint64, uint64]
}

func makeLPsStore(sb *collections.SchemaBuilder, cdc codec.BinaryCodec) LPs {
	return LPs{
		//Accounts: collections.NewIndexedMap(
		//	sb, LPsPrefix, "accounts",
		//	sdk.AccAddressKey, codec.CollValue[authtypes.BaseAccount](cdc),
		//	NewLPsIndexes(sb),
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
	}
}

func (s LPs) findLP(ctx sdk.Context, k *Keeper, o *types.DemandOrder) (*types.OnDemandLiquidity, error) {

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

//type LPsIndexes struct {
//	Number *indexes.Unique[uint64, sdk.AccAddress, authtypes.BaseAccount]
//}
//
//func NewLPsIndexes(sb *collections.SchemaBuilder) LPsIndexes {
//	return LPsIndexes{
//		Number: indexes.NewUnique(
//			sb, LPsIndexesAccNumberPrefix, "accounts_by_number",
//			collections.Uint64Key, sdk.AccAddressKey,
//			func(_ sdk.AccAddress, v authtypes.BaseAccount) (uint64, error) {
//				return v.AccountNumber, nil
//			},
//		),
//	}
//}
//
//func (a LPsIndexes) IndexesList() []collections.Index[sdk.AccAddress, authtypes.BaseAccount] {
//	return []collections.Index[sdk.AccAddress, authtypes.BaseAccount]{a.Number}
//}
