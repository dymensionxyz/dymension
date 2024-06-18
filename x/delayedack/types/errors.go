package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrCanOnlyUpdatePendingPacket = errorsmod.Register(ModuleName, 1, "can only update pending packet")
	ErrRollappPacketDoesNotExist  = errorsmod.Register(ModuleName, 2, "rollapp packet does not exist")
	ErrMismatchedStateRoots       = errorsmod.Register(ModuleName, 5, "mismatched state roots")
	ErrMismatchedSequencer        = errorsmod.Register(ModuleName, 6, "mismatched sequencer")
	ErrUnknownRequest             = errorsmod.Register(ModuleName, 8, "unknown request")
	ErrInvalidType                = errorsmod.Register(ModuleName, 9, "invalid type")
	ErrBadEIBCFee                 = errorsmod.Register(ModuleName, 10, "provided eibc fee is invalid")
)
