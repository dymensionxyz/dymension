package keeper

import (
	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/codec"
)

/*
	x := collections.NewItem(sb, collections.NewPrefix(1), "x", collections.Int64Value)
	z := collections.NewMap()
	collections.NewIndexedMap()
	indexes.NewMulti()
	collections.Range[]{}
	collections.New()
*/

type LPs struct {
	lps collections.Map
}

func createLPsStore(sb *collections.SchemaBuilder, cdc codec.BinaryCodec) LPs {

}

func (s LPs) findLP(o Order) *ProvideOnDemandLP {