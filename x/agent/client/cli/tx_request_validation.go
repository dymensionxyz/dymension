package cli

import (
	"encoding/hex"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/agent/types"
	"github.com/spf13/cobra"
)

const flagBounty = "bounty"

func CmdRequestValidation() *cobra.Command {
	cmd := &cobra.Command{Use: "request-validation [validator-id] [agent-id] [evidence-seq] [request-hash] [request-uri]", Short: "Request independent validation of an agent action", Args: cobra.ExactArgs(5), RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := client.GetClientTxContext(cmd)
		if err != nil {
			return err
		}
		bounty, err := cmd.Flags().GetString(flagBounty)
		if err != nil {
			return err
		}
		msg, err := newMsgRequestValidation(ctx.GetFromAddress().String(), args, bounty)
		if err != nil {
			return err
		}
		return tx.GenerateOrBroadcastTxCLI(ctx, cmd.Flags(), msg)
	}}
	cmd.Flags().String(flagBounty, "", "optional bounty, e.g. 1000adym, paid to the validator on its first attested verdict")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func newMsgRequestValidation(requester string, args []string, bounty string) (*types.MsgRequestValidation, error) {
	seq, err := strconv.ParseUint(args[2], 10, 64)
	if err != nil {
		return nil, err
	}
	hash, err := hex.DecodeString(args[3])
	if err != nil {
		return nil, err
	}
	msg := types.NewMsgRequestValidation(requester, args[0], args[1], seq, hash, args[4])
	if bounty != "" {
		msg.Bounty, err = sdk.ParseCoinNormalized(bounty)
		if err != nil {
			return nil, err
		}
	}
	return msg, nil
}
