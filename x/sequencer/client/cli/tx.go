package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/utils"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdCreateSequencer())
	cmd.AddCommand(CmdUpdateSequencer())
	cmd.AddCommand(CmdUnbond())
	cmd.AddCommand(CmdIncreaseBond())

	return cmd
}

func CmdCreateSequencer() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-sequencer [pubkey] [rollapp-id] [bond] [metadata]",
		Short: "Create a new sequencer for a rollapp",
		Args:  cobra.MinimumNArgs(3),
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
			)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdUpdateSequencer() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-sequencer [metadata]",
		Short: "Update a sequencer",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			metadata := new(types.SequencerMetadata)
			if err = utils.ParseJsonFromFile(args[0], metadata); err != nil {
				return
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg, err := types.NewMsgUpdateSequencerInformation(
				clientCtx.GetFromAddress().String(),
				metadata,
			)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdUnbond() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unbond",
		Short: "Unbond the sequencer",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgUnbond(
				clientCtx.GetFromAddress().String(),
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdIncreaseBond() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "increase-bond [amount]",
		Short: "Increase the bond of a sequencer",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			amount := args[0]
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			amountCoin, err := sdk.ParseCoinNormalized(amount)
			if err != nil {
				return err
			}

			msg := types.NewMsgIncreaseBond(
				clientCtx.GetFromAddress().String(),
				amountCoin,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
