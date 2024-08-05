package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func CmdTransferOwnership() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "transfer-ownership [rollapp-id] [new-owner]",
		Short:   "Transfer ownership of a rollapp to a new owner",
		Example: "dymd tx rollapp transfer-ownership ROLLAPP_CHAIN_ID <new_owner_address>",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// nolint:gofumpt
			argRollappId, newOwner := args[0], args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgTransferOwnership(
				clientCtx.GetFromAddress().String(),
				newOwner,
				argRollappId,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
