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
		Use:     "create-rollapp [rollapp-id] [alias] [bech32-prefix] [vm-type] [init-sequencer-address] [genesis_checksum] [metadata]",
		Short:   "Create a new rollapp",
		Example: "dymd tx rollapp create-rollapp ROLLAPP_CHAIN_ID Rollapp ethm EVM '<seq_address1>,<seq_address2>' <genesis_checksum> metadata.json",
		Args:    cobra.MinimumNArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			// nolint:gofumpt
			argRollappId, alias, argBech32Prefix, vmTypeStr := args[0], args[1], args[2], args[3]

			vmType, ok := types.Rollapp_VMType_value[strings.ToUpper(vmTypeStr)]
			if !ok || vmType == 0 {
				return types.ErrInvalidVMType
			}

			var genesisChecksum, argInitSequencerAddress string
			if len(args) > 4 {
				argInitSequencerAddress = args[4]
			}
			if len(args) > 5 {
				genesisChecksum = args[5]
			}

			metadata := new(types.RollappMetadata)
			if len(args) > 6 {
				if err := utils.ParseJsonFromFile(args[6], metadata); err != nil {
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
