package cli

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/utils"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func CmdCreateRollapp() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create-rollapp [rollapp-id] [alias] [vm-type]",
		Short:   "Create a new rollapp",
		Example: "dymd tx rollapp create-rollapp ROLLAPP_CHAIN_ID Rollapp EVM --init-sequencer '<seq_address1>,<seq_address2>' --genesis-checksum <genesis_checksum> --metadata metadata.json",
		Args:    cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			// nolint:gofumpt
			argRollappId, alias, vmTypeStr := args[0], args[1], args[2]

			vmType, ok := types.Rollapp_VMType_value[strings.ToUpper(vmTypeStr)]
			if !ok || vmType == 0 {
				return types.ErrInvalidVMType
			}

			initSequencer, err := cmd.Flags().GetString(FlagInitSequencer)
			if err != nil {
				return err
			}

			genesisChecksum, err := cmd.Flags().GetString(FlagGenesisChecksum)
			if err != nil {
				return err
			}

			bech32Prefix, err := cmd.Flags().GetString(FlagBech32Prefix)
			if err != nil {
				return err
			}

			metadataFlag, err := cmd.Flags().GetString(FlagMetadata)
			if err != nil {
				return err
			}

			metadata := new(types.RollappMetadata)
			if metadataFlag != "" {
				if err = utils.ParseJsonFromFile(metadataFlag, metadata); err != nil {
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
				initSequencer,
				bech32Prefix,
				genesisChecksum,
				alias,
				types.Rollapp_VMType(vmType),
				metadata,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetUpdateRollapp())
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
