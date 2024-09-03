package cli

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/utils"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func CmdCreateRollapp() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create-rollapp [rollapp-id] [alias] [vm-type]",
		Short:   "Create a new rollapp",
		Example: "dymd tx rollapp create-rollapp ROLLAPP_CHAIN_ID Rollapp EVM --init-sequencer '<seq_address1>,<seq_address2>' --genesis-checksum <genesis_checksum> --initial-supply 1000arax --native-denom native_denom.json --metadata metadata.json",
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

			genesisInfo, err := parseGenesisInfo(cmd)
			if err != nil {
				return err
			}

			metadata, err := parseMetadata(cmd)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgCreateRollapp(
				clientCtx.GetFromAddress().String(),
				argRollappId,
				initSequencer,
				alias,
				types.Rollapp_VMType(vmType),
				metadata,
				genesisInfo,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetUpdateRollapp())
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func parseGenesisInfo(cmd *cobra.Command) (types.GenesisInfo, error) {
	var (
		genesisInfo types.GenesisInfo
		err         error
	)

	genesisInfo.GenesisChecksum, err = cmd.Flags().GetString(FlagGenesisChecksum)
	if err != nil {
		return types.GenesisInfo{}, err
	}

	genesisInfo.Bech32Prefix, err = cmd.Flags().GetString(FlagBech32Prefix)
	if err != nil {
		return types.GenesisInfo{}, err
	}

	nativeDenomFlag, err := cmd.Flags().GetString(FlagNativeDenom)
	if err != nil {
		return types.GenesisInfo{}, err
	}

	genesisInfo.NativeDenom = new(types.DenomMetadata)
	if nativeDenomFlag != "" {
		if err = utils.ParseJsonFromFile(nativeDenomFlag, genesisInfo.NativeDenom); err != nil {
			return types.GenesisInfo{}, err
		}
	}

	initialSupplyFlag, err := cmd.Flags().GetString(FlagInitialSupply)
	if err != nil {
		return types.GenesisInfo{}, err
	}

	if initialSupplyFlag != "" {
		genesisInfo.InitialSupply, err = sdk.ParseCoinNormalized(initialSupplyFlag)
		if err != nil {
			return types.GenesisInfo{}, err
		}
	}

	return genesisInfo, nil
}

func parseMetadata(cmd *cobra.Command) (*types.RollappMetadata, error) {
	metadataFlag, err := cmd.Flags().GetString(FlagMetadata)
	if err != nil {
		return nil, err
	}

	metadata := new(types.RollappMetadata)
	if metadataFlag != "" {
		if err = utils.ParseJsonFromFile(metadataFlag, metadata); err != nil {
			return nil, err
		}
	}

	return metadata, nil
}
