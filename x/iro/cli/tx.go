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

	// "github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdCreateIRO())

	return cmd
}

func CmdCreateIRO() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-iro [rollappId] [allocation] [pre-launch-time] --curve [curve params]",
		Short: "Create a new IRO plan",
		Example: `dymd create-iro [rollappId] [allocation] [pre-launch-time] --curve "1.2,0.4,0"
        Optional:
        --start-time [start-time]
        --incentives-start [incentives-start]
        --incentives-epochs [incentives-epochs]
        `,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argRollappId := args[0]
			argAllocation := args[1]
			argPreLaunchTimeStr := args[2]

			allocationAmt, ok := math.NewIntFromString(argAllocation)
			if !ok {
				return fmt.Errorf("invalid allocation amount: %s", argAllocation)
			}

			var preLaunchTime time.Time
			if argPreLaunchTimeStr == "" {
				return errors.New("pre-launch time cannot be empty")
			} else if timeUnix, err := strconv.ParseInt(argPreLaunchTimeStr, 10, 64); err == nil { // unix time
				preLaunchTime = time.Unix(timeUnix, 0)
			} else if timeRFC, err := time.Parse(time.RFC3339, argPreLaunchTimeStr); err == nil { // RFC time
				preLaunchTime = timeRFC
			} else { // invalid input
				return errors.New("invalid start time format")
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
			var startTime time.Time
			timeStr, err := cmd.Flags().GetString(FlagStartTime)
			if err != nil {
				return err
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
				PreLaunchTime:   preLaunchTime,
				IncentivePlanParams: types.IncentivePlanParams{
					StartTimeAfterSettlement: incentivesStart,
					NumEpochsPaidOver:        incentivesEpochs,
				},
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

// buy
// sell
// claim
