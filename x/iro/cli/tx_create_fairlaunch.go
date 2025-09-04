package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

const (
	FlagLiquidityDenom = "liquidity-denom"
)

func CmdCreateFairLaunchIRO() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-fairlaunch-iro [rollapp-id]",
		Short: "Create a new Fair Launch Initial RollApp Offering (IRO) plan",
		Long: `Create a new Fair Launch Initial RollApp Offering (IRO) plan for a specific RollApp.

Fair Launch IROs use global parameters defined in the module params

Parameters:
  [rollapp-id]: The unique identifier of the RollApp for which the Fair Launch IRO is being created.

Optional Flags:
  --liquidity-denom: The denomination to use for liquidity (e.g., "adym", "ibc/B3504E092456BA618CC28AC671A71FB08C6CA0FD0BE7C8A5B5A3E2DD933CC9E4"). Default: "adym"
  --trading-disabled: Disables trading for the plan initially. Will require MsgEnableTrading to be executed later.
                     Default: false (trading enabled)

Examples:
  dymd tx iro create-fairlaunch-iro myrollapp1 --from mykey
  dymd tx iro create-fairlaunch-iro myrollapp2 --liquidity-denom ibc/B3504E092456BA618CC28AC671A71FB08C6CA0FD0BE7C8A5B5A3E2DD933CC9E4 --from mykey
  dymd tx iro create-fairlaunch-iro myrollapp3 --trading-disabled --from mykey
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argRollappId := args[0]

			// Parse liquidity denom flag (optional, defaults to "adym")
			liquidityDenom, err := cmd.Flags().GetString(FlagLiquidityDenom)
			if err != nil {
				return err
			}

			// Parse trading disabled flag (optional)
			tradingDisabled, err := cmd.Flags().GetBool(FlagTradingDisabled)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.MsgCreateFairLaunchPlan{
				Owner:          clientCtx.GetFromAddress().String(),
				RollappId:      argRollappId,
				TradingEnabled: !tradingDisabled,
				LiquidityDenom: liquidityDenom,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String(FlagLiquidityDenom, "adym", "The denomination to use for liquidity (default: adym)")
	cmd.Flags().Bool(FlagTradingDisabled, false, "Whether trading should be disabled initially")

	return cmd
}
