package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrInvalidGaugeWeight  = errorsmod.Register(ModuleName, 1, "invalid gauge weight")
	ErrInvalidDistribution = errorsmod.Register(ModuleName, 2, "invalid gauge weight distribution")
	ErrInvalidParams       = errorsmod.Register(ModuleName, 3, "invalid params")
	ErrInvalidGenesis      = errorsmod.Register(ModuleName, 4, "invalid genesis")
	ErrInvalidVote         = errorsmod.Register(ModuleName, 5, "invalid vote")
	ErrInvalidVoterInfo    = errorsmod.Register(ModuleName, 6, "invalid voter info")
)
