package types

func DefaultParams() Params {
	return Params{
		MinAllocationWeight: DefaultMinAllocationWeight,
		MinVotingPower:      DefaultMinVotingPower,
	}
}

func (p Params) Validate() error {
	if p.MinAllocationWeight.IsNegative() {
		return ErrInvalidParams.Wrapf("MinAllocationWeight must be >= 0, got %d", p.MinAllocationWeight.Int64())
	}
	if p.MinAllocationWeight.GT(hundred) {
		return ErrInvalidParams.Wrapf("MinAllocationWeight must be <= 100, got %d", p.MinAllocationWeight.Int64())
	}
	if p.MinVotingPower.IsNegative() {
		return ErrInvalidParams.Wrapf("MinVotingPower must be >= 0, got %d", p.MinVotingPower.Int64())
	}
	return nil
}
