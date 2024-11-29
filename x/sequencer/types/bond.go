package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/params"
)

var (
	BondDenom   = params.BaseDenom
	TestMinBond = sdk.NewCoin(BondDenom, sdk.NewInt(params.BaseDenomUnit))
)
