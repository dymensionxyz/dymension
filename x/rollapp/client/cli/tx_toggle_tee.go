package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func CmdToggleTEE() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "toggle-tee [rollapp-id] [enable]",
		Short:   "Toggle the TEE feature for a rollapp",
		Example: "dymd tx rollapp toggle-tee ROLLAPP_CHAIN_ID true",
		Args:    cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			argRollappId := args[0]
			enable := false
			if len(args) > 1 {
				var err error
				enable, err = strconv.ParseBool(args[1])
				if err != nil {
					return err
				}
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgToggleTEE(
				clientCtx.GetFromAddress().String(),
				argRollappId,
				enable,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
