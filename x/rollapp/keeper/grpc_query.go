package keeper

import (
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

var _ types.QueryServer = Keeper{}
