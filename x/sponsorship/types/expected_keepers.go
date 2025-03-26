package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	incentivestypes "github.com/dymensionxyz/dymension/v3/x/incentives/types"
)

// AccountKeeper defines the contract required for account APIs.
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
}

type StakingKeeper interface {
	GetValidator(ctx context.Context, addr sdk.ValAddress) (stakingtypes.Validator, error)
	GetDelegation(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (stakingtypes.Delegation, error)
	IterateDelegatorDelegations(ctx context.Context, delegator sdk.AccAddress, cb func(stakingtypes.Delegation) (stop bool)) error
}

type IncentivesKeeper interface {
	GetGaugeByID(ctx sdk.Context, gaugeID uint64) (*incentivestypes.Gauge, error)
	DistributeEndorsementRewards(ctx sdk.Context, user sdk.AccAddress, gaugeId uint64, rewards sdk.Coins) error
}

type BankKeeper interface {
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}
