package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/dymensionxyz/dymension/v3/x/forward/types"
)

func getRecoveryAddress(recovery types.Recovery) sdk.AccAddress {
	return recovery.MustAddr()
}
