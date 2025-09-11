package types

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
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
		p.MovingAverageSmoothingFactor.GTE(math.LegacyOneDec()) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "moving average smoothing factor must be between 0 and 1")
	}
	return nil
}

var DefaultVestingParams = Auction_VestingParams{
	VestingPeriod:               3 * 30 * 24 * time.Hour,
	VestingStartAfterAuctionEnd: 0,
}

var DefaultPumpParams = Auction_PumpParams{
	StartTimeAfterAuctionEnd: time.Hour,
	EpochIdentifier:          "month",
	NumEpochsPaidOver:        2,
	NumOfPumpsPerEpoch:       25,
}
