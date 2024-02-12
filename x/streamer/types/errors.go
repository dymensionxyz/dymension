package types

import (
	sdkerrors "cosmossdk.io/errors"
)

// The following regiisters various lockdrop errors.
var (
	ErrNoGaugeIdExist                = sdkerrors.Register(ModuleName, 1, "no gauge id exist")
	ErrDistrRecordNotPositiveWeight  = sdkerrors.Register(ModuleName, 2, "weight in record should be non negative")
	ErrDistrRecordNotRegisteredGauge = sdkerrors.Register(ModuleName, 3, "gauge was not registered")
	ErrDistrRecordRegisteredGauge    = sdkerrors.Register(ModuleName, 4, "gauge was already registered")
	ErrDistrRecordNotSorted          = sdkerrors.Register(ModuleName, 5, "gauges are not sorted")
	ErrDistrInfoNotPositiveWeight    = sdkerrors.Register(ModuleName, 6, "total distribution weight should be positive")

	ErrEmptyProposalRecords         = sdkerrors.Register(ModuleName, 10, "records are empty")
	ErrEmptyProposalGaugeIds        = sdkerrors.Register(ModuleName, 11, "gauge ids are empty")
	ErrDistrInfoTotalWeightNotEqual = sdkerrors.Register(ModuleName, 12, "total weight is not equal to sum of weights in records")

	ErrInvalidStreamStatus = sdkerrors.Register(ModuleName, 20, "invalid stream status")
)
