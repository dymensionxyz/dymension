package cli

import (
	"os"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

func CmdSubmitWrongCommitment() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-wrongcommitment [rollapp-id] [SLIndex] [DA Path] [inclusionproof.json]",
		Short: "Broadcast message SubmitWrongCommitment",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argRollappID := args[0]
			slIndex, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}
			daPath := args[2]
			path := args[3]

			fileContent, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgWrongCommitmentBatch(
				clientCtx.GetFromAddress().String(),
				argRollappID,
				slIndex,
				daPath,
				string(fileContent),
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
