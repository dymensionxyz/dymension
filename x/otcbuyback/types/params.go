package types

// DefaultParams returns the default parameters for the Otcbuyback module
func DefaultParams() Params {
	return Params{
		AcceptedTokens: []AcceptedToken{},
	}
}
