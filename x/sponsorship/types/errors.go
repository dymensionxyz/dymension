package types

import (
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var (
	ErrInvalidGaugeWeight  = gerrc.ErrInvalidArgument.Wrap("gauge weight")
	ErrInvalidDistribution = gerrc.ErrInvalidArgument.Wrap("gauge weight distribution")
	ErrInvalidParams       = gerrc.ErrInvalidArgument.Wrap("params")
	ErrInvalidGenesis      = gerrc.ErrInvalidArgument.Wrap("genesis")
	ErrInvalidVote         = gerrc.ErrInvalidArgument.Wrap("vote")
	ErrInvalidVoterInfo    = gerrc.ErrInvalidArgument.Wrap("voter info")
)
