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

func CmdRespondValidationAttested() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "respond-validation-attested [request-hash] [response] [response-uri] [response-hash] [tag] [token]",
		Short: "Respond to a validation request with an enclave-attested verdict",
		Long:  "Submit an attested verdict. Hashes are hex encoded; token is the raw attestation token (not base64 encoded).",
		Args:  cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			msg, err := newMsgRespondValidationAttested(ctx.GetFromAddress().String(), args)
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(ctx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func newMsgRespondValidationAttested(responder string, args []string) (*types.MsgRespondValidationAttested, error) {
	hash, err := hex.DecodeString(args[0])
	if err != nil {
		return nil, err
	}
	score, err := strconv.ParseUint(args[1], 10, 32)
	if err != nil {
		return nil, err
	}
	responseHash, err := hex.DecodeString(args[3])
	if err != nil {
		return nil, err
	}
	msg := &types.MsgRespondValidationAttested{Responder: responder, RequestHash: hash, Response: uint32(score), ResponseUri: args[2], ResponseHash: responseHash, Tag: args[4], Token: []byte(args[5])}
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	return msg, nil
}
