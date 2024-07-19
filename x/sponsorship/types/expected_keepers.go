package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	incentivestypes "github.com/osmosis-labs/osmosis/v15/x/incentives/types"
)

// AccountKeeper defines the contract required for account APIs.
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
}

type StakingKeeper interface {
	GetBondedValidatorsByPower(sdk.Context) []stakingtypes.Validator
	GetDelegatorValidator(ctx sdk.Context, delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress) (stakingtypes.Validator, error)
	GetDelegation(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (stakingtypes.Delegation, bool)
	IterateBondedValidatorsByPower(sdk.Context, func(index int64, validator stakingtypes.ValidatorI) (stop bool))
	IterateDelegations(
		ctx sdk.Context,
		delAddr sdk.AccAddress,
		fn func(index int64, del stakingtypes.DelegationI) (stop bool),
	)
}

type IncentivesKeeper interface {
	GetGaugeByID(ctx sdk.Context, gaugeID uint64) (*incentivestypes.Gauge, error)
}
