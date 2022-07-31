package keeper

import (
	"github.com/dymensionxyz/dymension/x/rollapp/types"
)

var _ types.QueryServer = Keeper{}
