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
)

var _ = strconv.Itoa(0)

func CmdUpdateState() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-state [rollapp-id] [start-height] [num-blocks] [da-path] [version] [bds]",
		Short: "Update rollapp state",
		Args:  cobra.ExactArgs(6),
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
			argVersion, err := cast.ToUint64E(args[4])
			if err != nil {
				return err
			}
			argBDs := new(types.BlockDescriptors)
			err = json.Unmarshal([]byte(args[5]), argBDs)
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
				argStartHeight,
				argNumBlocks,
				argDAPath,
				argVersion,
				argBDs,
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
