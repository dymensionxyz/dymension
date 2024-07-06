package cli

import (
	"encoding/json"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

var _ = strconv.Itoa(0)

// PermissionedAddresses .. TODO: refactor to be flag of []string
type PermissionedAddresses struct {
	Addresses []string `json:"addresses"`
}

func CmdCreateRollapp() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create-rollapp [rollapp-id] [init-sequencers-address] [bech32-prefix] [metadata.json]",
		Short:   "Create a new rollapp",
		Example: "dymd tx rollapp create-rollapp ROLLAPP_CHAIN_ID <seq_address> ethm metadata.json",
		Args:    cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argRollappId := args[0]

			argInitSequencerAddress := args[1]
			argBech32Prefix := args[2]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			genesisInfo := new(types.GenesisInfo)
			if err := json.Unmarshal([]byte(args[3]), genesisInfo); err != nil {
				return err
			}

			msg := types.NewMsgCreateRollapp(
				clientCtx.GetFromAddress().String(),
				argRollappId,
				argInitSequencerAddress,
				argBech32Prefix,
				genesisInfo,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetCreateRollapp())
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
