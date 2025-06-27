package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/utils"
	"github.com/dymensionxyz/dymension/v3/x/kas/types"
)

func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdSetupBridge())

	return cmd
}

func CmdSetupBridge() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup-bridge [foo]",
		Short: "foo",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argPubkey := args[0]
			argRollappId := args[1]
			bond := args[2]

			var metadata types.SequencerMetadata
			if len(args) == 4 {
				if err = utils.ParseJsonFromFile(args[3], &metadata); err != nil {
					return
				}
			}

			rewardAddr, _ := cmd.Flags().GetString(FlagRewardAddress)

			var whitelistedRelayers []string
			if relayers, _ := cmd.Flags().GetString(FlagWhitelistedRelayers); relayers != "" {
				whitelistedRelayers = strings.Split(relayers, ",")
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var pk cryptotypes.PubKey

			if err = clientCtx.Codec.UnmarshalInterfaceJSON([]byte(argPubkey), &pk); err != nil {
				return err
			}

			bondCoin, err := sdk.ParseCoinNormalized(bond)
			if err != nil {
				return err
			}

			msg, err := types.NewMsgCreateSequencer(
				clientCtx.GetFromAddress().String(),
				pk,
				argRollappId,
				&metadata,
				bondCoin,
				rewardAddr,
				whitelistedRelayers,
			)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String(FlagRewardAddress, "", "Sequencer reward address")
	cmd.Flags().String(FlagWhitelistedRelayers, "", "Sequencer whitelisted relayers")

	return cmd
}
