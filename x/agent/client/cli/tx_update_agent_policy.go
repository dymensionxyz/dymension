package cli

import (
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/agent/types"
	"github.com/dymensionxyz/dymension/v3/x/common/tee"
)

func CmdUpdateAgentPolicy() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-agent-policy [agent-id] [policy-json-file]",
		Short: "Schedule a timelocked rotation of an agent's attestation policy",
		Long:  "Schedule a timelocked policy rotation. policy-json-file is a JSON file holding the new tee.Policy (gcp_root_cert_pem, policy_values, policy_query, policy_structure). The rotation activates after policy_rotation_delay_blocks.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			bz, err := os.ReadFile(args[1])
			if err != nil {
				return err
			}
			var policy tee.Policy
			if err := clientCtx.Codec.UnmarshalJSON(bz, &policy); err != nil {
				return err
			}

			msg := types.NewMsgUpdateAgentPolicy(clientCtx.GetFromAddress().String(), args[0], policy)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
