package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	incentivestypes "github.com/dymensionxyz/dymension/v3/x/incentives/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
}

type IncentivesKeeper interface {
	GetGauges(ctx sdk.Context) []incentivestypes.Gauge
}

type StakingKeeper interface {
	GetAllValidators(ctx context.Context) ([]stakingtypes.Validator, error)
	GetValidatorDelegations(ctx context.Context, valAddr sdk.ValAddress) ([]stakingtypes.Delegation, error)
}
