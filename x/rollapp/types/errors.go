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
)
