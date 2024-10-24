package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var (
	ErrInvalidURL          = errorsmod.Wrap(gerrc.ErrInvalidArgument, "url")
	ErrInvalidMetadata     = errorsmod.Wrap(gerrc.ErrInvalidArgument, "metadata")
	ErrInvalidVMTypeUpdate = errorsmod.Wrap(gerrc.ErrInvalidArgument, "vm type update")
	ErrBeforePreLaunchTime = errorsmod.Wrap(gerrc.ErrFailedPrecondition, "before pre-launch time")

	ErrSequencerAlreadyExists    = gerrc.ErrAlreadyExists.Wrap("sequencer")
	ErrRollappNotFound           = gerrc.ErrNotFound.Wrap("rollapp")
	ErrSequencerNotFound         = gerrc.ErrNotFound.Wrap("sequencer")
	ErrUnbondNotAllowed          = gerrc.ErrFailedPrecondition.Wrap("unbond not allowed")
	ErrUnbondProposerOrSuccessor = errorsmod.Wrap(ErrUnbondNotAllowed, "proposer or successor")
	ErrInvalidDenom              = gerrc.ErrInvalidArgument.Wrap("denom")
	ErrInsufficientBond          = gerrc.ErrOutOfRange.Wrap("bond")

	// ErrNotActiveSequencer ..
	// Deprecated: ..
	ErrNotActiveSequencer = errorsmod.Register(ModuleName, 1007, "sequencer is not active")
	// ErrInvalidSequencerStatus ..
	// Deprecated: ..
	ErrInvalidSequencerStatus = errorsmod.Register(ModuleName, 1008, "invalid sequencer status")
	// ErrInvalidCoinDenom ..
	// Deprecated: ..
	ErrInvalidCoinDenom = errorsmod.Register(ModuleName, 1010, "invalid coin denomination")
	// ErrInvalidAddress ..
	// Deprecated: ..
	ErrInvalidAddress = errorsmod.Register(ModuleName, 1013, "invalid address")
	// ErrInvalidPubKey ..
	// Deprecated: ..
	ErrInvalidPubKey = errorsmod.Register(ModuleName, 1014, "invalid pubkey")
	// ErrInvalidCoins ..
	// Deprecated: ..
	ErrInvalidCoins = errorsmod.Register(ModuleName, 1015, "invalid coins")
	// ErrInvalidType ..
	// Deprecated: ..
	ErrInvalidType = errorsmod.Register(ModuleName, 1016, "invalid type")
	// ErrUnknownRequest ..
	// Deprecated: ..
	ErrUnknownRequest = errorsmod.Register(ModuleName, 1017, "unknown request")
	// ErrInvalidRequest ..
	// Deprecated: ..
	ErrInvalidRequest = errorsmod.Register(ModuleName, 1018, "invalid request")
	// ErrSequencerJailed ..
	// Deprecated: ..
	ErrSequencerJailed = errorsmod.Register(ModuleName, 1019, "sequencer is jailed")
	// ErrRotationInProgress ..
	// Deprecated: ..
	ErrRotationInProgress = errorsmod.Wrap(gerrc.ErrFailedPrecondition, "sequencer rotation in progress")
	// ErrNoProposer ..
	// Deprecated: ..
	ErrNoProposer = errorsmod.Wrap(gerrc.ErrNotFound, "proposer")
	// ErrNotInitialSequencer ..
	// Deprecated: ..
	ErrNotInitialSequencer = errorsmod.Wrap(gerrc.ErrFailedPrecondition, "not the initial sequencer")
)
