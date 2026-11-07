package cli

import (
	"encoding/hex"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/dymensionxyz/dymension/v3/x/agent/types"
	"github.com/spf13/cobra"
)

func CmdReclaimValidationBounty() *cobra.Command {
	cmd := &cobra.Command{Use: "reclaim-validation-bounty [request-hash]", Short: "Reclaim an expired, unclaimed validation bounty", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := client.GetClientTxContext(cmd)
		if err != nil {
			return err
		}
		hash, err := hex.DecodeString(args[0])
		if err != nil {
			return err
		}
		return tx.GenerateOrBroadcastTxCLI(ctx, cmd.Flags(), types.NewMsgReclaimValidationBounty(ctx.GetFromAddress().String(), hash))
	}}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
