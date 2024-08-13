package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/utils"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func CmdUpdateRollapp() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update-rollapp [rollapp-id] [init-sequencer] [genesis_checksum] [bech32-prefix] [metadata] ",
		Short:   "Update a new rollapp",
		Example: "dymd tx rollapp update-rollapp ROLLAPP_CHAIN_ID --init-sequencer '<seq_address1>,<seq_address2>' --genesis-checksum <genesis_checksum> --metadata metadata.json",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argRollappId := args[0]

			initSequencer, err := cmd.Flags().GetString(FlagInitSequencer)
			if err != nil {
				return
			}

			genesisChecksum, err := cmd.Flags().GetString(FlagGenesisChecksum)
			if err != nil {
				return
			}

			bech32Prefix, err := cmd.Flags().GetString(FlagBech32Prefix)
			if err != nil {
				return
			}

			metadataFlag, err := cmd.Flags().GetString(FlagMetadata)
			if err != nil {
				return
			}

			metadata := new(types.RollappMetadata)
			if metadataFlag != "" {
				if err = utils.ParseJsonFromFile(metadataFlag, metadata); err != nil {
					return
				}
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return
			}

			msg := types.NewMsgUpdateRollappInformation(
				clientCtx.GetFromAddress().String(),
				argRollappId,
				initSequencer,
				genesisChecksum,
				metadata,
				bech32Prefix,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetUpdateRollapp())
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
