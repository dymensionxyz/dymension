package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var (
	ErrInvalidURL                                      = errorsmod.Wrap(gerrc.ErrInvalidArgument, "url")
	ErrInvalidMetadata                                 = errorsmod.Wrap(gerrc.ErrInvalidArgument, "metadata")
	ErrInvalidVMTypeUpdate                             = errorsmod.Wrap(gerrc.ErrInvalidArgument, "vm type update")
	ErrBeforePreLaunchTime                             = errorsmod.Wrap(gerrc.ErrFailedPrecondition, "before pre-launch time")
	ErrSequencerAlreadyExists                          = gerrc.ErrAlreadyExists.Wrap("sequencer")
	ErrSequencerNotFound                               = gerrc.ErrNotFound.Wrap("sequencer")
	ErrUnbondNotAllowed                                = gerrc.ErrFailedPrecondition.Wrap("unbond not allowed")
	ErrUnbondProposerOrSuccessor                       = errorsmod.Wrap(ErrUnbondNotAllowed, "proposer or successor")
	ErrInvalidCoins                                    = gerrc.ErrInvalidArgument.Wrap("coin or coins")
	ErrInvalidDenom                                    = errorsmod.Wrap(ErrInvalidCoins, "denom")
	ErrInvalidCoinAmount                               = errorsmod.Wrap(ErrInvalidCoins, "amount")
	ErrInsufficientBond                                = gerrc.ErrOutOfRange.Wrap("bond")
	ErrRegisterSequencerWhileAwaitingLastProposerBlock = gerrc.ErrFailedPrecondition.Wrap("register sequencer while awaiting last proposer block")
	ErrNotInitialSequencer                             = errorsmod.Wrap(gerrc.ErrFailedPrecondition, "not the initial sequencer")
	ErrInvalidAddr                                     = gerrc.ErrInvalidArgument.Wrap("address")
	ErrInvalidPubKey                                   = gerrc.ErrInvalidArgument.Wrap("pubkey")
)
