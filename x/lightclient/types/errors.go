package types

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var (
	ErrStateRootMismatch    = errorsmod.Wrap(gerrc.ErrFault, "block descriptor state root does not match tendermint header app hash")
	ErrNextValHashMismatch  = errorsmod.Wrap(gerrc.ErrFault, "next validator hash on light client cons state does not match the sequencer for h+1 from the state info")
	ErrTimestampMismatch    = errorsmod.Wrap(gerrc.ErrFault, "block descriptor timestamp does not match tendermint header timestamp")
	ErrorHardForkInProgress = errorsmod.Wrap(gerrc.ErrFailedPrecondition, "update light client while fork in progress")
)
