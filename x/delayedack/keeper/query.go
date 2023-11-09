package keeper

import (
	"github.com/dymensionxyz/dymension/x/delayedack/types"
)

var _ types.QueryServer = Keeper{}
