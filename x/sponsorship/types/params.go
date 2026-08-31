package types

import errorsmod "cosmossdk.io/errors"

func DefaultParams() Params {
	return Params{
		MinAllocationWeight: DefaultMinAllocationWeight,
		MinVotingPower:      DefaultMinVotingPower,
	}
}

func (p Params) ValidateBasic() error {
	if p.MinAllocationWeight.IsNegative() {
		return errorsmod.Wrapf(ErrInvalidParams, "MinAllocationWeight must be >= 0, got %s", p.MinAllocationWeight)
	}
	if p.MinAllocationWeight.GT(MaxAllocationWeight) {
		return errorsmod.Wrapf(ErrInvalidParams, "MinAllocationWeight must be <= 100 * 10^18, got %s", p.MinAllocationWeight)
	}
	if p.MinVotingPower.IsNegative() {
		return errorsmod.Wrapf(ErrInvalidParams, "MinVotingPower must be >= 0, got %s", p.MinVotingPower)
	}
	return nil
}
