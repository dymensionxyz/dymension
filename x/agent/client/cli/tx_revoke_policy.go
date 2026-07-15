package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

func CmdRevokePolicy() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revoke-policy [fingerprint] [reason]",
		Short: "Revoke a TEE policy by fingerprint (governance authority only)",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			reason := ""
			if len(args) > 1 {
				reason = args[1]
			}
			msg := types.NewMsgRevokePolicy(clientCtx.GetFromAddress().String(), args[0], reason)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdUnrevokePolicy() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unrevoke-policy [fingerprint]",
		Short: "Remove a TEE policy fingerprint from the revocation denylist (governance authority only)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgUnrevokePolicy(clientCtx.GetFromAddress().String(), args[0])
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
