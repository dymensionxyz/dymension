package types

import (
	errorsmod "cosmossdk.io/errors"
)

// The following regiisters various lockdrop errors.
var (
	ErrNoGaugeIdExist                = errorsmod.Register(ModuleName, 1, "no gauge id exist")
	ErrDistrRecordNotPositiveWeight  = errorsmod.Register(ModuleName, 2, "weight in record should be non negative")
	ErrDistrRecordNotRegisteredGauge = errorsmod.Register(ModuleName, 3, "gauge was not registered")
	ErrDistrRecordRegisteredGauge    = errorsmod.Register(ModuleName, 4, "gauge was already registered")
	ErrDistrRecordNotSorted          = errorsmod.Register(ModuleName, 5, "gauges are not sorted")
	ErrDistrInfoNotPositiveWeight    = errorsmod.Register(ModuleName, 6, "total distribution weight should be positive")

	ErrEmptyProposalRecords         = errorsmod.Register(ModuleName, 10, "records are empty")
	ErrEmptyProposalGaugeIds        = errorsmod.Register(ModuleName, 11, "gauge ids are empty")
	ErrDistrInfoTotalWeightNotEqual = errorsmod.Register(ModuleName, 12, "total weight is not equal to sum of weights in records")

	ErrInvalidStreamStatus = errorsmod.Register(ModuleName, 20, "invalid stream status")
	ErrUnknownRequest      = errorsmod.Register(ModuleName, 21, "unknown request")
)
