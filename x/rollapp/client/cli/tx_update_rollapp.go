package cli

import (
	"github.com/cometbft/cometbft/libs/json"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func CmdUpdateRollapp() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update-rollapp [rollapp-id] [init-sequencer] [genesis_checksum] [alias] [metadata]",
		Short:   "Update a new rollapp",
		Example: "dymd tx rollapp update-rollapp ROLLAPP_CHAIN_ID --init-sequencer <seq_address> --genesis-checksum <genesis_checksum> --alias Rollapp --metadata metadata.json",
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argRollappId := args[0]

			initSequencerAddress, err := cmd.Flags().GetString("init-sequencer")
			if err != nil {
				return
			}

			genesisChecksum, err := cmd.Flags().GetString("genesis-checksum")
			if err != nil {
				return
			}

			alias, err := cmd.Flags().GetString("alias")
			if err != nil {
				return
			}

			metadataFlag, err := cmd.Flags().GetString("metadata")
			if err != nil {
				return
			}

			metadata := new(types.RollappMetadata)
			if err = json.Unmarshal([]byte(metadataFlag), metadata); err != nil {
				return
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return
			}

			msg := types.NewMsgUpdateRollappInformation(
				clientCtx.GetFromAddress().String(),
				argRollappId,
				initSequencerAddress,
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
