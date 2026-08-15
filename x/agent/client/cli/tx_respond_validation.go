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

func CmdRespondValidation() *cobra.Command {
	cmd := &cobra.Command{Use: "respond-validation [request-hash] [response] [response-uri] [response-hash] [tag]", Short: "Respond to a validation request", Args: cobra.ExactArgs(5), RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := client.GetClientTxContext(cmd)
		if err != nil {
			return err
		}
		hash, err := hex.DecodeString(args[0])
		if err != nil {
			return err
		}
		score, err := strconv.ParseUint(args[1], 10, 32)
		if err != nil {
			return err
		}
		var responseHash []byte
		if args[3] != "" {
			responseHash, err = hex.DecodeString(args[3])
			if err != nil {
				return err
			}
		}
		return tx.GenerateOrBroadcastTxCLI(ctx, cmd.Flags(), types.NewMsgRespondValidation(ctx.GetFromAddress().String(), hash, uint32(score), args[2], responseHash, args[4]))
	}}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
