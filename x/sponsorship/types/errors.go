package types

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var (
	ErrInvalidGaugeWeight  = errorsmod.Wrap(gerrc.ErrInvalidArgument, "invalid gauge weight")
	ErrInvalidDistribution = errorsmod.Wrap(gerrc.ErrInvalidArgument, "invalid gauge weight distribution")
	ErrInvalidParams       = errorsmod.Wrap(gerrc.ErrInvalidArgument, "invalid params")
	ErrInvalidGenesis      = errorsmod.Wrap(gerrc.ErrInvalidArgument, "invalid genesis")
	ErrInvalidVote         = errorsmod.Wrap(gerrc.ErrInvalidArgument, "invalid vote")
	ErrInvalidVoterInfo    = errorsmod.Wrap(gerrc.ErrInvalidArgument, "invalid voter info")
	ErrNoEndorsers         = errorsmod.Wrap(gerrc.ErrFailedPrecondition, "no endorsers")
)
