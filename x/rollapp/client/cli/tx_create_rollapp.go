package cli

import (
	"encoding/json"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

// PermissionedAddresses .. TODO: refactor to be flag of []string
type PermissionedAddresses struct {
	Addresses []string `json:"addresses"`
}

func CmdCreateRollapp() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create-rollapp [rollapp-id] [max-sequencers] [permissioned-addresses] [metadata.json]",
		Short:   "Create a new rollapp",
		Example: "dymd tx rollapp create-rollapp ROLLAPP_CHAIN_ID 10 '{\"Addresses\":[]}' metadata.json",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argRollappId := args[0]

			argMaxSequencers, err := cast.ToUint64E(args[1])
			if err != nil {
				return err
			}
			var argPermissionedAddresses PermissionedAddresses
			err = json.Unmarshal([]byte(args[2]), &argPermissionedAddresses)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgCreateRollapp(clientCtx.GetFromAddress().String(), argRollappId, argMaxSequencers, argPermissionedAddresses.Addresses)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetCreateRollapp())
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
