package types

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var (
	ErrStateRootsMismatch    = errorsmod.Wrap(gerrc.ErrFailedPrecondition, "block descriptor state root does not match tendermint header app hash")
	ErrValidatorHashMismatch = errorsmod.Wrap(gerrc.ErrFailedPrecondition, "next validator hash does not match the sequencer for h+1")
	ErrTimestampMismatch     = errorsmod.Wrap(gerrc.ErrFailedPrecondition, "block descriptor timestamp does not match tendermint header timestamp")
)
