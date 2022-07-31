package keeper

import (
	"github.com/dymensionxyz/dymension/x/sequencer/types"
)

var _ types.QueryServer = Keeper{}
