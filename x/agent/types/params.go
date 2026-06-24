package types

import (
	"fmt"

	"cosmossdk.io/math"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
)

const DefaultMaxActionBytes = 100_000

func DefaultParams() Params {
	return Params{
		AgentRegistrationFee: commontypes.Dym(math.NewInt(1)),
		MaxActionBytes:       DefaultMaxActionBytes,
	}
}

func (p Params) Validate() error {
	if !p.AgentRegistrationFee.IsValid() {
		return fmt.Errorf("invalid agent registration fee: %s", p.AgentRegistrationFee)
	}
	if p.MaxActionBytes == 0 {
		return fmt.Errorf("max action bytes must be positive")
	}
	return nil
}
