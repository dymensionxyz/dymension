package types

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// The following regiisters various lockdrop errors.
var (
	ErrDistrRecordNotPositiveWeight = errorsmod.Register(ModuleName, 2, "weight in record should be non negative")
	ErrDistrRecordRegisteredGauge   = errorsmod.Register(ModuleName, 4, "gauge was already registered")
	ErrDistrRecordNotSorted         = errorsmod.Register(ModuleName, 5, "gauges are not sorted")
	ErrDistrInfoNotPositiveWeight   = errorsmod.Register(ModuleName, 6, "total distribution weight should be positive")

	ErrInvalidStreamStatus = gerrc.ErrInvalidArgument.Wrap("stream status")
	ErrUnknownRequest      = errorsmod.Register(ModuleName, 21, "unknown request")
)
