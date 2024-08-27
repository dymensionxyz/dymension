package types

// DONTCOVER

import (
	errorsmod "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/iro module sentinel errors
var (
	ErrRollappGenesisChecksumNotSet = errorsmod.Register(ModuleName, 1101, "rollapp genesis checksum not set")
	ErrRollappTokenSymbolNotSet     = errorsmod.Register(ModuleName, 1102, "rollapp token symbol not set")
	ErrRollappSealed                = errorsmod.Register(ModuleName, 1103, "rollapp is sealed")
	ErrPlanExists                   = errorsmod.Register(ModuleName, 1104, "plan already exists")
	ErrInvalidEndTime               = errorsmod.Register(ModuleName, 1105, "invalid end time")
)
