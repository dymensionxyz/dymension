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

func (gi GenesisInfo) AllSet() bool {
	return gi.GenesisChecksum != "" &&
		gi.NativeDenom.IsSet() &&
		gi.Bech32Prefix != "" &&
		!gi.InitialSupply.IsNil()
}

func (gi GenesisInfo) Validate() error {
	if gi.Bech32Prefix != "" {
		if err := validateBech32Prefix(gi.Bech32Prefix); err != nil {
			return errors.Join(ErrInvalidBech32Prefix, err)
		}
	}

	if len(gi.GenesisChecksum) > maxGenesisChecksumLength {
		return ErrInvalidGenesisChecksum
	}

	if gi.NativeDenom.IsSet() {
		if err := gi.NativeDenom.Validate(); err != nil {
			return errorsmod.Wrap(ErrInvalidNativeDenom, err.Error())
		}
	}

	if !gi.InitialSupply.IsNil() {
		if !gi.InitialSupply.IsPositive() {
			return ErrInvalidInitialSupply
		}
	}

	if l := len(gi.Accounts()); l > maxAllowedGenesisAccounts {
		return fmt.Errorf("too many genesis accounts: %d", l)
	}

	accountSet := make(map[string]struct{})
	for _, a := range gi.Accounts() {
		if err := a.ValidateBasic(); err != nil {
			return errors.Join(gerrc.ErrInvalidArgument, err)
		}
		if _, exists := accountSet[a.Address]; exists {
			return fmt.Errorf("duplicate genesis account: %s", a.Address)
		}
		accountSet[a.Address] = struct{}{}
	}
	return nil
}

func (a GenesisAccount) ValidateBasic() error {
	if !a.Amount.IsNil() && !a.Amount.IsPositive() {
		return fmt.Errorf("invalid amount: %s %s", a.Address, a.Amount)
	}

	if _, err := sdk.AccAddressFromBech32(a.Address); err != nil {
		return err
	}
	return nil
}
