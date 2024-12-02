package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func CmdUpdateRollapp() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-rollapp [rollapp-id] [init-sequencer] [min-sequencer-bond] [genesis_checksum] [bech32-prefix] [native-denom] [metadata] ",
		Short: "Update a new rollapp",
		Example: `
		dymd tx rollapp update-rollapp ROLLAPP_CHAIN_ID
		--init-sequencer '<seq_address1>,<seq_address2>'
        --min-sequencer-bond 100
		--genesis-checksum <genesis_checksum>
		--initial-supply 1000000
		--native-denom native_denom.json
		--genesis-accounts '<acc1>:1000000,<acc2>:1000000'
		--metadata metadata.json`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argRollappId := args[0]

			initSequencer, err := cmd.Flags().GetString(FlagInitSequencer)
			if err != nil {
				return
			}

			minSeqBondS, err := cmd.Flags().GetString(FlagMinSequencerBond)
			if err != nil {
				return err
			}
			minSeqBond, ok := sdk.NewIntFromString(minSeqBondS)
			if !ok {
				return fmt.Errorf("invalid min sequencer bond: %s", minSeqBondS)
			}

			genesisInfo, err := parseGenesisInfo(cmd)
			if err != nil {
				return
			}

			metadata, err := parseMetadata(cmd)
			if err != nil {
				return
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return
			}

			msg := types.NewMsgUpdateRollappInformation(
				clientCtx.GetFromAddress().String(),
				argRollappId,
				initSequencer,
				commontypes.ADym(minSeqBond),
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
