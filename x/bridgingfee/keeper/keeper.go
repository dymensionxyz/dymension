package keeper

import (
	"cosmossdk.io/collections"
	"github.com/dymensionxyz/dymension/v3/x/bridgingfee/types"
)

type Keeper struct {
	params collections.Item[types.Params]
}
