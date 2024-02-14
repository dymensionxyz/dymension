package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/sequencer module sentinel errors
var (
	ErrSequencerExists          = sdkerrors.Register(ModuleName, 1000, "sequencer already exist for this address; must use new sequencer address")
	ErrInvalidSequencerAddress  = sdkerrors.Register(ModuleName, 1001, "invalid sequencer address")
	ErrUnknownRollappID         = sdkerrors.Register(ModuleName, 1002, "rollapp does not exist")
	ErrMaxSequencersLimit       = sdkerrors.Register(ModuleName, 1003, "too many sequencers for rollapp")
	ErrSequencerNotPermissioned = sdkerrors.Register(ModuleName, 1004, "sequencer is not permissioned for serving the rollapp")
	ErrUnknownSequencer         = sdkerrors.Register(ModuleName, 1005, "sequencer was not registered")
	ErrSequencerRollappMismatch = sdkerrors.Register(ModuleName, 1006, "sequencer was not registered for this rollapp")
	ErrNotActiveSequencer       = sdkerrors.Register(ModuleName, 1007, "sequencer is not active")
)
