package types

import (
	"errors"
	fmt "fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// We set the maximum amount of genesis accounts to 100
const maxAllowedGenesisAccounts = 100

func (gi GenesisInfo) GenesisTransferAmount() math.Int {
	total := math.ZeroInt()
	for _, a := range gi.GenesisAccounts {
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

	// validate max limit of genesis accounts
	if len(gi.GenesisAccounts) > maxAllowedGenesisAccounts {
		return fmt.Errorf("too many genesis accounts: %d", len(gi.GenesisAccounts))
	}

	for _, a := range gi.GenesisAccounts {
		if err := a.ValidateBasic(); err != nil {
			return errors.Join(gerrc.ErrInvalidArgument, err)
		}
	}
	return nil
}

func (a GenesisAccount) ValidateBasic() error {
	if !a.Amount.IsPositive() {
		return fmt.Errorf("invalid amount: %s %s", a.Address, a.Amount)
	}

	if _, err := sdk.AccAddressFromBech32(a.Address); err != nil {
		return err
	}
	return nil
}
