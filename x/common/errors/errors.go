package errors

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/dymension/v3/utils/gerr"
)

// This file should contain ubiquitous domain specific errors which warrant their own handling on top of gerr handling
// For example, if your caller code wants to differentiate between a generic failed precondition, and a failed precondition due to
// misbehavior, you would define a misbehavior error here.

var (
	ErrProtocolViolation          = errorsmod.Wrap(gerr.ErrFailedPrecondition, "actor is violating protocol")
	ErrSequencerProtocolViolation = errorsmod.Wrap(ErrProtocolViolation, "sequencer")
)
