package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/app/params"
)

// DefaultParams returns the default parameters for the Otcbuyback module
func DefaultParams() Params {
	return Params{
		AcceptedTokens: []string{params.BaseDenom},
	}
}

// Validate validates the parameters
func (p Params) ValidateBasic() error {
	if len(p.AcceptedTokens) == 0 {
		return fmt.Errorf("accepted tokens cannot be empty")
	}

	// Validate each token denom format (basic validation)
	for _, token := range p.AcceptedTokens {
		if err := sdk.ValidateDenom(token); err != nil {
			return err
		}
	}

	return nil
}
