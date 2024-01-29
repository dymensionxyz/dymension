package keeper

import (
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

var _ types.QueryServer = Keeper{}
