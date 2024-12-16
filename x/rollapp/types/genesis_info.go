package types

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// We set the maximum amount of genesis accounts to 100
const maxAllowedGenesisAccounts = 100

// Handling should be based on length and contents, not nil status
func (gi GenesisInfo) Accounts() []GenesisAccount {
	if gi.GenesisAccounts == nil {
		return nil
	}
	return gi.GenesisAccounts.Accounts
}

func (gi GenesisInfo) RequiresTransfer() bool {
	return 0 < len(gi.Accounts())
}

func (gi GenesisInfo) GenesisTransferAmount() math.Int {
	total := math.ZeroInt()
	for _, a := range gi.Accounts() {
		total = total.Add(a.Amount)
	}
	return total
}

// Launchable checks if the genesis info has all the necessary fields set
// - genesis checksum
// - bech32 prefix
// - initial supply
func (gi GenesisInfo) Launchable() bool {
	return gi.GenesisChecksum != "" &&
		gi.Bech32Prefix != "" &&
		!gi.InitialSupply.IsNil() // can be 0, but needs to be set
}

func (gi GenesisInfo) IROReady() bool {
	return gi.Launchable() && gi.NativeDenom.IsSet()
}

// ValidateBasic performs basic validation checks on the GenesisInfo.
// - bech32 prefix
// - genesis checksum
// - native denom, if set
// - initial supply >= 0, if set
//
// - valid genesis accounts
//   - no duplicates
//   - no more than 100 genesis accounts
//   - initial supply >= sum of genesis accounts
//
// - if no native denom,
//   - initial supply must be 0
//   - no genesis accounts
func (gi GenesisInfo) ValidateBasic() error {
	if gi.Bech32Prefix != "" {
		if err := validateBech32Prefix(gi.Bech32Prefix); err != nil {
			return errors.Join(ErrInvalidBech32Prefix, err)
		}
	}

	if len(gi.GenesisChecksum) > maxGenesisChecksumLength {
		return ErrInvalidGenesisChecksum
	}

	numGenesisAccounts := len(gi.Accounts())

	// if native denom is not set, initial supply must be 0 and no accounts
	if !gi.NativeDenom.IsSet() {
		if !gi.InitialSupply.IsNil() && !gi.InitialSupply.IsZero() {
			return errorsmod.Wrap(ErrNoNativeTokenRollapp, "non zero initial supply")
		}

		if numGenesisAccounts > 0 {
			return errorsmod.Wrap(ErrNoNativeTokenRollapp, "non empty genesis accounts")
		}

		return nil
	}

	if err := gi.NativeDenom.Validate(); err != nil {
		return errors.Join(ErrInvalidMetadata, err)
	}

	if !gi.InitialSupply.IsNil() && gi.InitialSupply.IsNegative() {
		return ErrInvalidInitialSupply
	}

	if numGenesisAccounts > 0 {
		if numGenesisAccounts > maxAllowedGenesisAccounts {
			return ErrTooManyGenesisAccounts
		}

		if gi.InitialSupply.IsNil() {
			return ErrInvalidInitialSupply
		}

		total := math.ZeroInt()
		accountSet := make(map[string]struct{})
		for _, a := range gi.Accounts() {
			if err := a.ValidateBasic(); err != nil {
				return errors.Join(gerrc.ErrInvalidArgument, err)
			}
			if _, exists := accountSet[a.Address]; exists {
				return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "duplicate genesis account: %s", a.Address)
			}
			accountSet[a.Address] = struct{}{}

			total = total.Add(a.Amount)
		}

		if total.GT(gi.InitialSupply) {
			return ErrInvalidInitialSupply
		}
	}

	return nil
}

func (a GenesisAccount) ValidateBasic() error {
	if a.Amount.IsNil() || !a.Amount.IsPositive() {
		return fmt.Errorf("invalid amount: %s", a.Address)
	}

	if _, err := sdk.AccAddressFromBech32(a.Address); err != nil {
		return err
	}
	return nil
}
