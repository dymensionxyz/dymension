package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
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
	ErrInvalidSequencerStatus   = sdkerrors.Register(ModuleName, 1008, "invalid sequencer status")
	ErrInvalidSequencerTokens   = sdkerrors.Register(ModuleName, 1009, "invalid sequencer tokens")
	ErrInvalidCoinDenom         = sdkerrors.Register(ModuleName, 1010, "invalid coin denomination")
	ErrInsufficientBond         = sdkerrors.Register(ModuleName, 1011, "insufficient bond")
	ErrRollappJailed            = sdkerrors.Register(ModuleName, 1012, "rollapp is jailed")
	ErrInvalidAddress           = sdkerrors.Register(ModuleName, 1013, "invalid address")
	ErrInvalidPubKey            = sdkerrors.Register(ModuleName, 1014, "invalid pubkey")
	ErrInvalidCoins             = sdkerrors.Register(ModuleName, 1015, "invalid coins")
	ErrInvalidType              = sdkerrors.Register(ModuleName, 1016, "invalid type")
	ErrUnknownRequest           = sdkerrors.Register(ModuleName, 1017, "unknown request")
	ErrInvalidRequest           = sdkerrors.Register(ModuleName, 1018, "invalid request")
)
