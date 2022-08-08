package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/sequencer module sentinel errors
var (
	ErrSequencerExists         = sdkerrors.Register(ModuleName, 1000, "sequencer already exist for this address; must use new sequencer address")
	ErrInvalidSequencerAddress = sdkerrors.Register(ModuleName, 1001, "invalid sequencer address")
)
