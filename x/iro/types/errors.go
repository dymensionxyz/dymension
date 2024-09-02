package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

// x/iro module sentinel errors
var (
	ErrRollappGenesisChecksumNotSet = errorsmod.Register(ModuleName, 1101, "rollapp genesis checksum not set")
	ErrRollappTokenSymbolNotSet     = errorsmod.Register(ModuleName, 1102, "rollapp token symbol not set")
	ErrRollappSealed                = errorsmod.Register(ModuleName, 1103, "rollapp is sealed")
	ErrPlanExists                   = errorsmod.Register(ModuleName, 1104, "plan already exists")
	ErrInvalidEndTime               = errorsmod.Register(ModuleName, 1105, "invalid end time")
	ErrPlanSettled                  = errorsmod.Register(ModuleName, 1106, "plan is settled")
	ErrPlanNotStarted               = errorsmod.Register(ModuleName, 1107, "plan has not started")
	ErrPlanNotFound                 = errorsmod.Register(ModuleName, 1108, "plan not found")
	ErrInvalidExpectedOutAmount     = errorsmod.Register(ModuleName, 1109, "invalid expected out amount")
	ErrInvalidMinCost               = errorsmod.Register(ModuleName, 1110, "invalid minimum cost")
	ErrInvalidBondingCurve          = errorsmod.Register(ModuleName, 1111, "invalid bonding curve params")
	ErrInvalidRollappGenesisState   = errorsmod.Register(ModuleName, 1112, "invalid rollapp genesis state")
	ErrInvalidCost                  = errorsmod.Register(ModuleName, 1113, "invalid cost")
	ErrInvalidAllocation            = errorsmod.Register(ModuleName, 1114, "invalid allocation")
	ErrFailedBootstrapLiquidityPool = errorsmod.Register(ModuleName, 1115, "failed to bootstrap liquidity pool")
	ErrPlanNotSettled               = errorsmod.Register(ModuleName, 1116, "plan is not settled")
	ErrNoTokensToClaim              = errorsmod.Register(ModuleName, 1117, "no tokens to claim")
	ErrInsufficientTokens           = errorsmod.Register(ModuleName, 1118, "insufficient tokens")
)
