package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
	"github.com/spf13/cobra"
)

func CmdClaim() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim [plan-id]",
		Short: "Claim tokens after the plan is settled",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			planID := args[0]

			msg := types.MsgClaim{
				Claimer: clientCtx.GetFromAddress().String(),
				PlanId:  planID,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
