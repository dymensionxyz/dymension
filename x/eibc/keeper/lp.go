package keeper

import (
	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

var LPsPrefix = collections.NewPrefix(0)
var LPsIndexesAccNumberPrefix = collections.NewPrefix(1)

type LPs struct {
	//Accounts *collections.IndexedMap[sdk.AccAddress, authtypes.BaseAccount, LPsIndexes]

	/*
		Ideas
		<rollapp,denom> ->
	*/
	//lps collections.IndexedMap[
	//collections.Triple[rollapp,denom,]
	m collections.Map[int64, int64]
}

func makeLPsStore(sb *collections.SchemaBuilder, cdc codec.BinaryCodec) LPs {
	return LPs{
		//Accounts: collections.NewIndexedMap(
		//	sb, LPsPrefix, "accounts",
		//	sdk.AccAddressKey, codec.CollValue[authtypes.BaseAccount](cdc),
		//	NewLPsIndexes(sb),
		m: collections.NewMap(sb,
			collections.NewPrefix(1), "m",
			collections.Int64Key, collections.Int64Value),
	}
}

func (s LPs) findLP(o types.DemandOrder) *types.OnDemandLiquidity {
	return nil
}

type LPsIndexes struct {
	Number *indexes.Unique[uint64, sdk.AccAddress, authtypes.BaseAccount]
}

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

func (a LPsIndexes) IndexesList() []collections.Index[sdk.AccAddress, authtypes.BaseAccount] {
	return []collections.Index[sdk.AccAddress, authtypes.BaseAccount]{a.Number}
}
