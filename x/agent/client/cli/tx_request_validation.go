package cli

import (
	"encoding/hex"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/dymensionxyz/dymension/v3/x/agent/types"
	"github.com/spf13/cobra"
)

func CmdRequestValidation() *cobra.Command {
	cmd := &cobra.Command{Use: "request-validation [validator-id] [agent-id] [evidence-seq] [request-hash] [request-uri]", Short: "Request independent validation of an agent action", Args: cobra.ExactArgs(5), RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := client.GetClientTxContext(cmd)
		if err != nil {
			return err
		}
		seq, err := strconv.ParseUint(args[2], 10, 64)
		if err != nil {
			return err
		}
		hash, err := hex.DecodeString(args[3])
		if err != nil {
			return err
		}
		return tx.GenerateOrBroadcastTxCLI(ctx, cmd.Flags(), types.NewMsgRequestValidation(ctx.GetFromAddress().String(), args[0], args[1], seq, hash, args[4]))
	}}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
