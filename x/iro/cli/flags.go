package cli

import (
	"time"

	"github.com/dymensionxyz/dymension/v3/x/iro/types"
	flag "github.com/spf13/pflag"
)

/*
	BondingCurve    BondingCurve                           `protobuf:"bytes,4,opt,name=bonding_curve,json=bondingCurve,proto3" json:"bonding_curve"`
	// The start time of the plan.
	StartTime time.Time `protobuf:"bytes,5,opt,name=start_time,json=startTime,proto3,stdtime" json:"start_time"`
	// The time before which the rollapp cannot be started.
	PreLaunchTime time.Time `protobuf:"bytes,6,opt,name=pre_launch_time,json=preLaunchTime,proto3,stdtime" json:"pre_launch_time"`
	// The incentive plan parameters for the tokens left after the plan is settled.
	IncentivePlanParams IncentivePlanParams `protobuf:"bytes,7,opt,name=incentive_plan_params,json=incentivePlanParams,proto3" json:"incentive_plan_params"`
}
*/

// Flags for streamer module tx commands.
const (
	FlagStartTime                              = "start-time"
	FlagBondingCurve                           = "curve"
	FlagIncentivesStartDurationAfterSettlement = "incentives-start"
	FlagIncentivesEpochs                       = "incentives-epochs"
)

var (
	defaultIncentivePlanParams_epochs = types.DefaultIncentivePlanMinimumNumEpochsPaidOver
	defaultIncentivePlanParams_start  = 7 * 24 * time.Hour
)

// FlagSetCreatePlan returns flags for creating gauges.
func FlagSetCreatePlan() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagStartTime, "", "The start time of the plan. Default is the current time.")
	fs.String(FlagBondingCurve, "", "The bonding curve parameters.")
	fs.Duration(FlagIncentivesStartDurationAfterSettlement, defaultIncentivePlanParams_start, "The duration after the plan is settled to start the incentives.")
	fs.Uint64(FlagIncentivesEpochs, defaultIncentivePlanParams_epochs, "The number of epochs for the incentives.")

	return fs
}
