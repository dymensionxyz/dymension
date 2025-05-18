package types

// NewParams creates a new Params instance
func NewParams(maxIterationsPerBlock uint64) Params {
	return Params{
		MaxIterationsPerBlock: maxIterationsPerBlock,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultMaxIterationsPerBlock)
}

// Validate validates the set of params
func (p Params) ValidateBasic() error {
	return nil
}
