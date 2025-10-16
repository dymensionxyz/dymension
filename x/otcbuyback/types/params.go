package types

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	streamertypes "github.com/dymensionxyz/dymension/v3/x/streamer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// DefaultParams returns the default parameters for the Otcbuyback module
func DefaultParams() Params {
	return Params{
		MovingAverageSmoothingFactor: math.LegacyNewDecWithPrec(1, 1), // 0.1
	}
}

// ValidateBasic performs basic validation on the Params
func (p Params) ValidateBasic() error {
	if p.MovingAverageSmoothingFactor.IsNil() ||
		p.MovingAverageSmoothingFactor.IsNegative() ||
		p.MovingAverageSmoothingFactor.GT(math.LegacyOneDec()) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "moving average smoothing factor must be between 0 and 1")
	}
	return nil
}

var DefaultVestingParams = Auction_VestingParams{
	VestingDelay: 0, // No delay by default
}

var DefaultPumpParams = Auction_PumpParams{
	PumpDelay:          0, // Start pumping immediately after auction start
	PumpInterval:       time.Hour,
	EpochIdentifier:    "month",
	NumEpochs:          2,
	NumOfPumpsPerEpoch: 25,
	PumpDistr:          streamertypes.PumpDistr_PUMP_DISTR_UNIFORM,
}
