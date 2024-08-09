package types

import (
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// The following registers various lockdrop errors.
var (
	ErrDistrRecordNotPositiveWeight = gerrc.ErrOutOfRange.Wrap("weight in record should be non negative")
	ErrDistrRecordRegisteredGauge   = gerrc.ErrAlreadyExists.Wrap("gauge")
	ErrDistrRecordNotSorted         = gerrc.ErrInvalidArgument.Wrap("not sorted")
	ErrDistrInfoNotPositiveWeight   = gerrc.ErrOutOfRange.Wrap("total distribution weight should be positive")

	ErrInvalidStreamStatus = gerrc.ErrFailedPrecondition.Wrap("stream status")
	ErrUnknownRequest      = gerrc.ErrInvalidArgument.Wrap("request")
)
