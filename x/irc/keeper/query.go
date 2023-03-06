package keeper

import (
	"github.com/dymensionxyz/dymension/x/irc/types"
)

var _ types.QueryServer = Keeper{}
