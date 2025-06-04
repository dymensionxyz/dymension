package types

import (
	"errors"

	errorsmod "cosmossdk.io/errors"
	hyperutil "github.com/bcp-innovations/hyperlane-cosmos/util"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func DefaultGenesis() *GenesisState {
	return &GenesisState{}
}

func (genState GenesisState) Validate() error {
	if genState.Mailbox != "" {
		if _, err := hyperutil.DecodeHexAddress(genState.Mailbox); err != nil {
			return errorsmod.Wrapf(errors.Join(err, gerrc.ErrInvalidArgument), "mailbox")
		}
	}
	if genState.Ism != "" {
		if _, err := hyperutil.DecodeHexAddress(genState.Ism); err != nil {
			return errorsmod.Wrapf(errors.Join(err, gerrc.ErrInvalidArgument), "ism")
		}
	}
	if genState.Outpoint != nil {
		if err := genState.Outpoint.ValidateBasic(); err != nil {
			return errorsmod.Wrapf(errors.Join(err, gerrc.ErrInvalidArgument), "outpoint")
		}
	}
	for _, w := range genState.ProcessedWithdrawals {
		if err := w.ValidateBasic(); err != nil {
			return errorsmod.Wrapf(errors.Join(err, gerrc.ErrInvalidArgument), "processed withdrawal")
		}
	}
	return nil
}
