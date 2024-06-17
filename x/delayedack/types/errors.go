package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrCanOnlyUpdatePendingPacket = errorsmod.Register(ModuleName, 1, "can only update pending packet")
	ErrRollappPacketDoesNotExist  = errorsmod.Register(ModuleName, 2, "rollapp packet does not exist")
	ErrUnknownRequest             = errorsmod.Register(ModuleName, 8, "unknown request")
)
