package cli

import (
	"time"

	flag "github.com/spf13/pflag"

	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

// Flags for streamer module tx commands.
const (
	FlagStartTime                              = "start-time"
	FlagBondingCurve                           = "curve"
	FlagIncentivesStartDurationAfterSettlement = "incentives-start"
	FlagIncentivesEpochs                       = "incentives-epochs"
	FlagLiquidityPart                          = "liquidity-part"
	FlagVestingDuration                        = "vesting-duration"
	FlagVestingStartTimeAfterSettlement        = "vesting-start-time"
)

var (
	defaultIncentivePlanParams_epochs = types.DefaultIncentivePlanMinimumNumEpochsPaidOver
	defaultIncentivePlanParams_start  = 7 * 24 * time.Hour
	defaultLiquidityPart              = float64(1.0)
	defaultVestingDuration            = 100 * 24 * time.Hour
	defaultVestingStartTime           = 0 * time.Minute
)

// FlagSetCreatePlan returns flags for creating gauges.
func FlagSetCreatePlan() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagStartTime, "", "The start time of the plan. Default is the current time.")
	fs.String(FlagBondingCurve, "", "The bonding curve parameters.")
	fs.Duration(FlagIncentivesStartDurationAfterSettlement, defaultIncentivePlanParams_start, "The duration after the plan is settled to start the incentives.")
	fs.Uint64(FlagIncentivesEpochs, defaultIncentivePlanParams_epochs, "The number of epochs for the incentives.")
	fs.Float64(FlagLiquidityPart, defaultLiquidityPart, "The part of the total liquidity to allocate to the plan.")
	fs.Duration(FlagVestingDuration, defaultVestingDuration, "The duration of the vesting period.")
	fs.Duration(FlagVestingStartTimeAfterSettlement, defaultVestingStartTime, "The start time of the vesting period after the plan is settled.")

	return fs
}
