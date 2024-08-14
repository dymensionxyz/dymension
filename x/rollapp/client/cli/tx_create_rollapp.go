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
		Use:     "create-rollapp [rollapp-id] [alias] [vm-type] [init-sequencer-address] [metadata] [bech32-prefix] [genesis_checksum]",
		Short:   "Create a new rollapp",
		Example: "dymd tx rollapp create-rollapp ROLLAPP_CHAIN_ID Rollapp EVM '<seq_address1>,<seq_address2>' metadata.json ethm <genesis_checksum>",
		Args:    cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			// nolint:gofumpt
			argRollappId, alias, vmTypeStr := args[0], args[1], args[2]

			vmType, ok := types.Rollapp_VMType_value[strings.ToUpper(vmTypeStr)]
			if !ok || vmType == 0 {
				return types.ErrInvalidVMType
			}

			var genesisChecksum, argInitSequencerAddress, argBech32Prefix string
			if len(args) > 3 {
				argInitSequencerAddress = args[3]
			}

			metadata := new(types.RollappMetadata)
			if len(args) > 4 {
				if err := utils.ParseJsonFromFile(args[4], metadata); err != nil {
					return err
				}
			}

			if len(args) > 5 {
				argBech32Prefix = args[5]
			}

			if len(args) > 6 {
				genesisChecksum = args[6]
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
				types.Rollapp_VMType(vmType),
				metadata,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetCreateRollapp())
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
