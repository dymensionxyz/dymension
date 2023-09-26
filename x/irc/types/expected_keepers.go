package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	ibc "github.com/cosmos/ibc-go/v6/modules/core"
	rollapptypes "github.com/dymensionxyz/dymension/x/rollapp/types"
)

// RollappKeeper defines the expected rollapp keeper used for retrieve rollapp.
type RollappKeeper interface {
	GetRollapp(ctx sdk.Context, rollappId string) (val rollapptypes.Rollapp, found bool)
	FindStateInfoByHeight(ctx sdk.Context, rollappId string, height uint64) (*rollapptypes.StateInfo, error)
	// Methods imported from rollapp should be defined here
}

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) types.AccountI
	// Methods imported from account should be defined here
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	// Methods imported from bank should be defined here
}

// IBCKeeper defines the expected interface needed to transffer the handling to the IBC stack.
type IBCKeeper interface {
	ibc.IBCMsgI
}
