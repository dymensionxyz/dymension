package keeper

import (
	"github.com/dymensionxyz/dymension/v3/x/kas/types"
)

var _ types.QueryServer = Keeper{}
