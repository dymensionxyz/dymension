package types

import (
	"cosmossdk.io/math"
)

var (
	hundred = math.NewInt(100)

	DefaultMinAllocationWeight = math.NewInt(10) // 10%
	DefaultMinVotingPower      = math.NewInt(1)
)
