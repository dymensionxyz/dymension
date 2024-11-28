package types

import (
	"errors"
	"fmt"
	"slices"

	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// HubRecipient is the address of `x/rollapp` module's account on the rollapp chain.
const HubRecipient = "dym1mk7pw34ypusacm29m92zshgxee3yreums8avur"

type GenesisBridgeValidator struct {
	rollapp GenesisBridgeData // what the rollapp sent over IBC
	hub     GenesisInfo       // what the rollapp thinks is correct
}

func NewGenesisBridgeValidator(
	rollappGenesis GenesisBridgeData,
	hubGenesis GenesisInfo,
) *GenesisBridgeValidator {
	return &GenesisBridgeValidator{
		rollapp: rollappGenesis,
		hub:     hubGenesis,
	}
}

func (v *GenesisBridgeValidator) Validate() error {
	if err := v.rollapp.ValidateBasic(); err != nil {
		return errorsmod.Wrap(err, "validate basic genesis bridge data")
	}

	if err := validateAgainstHub(v.rollapp.GenesisInfo, v.hub); err != nil {
		return errorsmod.Wrap(err, "validate against rollapp")
	}

	err := v.rollapp.NativeDenom.Validate()
	if err != nil {
		return errorsmod.Wrap(errors.Join(gerrc.ErrInvalidArgument, err), "metadata validate")
	}

	err = v.validateGenesisTransfer()
	if err != nil {
		return errorsmod.Wrap(err, "validate genesis transfer")
	}

	return nil
}

func validateAgainstHub(rollapp GenesisBridgeInfo, hub GenesisInfo) error {
	if rollapp.GenesisChecksum != hub.GenesisChecksum {
		return fmt.Errorf("genesis checksum mismatch: expected: %v, got: %v", hub.GenesisChecksum, rollapp.GenesisChecksum)
	}

	if rollapp.Bech32Prefix != hub.Bech32Prefix {
		return fmt.Errorf("bech32 prefix mismatch: expected: %v, got: %v", hub.Bech32Prefix, rollapp.Bech32Prefix)
	}

	if rollapp.NativeDenom != hub.NativeDenom {
		return fmt.Errorf("native denom mismatch: expected: %v, got: %v", hub.NativeDenom, rollapp.NativeDenom)
	}

	if !rollapp.InitialSupply.Equal(hub.InitialSupply) {
		return fmt.Errorf("initial supply mismatch: expected: %v, got: %v", hub.InitialSupply, rollapp.InitialSupply)
	}

	err := compareGenesisAccounts(hub.Accounts(), rollapp.Accounts())
	if err != nil {
		return errorsmod.Wrap(err, "genesis accounts mismatch")
	}
	return nil
}

func compareGenesisAccounts(raCommitted []GenesisAccount, gbData []GenesisAccount) error {
	if len(raCommitted) != len(gbData) {
		return fmt.Errorf("genesis accounts length mismatch: expected %d, got %d", len(raCommitted), len(gbData))
	}

	for _, acc := range raCommitted {
		found := slices.ContainsFunc(gbData, func(dataAcc GenesisAccount) bool {
			return dataAcc.Address == acc.Address && dataAcc.Amount.Equal(acc.Amount)
		})

		if !found {
			return fmt.Errorf("genesis account mismatch: account %s with amount %v not found in data", acc.Address, acc.Amount)
		}
	}

	return nil
}

// validateGenesisTransfer validates the genesis transfer.
func (v *GenesisBridgeValidator) validateGenesisTransfer() error {
	gTransfer := v.rollapp.GenesisTransfer
	requiresTransfer := v.hub.RequiresTransfer()

	// required but not present
	if requiresTransfer && gTransfer == nil {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "genesis transfer required")
	}
	// not required but present
	if !requiresTransfer && gTransfer != nil {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "genesis transfer not expected")
	}
	if gTransfer == nil {
		return nil
	}

	// validate the receiver
	if gTransfer.Receiver != HubRecipient {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "receiver mismatch")
	}

	// validate that the transfer amount matches the expected amount, which is the sum of all genesis accounts
	expectedAmount := v.hub.GenesisTransferAmount()
	if expectedAmount.String() != gTransfer.Amount {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "amount mismatch")
	}

	return nil
}
