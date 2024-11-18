package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var (
	ErrInvalidURL                = errorsmod.Wrap(gerrc.ErrInvalidArgument, "url")
	ErrInvalidMetadata           = errorsmod.Wrap(gerrc.ErrInvalidArgument, "metadata")
	ErrInvalidVMTypeUpdate       = errorsmod.Wrap(gerrc.ErrInvalidArgument, "vm type update")
	ErrBeforePreLaunchTime       = errorsmod.Wrap(gerrc.ErrFailedPrecondition, "before pre-launch time")
	ErrNotProposer               = gerrc.ErrInvalidArgument.Wrap("sequencer is not proposer")
	ErrSequencerAlreadyExists    = gerrc.ErrAlreadyExists.Wrap("sequencer")
	ErrSequencerNotFound         = gerrc.ErrNotFound.Wrap("sequencer")
	ErrUnbondNotAllowed          = gerrc.ErrFailedPrecondition.Wrap("unbond not allowed")
	ErrUnbondProposerOrSuccessor = errorsmod.Wrap(ErrUnbondNotAllowed, "proposer or successor")
	ErrInvalidCoins              = gerrc.ErrInvalidArgument.Wrap("coin or coins")
	ErrInvalidDenom              = errorsmod.Wrap(ErrInvalidCoins, "denom")
	ErrInsufficientBond          = gerrc.ErrOutOfRange.Wrap("bond")
	ErrNotInitialSequencer       = errorsmod.Wrap(gerrc.ErrFailedPrecondition, "not the initial sequencer")
	ErrInvalidAddr               = gerrc.ErrInvalidArgument.Wrap("address")
	ErrInvalidPubKey             = gerrc.ErrInvalidArgument.Wrap("pubkey")
)
