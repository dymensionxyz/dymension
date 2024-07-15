package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

// x/sequencer module sentinel errors
var (
	ErrSequencerExists          = errorsmod.Register(ModuleName, 1000, "sequencer already exist for this address; must use new sequencer address")
	ErrInvalidSequencerAddress  = errorsmod.Register(ModuleName, 1001, "invalid sequencer address")
	ErrUnknownRollappID         = errorsmod.Register(ModuleName, 1002, "rollapp does not exist")
	ErrMaxSequencersLimit       = errorsmod.Register(ModuleName, 1003, "too many sequencers for rollapp")
	ErrSequencerNotPermissioned = errorsmod.Register(ModuleName, 1004, "sequencer is not permissioned for serving the rollapp")
	ErrUnknownSequencer         = errorsmod.Register(ModuleName, 1005, "sequencer was not registered")
	ErrSequencerRollappMismatch = errorsmod.Register(ModuleName, 1006, "sequencer was not registered for this rollapp")
	ErrNotActiveSequencer       = errorsmod.Register(ModuleName, 1007, "sequencer is not active")
	ErrInvalidSequencerStatus   = errorsmod.Register(ModuleName, 1008, "invalid sequencer status")
	ErrInvalidSequencerTokens   = errorsmod.Register(ModuleName, 1009, "invalid sequencer tokens")
	ErrInvalidCoinDenom         = errorsmod.Register(ModuleName, 1010, "invalid coin denomination")
	ErrInsufficientBond         = errorsmod.Register(ModuleName, 1011, "insufficient bond")
	ErrRollappJailed            = errorsmod.Register(ModuleName, 1012, "rollapp is jailed")
	ErrInvalidAddress           = errorsmod.Register(ModuleName, 1013, "invalid address")
	ErrInvalidPubKey            = errorsmod.Register(ModuleName, 1014, "invalid pubkey")
	ErrInvalidCoins             = errorsmod.Register(ModuleName, 1015, "invalid coins")
	ErrInvalidType              = errorsmod.Register(ModuleName, 1016, "invalid type")
	ErrUnknownRequest           = errorsmod.Register(ModuleName, 1017, "unknown request")
	ErrInvalidRequest           = errorsmod.Register(ModuleName, 1018, "invalid request")
	ErrNoProposer               = errorsmod.Register(ModuleName, 1019, "no proposer found")
)
