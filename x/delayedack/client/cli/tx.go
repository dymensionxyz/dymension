package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
)

// GetTxCmd returns the transaction commands for this module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdFinalizePacket())

	return cmd
}

func CmdFinalizePacket() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "finalize-packet [rollapp-id] [packet-type] [packet-src-channel] [packet-sequence] --from <voter>",
		Short:   "Finalize a specified packet",
		Example: "", // TODO
		Args:    cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.MsgFinalizePacket{
				Sender:           clientCtx.GetFromAddress().String(),
				RollappId:        "",
				PacketType:       "",
				PacketSrcChannel: "",
				PacketSequence:   "",
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
