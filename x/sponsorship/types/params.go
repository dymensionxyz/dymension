package types

import "fmt"

func DefaultParams() Params {
	return Params{
		MinAllocationWeight: DefaultMinAllocationWeight,
		MinVotingPower:      DefaultMinVotingPower,
		EpochIdentifier:     DefaultEpochIdentifier,
	}
}

func (p Params) ValidateBasic() error {
	if p.MinAllocationWeight.IsNegative() {
		return ErrInvalidParams.Wrapf("MinAllocationWeight must be >= 0, got %s", p.MinAllocationWeight)
	}
	if p.MinAllocationWeight.GT(MaxAllocationWeight) {
		return ErrInvalidParams.Wrapf("MinAllocationWeight must be <= 100 * 10^18, got %s", p.MinAllocationWeight)
	}
	if p.MinVotingPower.IsNegative() {
		return ErrInvalidParams.Wrapf("MinVotingPower must be >= 0, got %s", p.MinVotingPower)
	}
	if p.EpochIdentifier == "" {
		return fmt.Errorf("epoch identifier cannot be empty")
	}
	return nil
}
