package types

import (
	"cosmossdk.io/math"
)

var (
	hundred = math.NewInt(100)

	DefaultMinAllocationWeight = math.NewInt(1) // 1%
	DefaultMinVotingPower      = math.NewInt(1) // 1 DYM
)
