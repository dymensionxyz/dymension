package cli

import (
	"strconv"

	"encoding/json"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	shared "github.com/dymensionxyz/dymension/shared/types"
)

var _ = strconv.Itoa(0)

func CmdCreateRollapp() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-rollapp [rollapp-id] [code-stamp] [genesis-path] [max-withholding-blocks] [max-sequencers] [permissioned-addresses]",
		Short: "Create a new rollapp",
		Args:  cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argRollappId := args[0]
			argCodeStamp := args[1]
			argGenesisPath := args[2]
			argMaxWithholdingBlocks, err := cast.ToUint64E(args[3])
			if err != nil {
				return err
			}
			argMaxSequencers, err := cast.ToUint64E(args[4])
			if err != nil {
				return err
			}
			argPermissionedAddresses := new(shared.Sequencers)
			err = json.Unmarshal([]byte(args[5]), argPermissionedAddresses)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgCreateRollapp(
				clientCtx.GetFromAddress().String(),
				argRollappId,
				argCodeStamp,
				argGenesisPath,
				argMaxWithholdingBlocks,
				argMaxSequencers,
				argPermissionedAddresses,
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
