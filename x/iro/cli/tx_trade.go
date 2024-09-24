package cli

import (
	"fmt"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

func CmdBuy() *cobra.Command {
	return createBuySellCmd(
		"buy [plan-id] [amount] [expected-out-amount]",
		"Buy allocation from an IRO plan",
		true,
	)
}

func CmdSell() *cobra.Command {
	return createBuySellCmd(
		"sell [plan-id] [amount] [expected-out-amount]",
		"Sell allocation from an IRO plan",
		false,
	)
}

func createBuySellCmd(use string, short string, isBuy bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			planID := args[0]
			argAmount := args[1]
			argExpectedAmount := args[2]

			amount, ok := math.NewIntFromString(argAmount)
			if !ok {
				return fmt.Errorf("invalid amount: %s", argAmount)
			}

			expectedAmount, ok := math.NewIntFromString(argExpectedAmount)
			if !ok {
				return fmt.Errorf("invalid expected out amount: %s", argExpectedAmount)
			}

			var msg sdk.Msg
			if isBuy {
				msg = &types.MsgBuy{
					Buyer:         clientCtx.GetFromAddress().String(),
					PlanId:        planID,
					Amount:        amount,
					MaxCostAmount: expectedAmount,
				}
			} else {
				msg = &types.MsgSell{
					Seller:          clientCtx.GetFromAddress().String(),
					PlanId:          planID,
					Amount:          amount,
					MinIncomeAmount: expectedAmount,
				}
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
