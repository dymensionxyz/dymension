package types

import (
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var (
	ErrCanOnlyUpdatePendingPacket = gerrc.ErrFailedPrecondition.Wrap("can only update pending packet")
	ErrRollappPacketDoesNotExist  = gerrc.ErrNotFound.Wrap("rollapp packet")
	ErrRollappPacketAlreadyExists = gerrc.ErrAlreadyExists.Wrap("rollapp packet")
	ErrBadEIBCFee                 = gerrc.ErrInvalidArgument.Wrap("eibc fee")
)
