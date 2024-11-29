package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/params"
	"github.com/dymensionxyz/sdk-utils/utils/ucoin"
)

var (
	BondDenom = params.BaseDenom
	// for tests, real value is supplied by rollapp keeper
	TestMinBond = ucoin.SimpleMul(sdk.NewCoin(BondDenom, sdk.NewInt(params.BaseDenomUnit)), 100)
)
