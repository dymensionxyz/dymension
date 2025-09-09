package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultParams returns the default parameters for the Otcbuyback module
func DefaultParams() Params {
	return Params{
		AcceptedTokens: []AcceptedToken{},
	}
}

// Validate validates the parameters
func (p Params) ValidateBasic() error {
	if len(p.AcceptedTokens) == 0 {
		return fmt.Errorf("accepted tokens cannot be empty")
	}

	// Validate each token denom format (basic validation)
	for _, token := range p.AcceptedTokens {
		if err := sdk.ValidateDenom(token.Token); err != nil {
			return err
		}
	}

	return nil
}
