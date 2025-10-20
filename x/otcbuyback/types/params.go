package types

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	common "github.com/dymensionxyz/dymension/v3/x/common/types"
	streamertypes "github.com/dymensionxyz/dymension/v3/x/streamer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// DefaultParams returns the default parameters for the Otcbuyback module
func DefaultParams() Params {
	return Params{
		MovingAverageSmoothingFactor: math.LegacyNewDecWithPrec(1, 1), // 0.1
		MaxPurchaseNumber:            20,                              // Maximum 20 purchases per user per auction
		MinPurchaseAmount:            math.ZeroInt(),                  // No minimum
		MinSoldDifferenceToPump:      common.DYM.MulRaw(1000),         // Minimum 1000 DYM sold to trigger pump
	}
}

// ValidateBasic performs basic validation on the Params
func (p Params) ValidateBasic() error {
	if p.MovingAverageSmoothingFactor.IsNil() ||
		p.MovingAverageSmoothingFactor.IsNegative() ||
		p.MovingAverageSmoothingFactor.GT(math.LegacyOneDec()) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "moving average smoothing factor must be between 0 and 1")
	}

	if p.MaxPurchaseNumber == 0 {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "max purchase number must be positive")
	}

	if p.MinPurchaseAmount.IsNil() || p.MinPurchaseAmount.IsNegative() {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "min purchase amount must be non-negative")
	}

	if p.MinSoldDifferenceToPump.IsNil() || p.MinSoldDifferenceToPump.IsNegative() {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "min sold difference to pump must be non-negative")
	}

	return nil
}

var DefaultPumpParams = Auction_PumpParams{
	PumpDelay:          time.Hour,           // Start a pump stream after 1 hour of creation
	PumpInterval:       15 * 24 * time.Hour, // Create pump streams every 15 days
	EpochIdentifier:    "month",
	NumEpochs:          2,
	NumOfPumpsPerEpoch: 25,
	PumpDistr:          streamertypes.PumpDistr_PUMP_DISTR_UNIFORM,
}
