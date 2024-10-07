package cli

import (
	"encoding/json"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

var _ = strconv.Itoa(0)

func CmdUpdateState() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-state [rollapp-id] [start-height] [num-blocks] [da-path] [bds]",
		Short: "Update rollapp state",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argRollappId := args[0]
			argStartHeight, err := cast.ToUint64E(args[1])
			if err != nil {
				return err
			}
			argNumBlocks, err := cast.ToUint64E(args[2])
			if err != nil {
				return err
			}
			argDAPath := args[3]
			argBDs := new(types.BlockDescriptors)
			err = json.Unmarshal([]byte(args[4]), argBDs)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdateState(
				clientCtx.GetFromAddress().String(),
				argRollappId,
				argDAPath,
				argStartHeight,
				argNumBlocks,
				argBDs,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
