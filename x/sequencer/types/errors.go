package types

// DONTCOVER

import (
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// x/sequencer module sentinel errors
var (
	ErrSequencerExists          = gerrc.ErrAlreadyExists.Wrap("sequencer")
	ErrUnknownRollappID         = gerrc.ErrNotFound.Wrap("rollapp does not exist")
	ErrUnknownSequencer         = gerrc.ErrNotFound.Wrap("sequencer was not registered")
	ErrSequencerRollappMismatch = gerrc.ErrFailedPrecondition.Wrap("sequencer was not registered for this rollapp")
	ErrNotActiveSequencer       = gerrc.ErrFailedPrecondition.Wrap("sequencer is not active")
	ErrInvalidSequencerStatus   = gerrc.ErrInvalidArgument.Wrap("sequencer status")
	ErrInvalidCoinDenom         = gerrc.ErrInvalidArgument.Wrap("coin denomination")
	ErrInsufficientBond         = gerrc.ErrOutOfRange.Wrap("insufficient bond")
	ErrRollappJailed            = gerrc.ErrFailedPrecondition.Wrap("rollapp is jailed")
	ErrInvalidAddress           = gerrc.ErrInvalidArgument.Wrap("address")
	ErrInvalidPubKey            = gerrc.ErrInvalidArgument.Wrap("pubkey")
	ErrInvalidCoins             = gerrc.ErrInvalidArgument.Wrap("coins")
	ErrInvalidType              = gerrc.ErrInvalidArgument.Wrap("type")
	ErrInvalidRequest           = gerrc.ErrInvalidArgument.Wrap("request")
	ErrSequencerJailed          = gerrc.ErrFailedPrecondition.Wrap("sequencer is jailed")
	ErrNotInitialSequencer      = gerrc.ErrFailedPrecondition.Wrap("not the initial sequencer")
	ErrInvalidURL               = gerrc.ErrInvalidArgument.Wrap("url")
	ErrInvalidMetadata          = gerrc.ErrInvalidArgument.Wrap("metadata")
)
