package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func CmdAddApp() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add-app [name] [rollapp-id] [description] [logo] [url] [order]",
		Short:   "Add a new app",
		Example: "dymd tx app create-app 'app1' 'rollapp_1234-1' 1 'A description' '/logos/apps/app1.jpeg' 'https://app1.com/'",
		Args:    cobra.MinimumNArgs(5),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			var (
				name              = args[0]
				rollappId         = args[1]
				description       = args[2]
				logo              = args[3]
				url               = args[4]
				order       int64 = -1
			)

			if len(args) == 6 {
				order, err = strconv.ParseInt(args[5], 10, 32)
				if err != nil {
					return err
				}
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgAddApp(
				clientCtx.GetFromAddress().String(),
				name,
				rollappId,
				description,
				logo,
				url,
				int32(order),
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
