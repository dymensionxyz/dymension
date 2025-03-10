package cli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"cosmossdk.io/math"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

func CmdCreateIRO() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-iro [rollapp-id] [allocation] [duration]",
		Short: "Create a new Initial RollApp Offering (IRO) plan",
		Long: `Create a new Initial RollApp Offering (IRO) plan for a specific RollApp.

Parameters:
  [rollapp-id]  : The unique identifier of the RollApp for which the IRO is being created.
  [allocation]  : The total amount of tokens to be allocated for the IRO.
  [duration]    : The duration of the IRO plan (e.g., "24h", "30m", "1h30m").

Required Flags:
  --curve           : The bonding curve parameters in the format "M,N,C" where the curve is defined as p(x) = M * x^N + C.

Optional Flags:
  --start-time      : The time when the IRO will start. Can be Unix timestamp or RFC3339 format (e.g., "2023-10-01T00:00:00Z").
                      Default: Current time
  --incentives-start: The duration after settlement when incentives distribution starts.
                      Default: 7 days (168h)
  --incentives-epochs: The number of epochs over which incentives will be distributed (1 minute per epoch).
                      Default: 3000
  --liquidity-part  : The part of the total liquidity to allocate to the plan (0.0 to 1.0).
                      Default: 1.0
  --vesting-duration: The duration of the vesting period.
                      Default: 100 days (2400h)
  --vesting-start-time: The start time of the vesting period after settlement.
                      Default: 0m
  --trading-disabled: Disables trading for the plan. Will require MsgEnableTrading to be executed later on.
                      Default: false

Examples:
  dymd tx iro create-iro myrollapp1 1000000000 24h --curve "1.2,0.4,0" --from mykey
  dymd tx iro create-iro myrollapp2 500000000 30m --curve "1.5,0.5,100" --start-time "2023-10-01T00:00:00Z" --incentives-start 24h --incentives-epochs 3000 --from mykey
  dymd tx iro create-iro myrollapp3 2000000000 48h --curve "1.3,0.3,50" --trading-disabled=true --from mykey
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argRollappId := args[0]
			argAllocation := args[1]
			argDurationStr := args[2]

			allocationAmt, ok := math.NewIntFromString(argAllocation)
			if !ok {
				return fmt.Errorf("invalid allocation amount: %s", argAllocation)
			}

			// Parse the string into a time.Duration
			planDuration, err := time.ParseDuration(argDurationStr)
			if err != nil {
				return err
			}

			// Parse curve flag
			curveStr, err := cmd.Flags().GetString(FlagBondingCurve)
			if err != nil {
				return err
			}
			curve, err := ParseBondingCurve(curveStr)
			if err != nil {
				return errors.Join(types.ErrInvalidBondingCurve, err)
			}

			/* ----------------------------- optional flags ----------------------------- */
			tradingDisabled, err := cmd.Flags().GetBool(FlagTradingDisabled)
			if err != nil {
				return err
			}

			var startTime time.Time
			timeStr, err := cmd.Flags().GetString(FlagStartTime)
			if err != nil {
				return err
			}

			// If trading is disabled, start time should not be provided
			if tradingDisabled && timeStr != "" {
				return errors.New("start-time cannot be set when trading is disabled")
			}

			if timeStr == "" { // empty start time
				startTime = time.Unix(0, 0)
			} else if timeUnix, err := strconv.ParseInt(timeStr, 10, 64); err == nil { // unix time
				startTime = time.Unix(timeUnix, 0)
			} else if timeRFC, err := time.Parse(time.RFC3339, timeStr); err == nil { // RFC time
				startTime = timeRFC
			} else { // invalid input
				return errors.New("invalid start time format")
			}

			incentivesStart, err := cmd.Flags().GetDuration(FlagIncentivesStartDurationAfterSettlement)
			if err != nil {
				return err
			}

			incentivesEpochs, err := cmd.Flags().GetUint64(FlagIncentivesEpochs)
			if err != nil {
				return err
			}

			liquidityPart, err := cmd.Flags().GetFloat64(FlagLiquidityPart)
			if err != nil {
				return err
			}

			vestingDuration, err := cmd.Flags().GetDuration(FlagVestingDuration)
			if err != nil {
				return err
			}

			vestingStartTimeAfterSettlement, err := cmd.Flags().GetDuration(FlagVestingStartTimeAfterSettlement)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.MsgCreatePlan{
				Owner:           clientCtx.GetFromAddress().String(),
				RollappId:       argRollappId,
				AllocatedAmount: allocationAmt,
				BondingCurve:    curve,
				StartTime:       startTime,
				IroPlanDuration: planDuration,
				IncentivePlanParams: types.IncentivePlanParams{
					StartTimeAfterSettlement: incentivesStart,
					NumEpochsPaidOver:        incentivesEpochs,
				},
				LiquidityPart:                   math.LegacyMustNewDecFromStr(fmt.Sprintf("%f", liquidityPart)),
				VestingDuration:                 vestingDuration,
				VestingStartTimeAfterSettlement: vestingStartTimeAfterSettlement,
				TradingEnabled:                  !tradingDisabled,
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().AddFlagSet(FlagSetCreatePlan())

	return cmd
}

// ParseBondingCurve parses the bonding curve string into a BondingCurve struct
// expected format: "M,N,C" for p(x) = M * x^N + C
func ParseBondingCurve(curveStr string) (types.BondingCurve, error) {
	var curve types.BondingCurve

	curveParams := strings.Split(curveStr, ",")
	if len(curveParams) != 3 {
		return curve, errors.New("invalid bonding curve parameters")
	}

	M, err := math.LegacyNewDecFromStr(curveParams[0])
	if err != nil {
		return curve, errors.New("invalid M parameter")
	}

	N, err := math.LegacyNewDecFromStr(curveParams[1])
	if err != nil {
		return curve, errors.New("invalid N parameter")
	}

	C, err := math.LegacyNewDecFromStr(curveParams[2])
	if err != nil {
		return curve, errors.New("invalid C parameter")
	}

	curve = types.NewBondingCurve(M, N, C)
	return curve, curve.ValidateBasic()
}
