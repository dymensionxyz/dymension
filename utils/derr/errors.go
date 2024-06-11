package derr

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/dymension/v3/utils/gerr"
)

var (
	ErrFraud                            = errorsmod.Wrap(gerr.ErrFailedPrecondition, "actor is violating protocol")
	ErrViolatesDymensionRollappStandard = errorsmod.Wrap(ErrFraud, "rollapp does not meet dymension rollapp standard")
)
