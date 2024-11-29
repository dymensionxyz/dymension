package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/params"
	"github.com/dymensionxyz/sdk-utils/utils/ucoin"
)

var (
	BondDenom = params.BaseDenom
	// for tests, real value is supplied by rollapp keeper
	TestMinBondDym = int64(100)
	TestMinBond    = ucoin.SimpleMul(sdk.NewCoin(BondDenom, sdk.NewInt(params.BaseDenomUnit)), TestMinBondDym)
)
