package cli

import (
    "strconv"
	
	"github.com/spf13/cobra"
    "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

var _ = strconv.Itoa(0)

func CmdSubmitFraud() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-fraud [rollapp-id]",
		Short: "Broadcast message SubmitFraud",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
      		 argRollappID := args[0]
            
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgSubmitFraud(
				clientCtx.GetFromAddress().String(),
				argRollappID,
				
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

    return cmd
}