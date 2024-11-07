package types

var (
	KeyMinAllocationWeight = []byte("MinAllocationWeight")
	KeyMinVotingPower      = []byte("MinVotingPower")
)

func DefaultParams() Params {
	return Params{
		MinAllocationWeight: DefaultMinAllocationWeight,
		MinVotingPower:      DefaultMinVotingPower,
	}
}

func (p Params) Validate() error {
	if p.MinAllocationWeight.IsNegative() {
		return ErrInvalidParams.Wrapf("MinAllocationWeight must be >= 0, got %s", p.MinAllocationWeight)
	}
	if p.MinAllocationWeight.GT(MaxAllocationWeight) {
		return ErrInvalidParams.Wrapf("MinAllocationWeight must be <= 100 * 10^18, got %s", p.MinAllocationWeight)
	}
	if p.MinVotingPower.IsNegative() {
		return ErrInvalidParams.Wrapf("MinVotingPower must be >= 0, got %s", p.MinVotingPower)
	}
	return nil
}
