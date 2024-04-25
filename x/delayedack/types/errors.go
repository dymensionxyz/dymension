package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrCanOnlyUpdatePendingPacket = errorsmod.Register(ModuleName, 1, "can only update pending packet")
	ErrRollappPacketDoesNotExist  = errorsmod.Register(ModuleName, 2, "rollapp packet does not exist")
	ErrEmptyEpochIdentifier       = errorsmod.Register(ModuleName, 4, "empty epoch identifier")
	ErrMismatchedStateRoots       = errorsmod.Register(ModuleName, 5, "mismatched state roots")
	ErrMismatchedSequencer        = errorsmod.Register(ModuleName, 6, "mismatched sequencer")
	ErrMissingEIBCMetadata        = errorsmod.Register(ModuleName, 7, "missing eibc metadata")
)
