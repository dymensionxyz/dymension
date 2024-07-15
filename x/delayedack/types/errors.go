package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrCanOnlyUpdatePendingPacket = errorsmod.Register(ModuleName, 1, "can only update pending packet")
	ErrRollappPacketDoesNotExist  = errorsmod.Register(ModuleName, 2, "rollapp packet does not exist")
	ErrRollappPacketAlreadyExists = errorsmod.Register(ModuleName, 3, "rollapp packet already exists")
	ErrUnknownRequest             = errorsmod.Register(ModuleName, 8, "unknown request")
	ErrBadEIBCFee                 = errorsmod.Register(ModuleName, 10, "provided eibc fee is invalid")
)
