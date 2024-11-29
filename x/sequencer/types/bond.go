package types

import (
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/sdk-utils/utils/ucoin"
)

var (
	// for tests, real value is supplied by rollapp keeper
	TestMinBondDym = int64(100)
	TestMinBond    = ucoin.SimpleMul(commontypes.DYMCoin, TestMinBondDym)
)
