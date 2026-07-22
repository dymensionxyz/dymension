package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

const flagTag2 = "tag2"

func CmdSubmitFeedback() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-feedback [agent-id] [score] [tag1] [evidence-seq]",
		Short: "Submit or overwrite your feedback score for an agent",
		Long: "Submit feedback for an agent. score is fixed-point with 2 decimals (0..10000 == 0.00%..100.00%). " +
			"tag1 names the rated dimension, e.g. liveness. evidence-seq must reference an existing action-log entry of the agent.",
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			score, err := strconv.ParseUint(args[1], 10, 32)
			if err != nil {
				return err
			}
			evidenceSeq, err := strconv.ParseUint(args[3], 10, 64)
			if err != nil {
				return err
			}
			tag2, err := cmd.Flags().GetString(flagTag2)
			if err != nil {
				return err
			}

			msg := types.NewMsgSubmitFeedback(clientCtx.GetFromAddress().String(), args[0], uint32(score), args[2], tag2, evidenceSeq)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagTag2, "", "optional secondary dimension tag")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
