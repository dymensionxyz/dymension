package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/rollapp module sentinel errors
var (
	ErrRollappExists                  = sdkerrors.Register(ModuleName, 1000, "rollapp already exist for this rollapp-id; must use new rollapp-id")
	ErrInvalidwMaxSequencers          = sdkerrors.Register(ModuleName, 1001, "invalid max sequencers")
	ErrInvalidwMaxWithholding         = sdkerrors.Register(ModuleName, 1002, "invalid max withholding")
	ErrInvalidPermissionedAddress     = sdkerrors.Register(ModuleName, 1003, "invalid permissioned address")
	ErrPermissionedAddressesDuplicate = sdkerrors.Register(ModuleName, 1004, "permissioned-address has duplicates")
	ErrInvalidNumBlocks               = sdkerrors.Register(ModuleName, 1005, "invalid number of blocks")
	ErrInvalidBlockSequence           = sdkerrors.Register(ModuleName, 1006, "invalid block sequence")
	ErrUnknownRollappId               = sdkerrors.Register(ModuleName, 1007, "rollapp does not exist")
	ErrVersionMismatch                = sdkerrors.Register(ModuleName, 1008, "rollapp version mismatch")
	ErrWrongBlockHeight               = sdkerrors.Register(ModuleName, 1009, "start-height does not match rollapps state")
)
