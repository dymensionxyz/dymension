package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"

	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

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

// RollAppKeeper defines the expected interface needed to retrieve account balances.
type RollAppKeeper interface {
	GetRollapp(ctx sdk.Context, id string) (rollapptypes.Rollapp, bool)
	// Methods imported from rollapp should be defined here
}
