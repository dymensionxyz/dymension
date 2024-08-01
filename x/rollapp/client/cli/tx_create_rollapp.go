package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/utils"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func CmdCreateRollapp() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create-rollapp [rollapp-id] [alias] [bech32-prefix] [init-sequencer-address] [genesis_checksum] [metadata]",
		Short:   "Create a new rollapp",
		Example: "dymd tx rollapp create-rollapp ROLLAPP_CHAIN_ID Rollapp ethm <seq_address> <genesis_checksum> metadata.json",
		Args:    cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			// nolint:gofumpt
			argRollappId, alias, argBech32Prefix := args[0], args[1], args[2]

			var genesisChecksum, argInitSequencerAddress string
			if len(args) > 3 {
				argInitSequencerAddress = args[3]
			}
			if len(args) > 4 {
				genesisChecksum = args[4]
			}

			metadata := new(types.RollappMetadata)
			if len(args) > 5 {
				if err := utils.ParseJsonFromFile(args[5], metadata); err != nil {
					return err
				}
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgCreateRollapp(
				clientCtx.GetFromAddress().String(),
				argRollappId,
				argInitSequencerAddress,
				argBech32Prefix,
				genesisChecksum,
				alias,
				metadata,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetCreateRollapp())
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
