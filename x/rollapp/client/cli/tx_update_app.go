package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func CmdUpdateApp() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update-app [id] [name] [rollapp-id] [description] [logo] [url] [order]",
		Short:   "Update an app",
		Example: "dymd tx rollapp update-app 1 'app1' 'rollapp_1234-1' 'A description' '/logos/apps/app1.jpeg' 'https://app1.com/' 3",
		Args:    cobra.MinimumNArgs(6),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			var (
				name              = args[1]
				rollappId         = args[2]
				description       = args[3]
				logo              = args[4]
				url               = args[5]
				order       int64 = -1
			)

			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			if len(args) == 7 {
				order, err = strconv.ParseInt(args[6], 10, 32)
				if err != nil {
					return err
				}
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdateApp(
				clientCtx.GetFromAddress().String(),
				id,
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
