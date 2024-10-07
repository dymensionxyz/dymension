package cli

import (
	"fmt"
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
		Use:   "create-rollapp [rollapp-id] [alias] [vm-type]",
		Short: "Create a new rollapp",
		Example: `
		dymd tx rollapp create-rollapp myrollapp_12345-1 RollappAlias EVM 
		// optional flags:
		--init-sequencer '<seq_address1>,<seq_address2>'
		--genesis-checksum <genesis_checksum>
		--initial-supply 1000000
		--native-denom native_denom.json
		--genesis-accounts '<acc1>:1000000,<acc2>:1000000'
		--metadata metadata.json
		`,
		Args: cobra.MinimumNArgs(3),
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

func parseGenesisInfo(cmd *cobra.Command) (*types.GenesisInfo, error) {
	var (
		genesisInfo = &types.GenesisInfo{}
		err         error
		ok          bool
	)

	genesisInfo.GenesisChecksum, err = cmd.Flags().GetString(FlagGenesisChecksum)
	if err != nil {
		return nil, err
	}

	genesisInfo.Bech32Prefix, err = cmd.Flags().GetString(FlagBech32Prefix)
	if err != nil {
		return nil, err
	}

	nativeDenomFlag, err := cmd.Flags().GetString(FlagNativeDenom)
	if err != nil {
		return nil, err
	}

	if nativeDenomFlag != "" {
		if err = utils.ParseJsonFromFile(nativeDenomFlag, &genesisInfo.NativeDenom); err != nil {
			return nil, err
		}
	}

	initialSupplyFlag, err := cmd.Flags().GetString(FlagInitialSupply)
	if err != nil {
		return nil, err
	}

	if initialSupplyFlag != "" {
		genesisInfo.InitialSupply, ok = sdk.NewIntFromString(initialSupplyFlag)
		if !ok {
			return nil, fmt.Errorf("invalid initial supply: %s", initialSupplyFlag)
		}
	}

	// Parsing genesis accounts
	genesisAccountsFlag, err := cmd.Flags().GetString(FlagGenesisAccounts)
	if err != nil {
		return nil, err
	}

	if genesisAccountsFlag != "" {
		// split the string by comma
		genesisAccounts := strings.Split(genesisAccountsFlag, ",")
		for _, acc := range genesisAccounts {
			// split the account by colon
			accParts := strings.Split(acc, ":")
			if len(accParts) != 2 {
				return nil, fmt.Errorf("invalid genesis account: %s", acc)
			}

			accAddr, accAmt := accParts[0], accParts[1]
			amt, ok := sdk.NewIntFromString(accAmt)
			if !ok {
				return nil, fmt.Errorf("invalid genesis account amount: %s", accAmt)
			}

			genesisInfo.GenesisAccounts = append(genesisInfo.GenesisAccounts, types.GenesisAccount{
				Address: accAddr,
				Amount:  amt,
			})
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
